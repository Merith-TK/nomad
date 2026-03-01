package main

import (
	"context"
	"fmt"
	"image/color"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/merith-tk/nomad/pkg/streamdeck"
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

	// Demo: Use the first device
	info := devices[0]
	if info.Model.PixelSize == 0 {
		fmt.Println("First device has no display (e.g., Pedal). Skipping image test.")
		return
	}

	fmt.Printf("Opening %s for demo...\n", info.Model.Name)

	dev, err := streamdeck.Open(info.Path)
	if err != nil {
		log.Fatal("Failed to open device:", err)
	}
	defer dev.Close()

	// Set brightness
	if err := dev.SetBrightness(75); err != nil {
		log.Printf("SetBrightness failed: %v", err)
	}

	// Demo: Set each key to a different color
	fmt.Println("\n[*] Setting rainbow colors on all keys...")
	colors := []color.RGBA{
		{255, 0, 0, 255},     // Red
		{255, 127, 0, 255},   // Orange
		{255, 255, 0, 255},   // Yellow
		{0, 255, 0, 255},     // Green
		{0, 255, 255, 255},   // Cyan
		{0, 127, 255, 255},   // Light blue
		{0, 0, 255, 255},     // Blue
		{127, 0, 255, 255},   // Purple
		{255, 0, 255, 255},   // Magenta
		{255, 0, 127, 255},   // Pink
		{128, 128, 128, 255}, // Gray
		{255, 255, 255, 255}, // White
		{64, 64, 64, 255},    // Dark gray
		{128, 64, 0, 255},    // Brown
		{0, 128, 64, 255},    // Teal
	}

	for i := 0; i < dev.Keys() && i < len(colors); i++ {
		col, row := dev.KeyToCoord(i)
		fmt.Printf("  Key %d (col %d, row %d): setting color...\n", i, col, row)
		if err := dev.SetKeyColor(i, colors[i]); err != nil {
			log.Printf("SetKeyColor %d failed: %v", i, err)
		}
	}

	fmt.Println("\n[OK] All keys colored!")
	fmt.Println("\n[*] Press any key on the Stream Deck (Ctrl+C to exit)...\n")

	// Listen for key presses
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
		col, row := dev.KeyToCoord(event.Key)
		if event.Pressed {
			fmt.Printf("[D] Key %d pressed  (col %d, row %d)\n", event.Key, col, row)
		} else {
			fmt.Printf("[U] Key %d released (col %d, row %d)\n", event.Key, col, row)
		}
	}

	fmt.Println("Done!")
}
