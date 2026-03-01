package main

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/merith-tk/nomad/pkg/scripting"
	"github.com/merith-tk/nomad/pkg/streamdeck"
)

// Config paths
const (
	// Development config path (relative to working directory)
	devConfigPath = ".nomad/interface/streamdeck/config"
)

func main() {
	// Initialize the streamdeck library
	if err := streamdeck.Init(); err != nil {
		log.Fatal("Failed to init streamdeck:", err)
	}
	defer streamdeck.Exit()

	// Probe for all Stream Deck devices
	fmt.Println("\n[*] Scanning for Stream Deck devices...\n")

	devices, err := streamdeck.Enumerate()
	if err != nil {
		log.Fatal("Failed to enumerate devices:", err)
	}

	if len(devices) == 0 {
		fmt.Println("No Stream Deck devices found.")
		return
	}

	fmt.Printf("Found %d Stream Deck device(s):\n\n", len(devices))

	for i, info := range devices {
		fmt.Printf("Device #%d:\n", i+1)
		streamdeck.PrintDeviceInfo(info)
		fmt.Println()
	}

	// Use the first device
	info := devices[0]
	if info.Model.PixelSize == 0 {
		fmt.Println("First device has no display (e.g., Pedal). Skipping.")
		return
	}

	fmt.Printf("Opening %s...\n", info.Model.Name)

	dev, err := streamdeck.Open(info.Path)
	if err != nil {
		log.Fatal("Failed to open device:", err)
	}
	defer dev.Close()

	// Set brightness
	if err := dev.SetBrightness(75); err != nil {
		log.Printf("SetBrightness failed: %v", err)
	}

	// Determine config path
	configPath := devConfigPath

	// Ensure config directory exists
	if err := os.MkdirAll(configPath, 0755); err != nil {
		log.Fatal("Failed to create config directory:", err)
	}

	absConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		log.Fatal("Failed to get absolute config path:", err)
	}

	fmt.Printf("\n[*] Config directory: %s\n", absConfigPath)

	// Create script manager and boot (loads scripts, starts background workers)
	fmt.Println("[*] Booting script manager...")
	scriptMgr := scripting.NewScriptManager(dev, absConfigPath)

	// Create a context for the entire application
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Boot scripts (shows loading indicator, loads all scripts)
	if err := scriptMgr.Boot(ctx); err != nil {
		log.Printf("Warning: Script boot error: %v", err)
	}
	defer scriptMgr.Shutdown()

	// Create navigator
	nav := streamdeck.NewNavigator(dev, absConfigPath)

	// Set up passive key updates from scripts
	scriptMgr.SetKeyUpdateCallback(func(keyIndex int, appearance *scripting.KeyAppearance) {
		if appearance == nil {
			return
		}
		// Apply appearance to key
		c := color.RGBA{
			R: uint8(appearance.Color[0]),
			G: uint8(appearance.Color[1]),
			B: uint8(appearance.Color[2]),
			A: 255,
		}
		if appearance.Text != "" {
			// Create text image with appearance colors
			img := nav.CreateTextImageWithColors(
				appearance.Text,
				c,
				color.RGBA{
					R: uint8(appearance.TextColor[0]),
					G: uint8(appearance.TextColor[1]),
					B: uint8(appearance.TextColor[2]),
					A: 255,
				},
			)
			dev.SetImage(keyIndex, img)
		} else {
			dev.SetKeyColor(keyIndex, c)
		}
	})

	// Start the passive update loop (15fps)
	scriptMgr.StartPassiveLoop()

	// Helper to update visible scripts
	updateVisibleScripts := func() {
		scriptMgr.SetVisibleScripts(nav.GetVisibleScripts())
	}

	// Render initial page
	fmt.Println("[*] Loading page...")
	if err := nav.RenderPage(); err != nil {
		log.Printf("Warning: RenderPage failed: %v", err)
	}

	// Show current path
	page, _ := nav.LoadPage()
	if page != nil {
		fmt.Printf("[*] Current: %s (%d items, page %d/%d)\n",
			page.Path, len(page.Items), page.PageIndex+1, page.TotalPages)
	}

	fmt.Println("\n[*] Navigation ready (Ctrl+C to exit)...")
	fmt.Println("    - Column 0: Reserved (Back, Toggle1, Toggle2)")
	fmt.Println("    - Columns 1-4: Folder/action buttons")
	fmt.Println("    - Press '<-' to go back\n")

	// Update visible scripts for initial page
	updateVisibleScripts()

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n\nExiting...")
		cancel()
	}()

	// Listen for key events
	events := make(chan streamdeck.KeyEvent, 10)
	dev.ListenKeys(ctx, events)

	for event := range events {
		// Only handle key presses, not releases
		if !event.Pressed {
			continue
		}

		col, row := dev.KeyToCoord(event.Key)
		fmt.Printf("[D] Key %d pressed (col %d, row %d)\n", event.Key, col, row)

		// Handle the key press
		item, navigated, err := nav.HandleKeyPress(event.Key)
		if err != nil {
			log.Printf("Error handling key: %v", err)
			continue
		}

		// Check for toggle state changes
		if event.Key == streamdeck.KeyToggle1 {
			fmt.Printf("[*] Toggle1: %v\n", nav.GetToggleState(streamdeck.KeyToggle1))
			continue
		}
		if event.Key == streamdeck.KeyToggle2 {
			fmt.Printf("[*] Toggle2: %v\n", nav.GetToggleState(streamdeck.KeyToggle2))
			continue
		}

		if navigated {
			// Page changed, re-render
			if err := nav.RenderPage(); err != nil {
				log.Printf("RenderPage failed: %v", err)
			}

			// Update visible scripts for passive updates
			updateVisibleScripts()

			page, _ := nav.LoadPage()
			if page != nil {
				relPath, _ := filepath.Rel(absConfigPath, page.Path)
				if relPath == "." {
					relPath = "/"
				} else {
					relPath = "/" + relPath
				}
				fmt.Printf("[*] Navigated to: %s (%d items)\n", relPath, len(page.Items))
			}
		} else if item != nil {
			// Action/script triggered
			fmt.Printf("[*] Action triggered: %s\n", item.Name)
			if item.Script != "" {
				fmt.Printf("    Script: %s\n", item.Script)
				if err := scriptMgr.TriggerScript(item.Script); err != nil {
					log.Printf("Script error: %v", err)
				}
			}
		}
	}

	fmt.Println("Done!")
}
