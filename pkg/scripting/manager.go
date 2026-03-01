package scripting

import (
	"context"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/merith-tk/nomad/pkg/streamdeck"
	lua "github.com/yuin/gopher-lua"
)

const (
	// PassiveFPS is the rate at which passive functions are called.
	PassiveFPS = 15
	// PassiveInterval is the duration between passive calls.
	PassiveInterval = time.Second / PassiveFPS
)

// ScriptManager coordinates all script runners and the passive loop.
type ScriptManager struct {
	mu sync.RWMutex

	device    *streamdeck.Device
	configDir string

	// All loaded script runners, keyed by script path
	runners map[string]*ScriptRunner

	// Context for lifecycle management
	ctx    context.Context
	cancel context.CancelFunc

	// Passive loop
	passiveRunning bool
	visibleScripts map[string]int // script path -> key index (currently visible)

	// Boot animation
	bootScriptPath string

	// Callback when passive wants to update a key
	onKeyUpdate func(keyIndex int, appearance *KeyAppearance)
}

// NewScriptManager creates a new script manager.
func NewScriptManager(dev *streamdeck.Device, configDir string) *ScriptManager {
	return &ScriptManager{
		device:         dev,
		configDir:      configDir,
		runners:        make(map[string]*ScriptRunner),
		visibleScripts: make(map[string]int),
	}
}

// SetKeyUpdateCallback sets the callback for passive key updates.
func (m *ScriptManager) SetKeyUpdateCallback(cb func(keyIndex int, appearance *KeyAppearance)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onKeyUpdate = cb
}

// Boot scans the config directory and loads all scripts.
// Shows a loading animation during boot.
func (m *ScriptManager) Boot(ctx context.Context) error {
	m.mu.Lock()
	m.ctx, m.cancel = context.WithCancel(ctx)
	m.mu.Unlock()

	// Show loading indicator on center key (key 7 for 5x3 layout)
	centerKey := 7
	if m.device != nil && m.device.Model.Keys > 0 {
		centerKey = (m.device.Model.Rows/2)*m.device.Model.Cols + (m.device.Model.Cols / 2)
	}

	// Clear and show loading
	if m.device != nil {
		m.device.Clear()
		m.device.SetKeyColor(centerKey, color.RGBA{255, 165, 0, 255}) // Orange "loading"
	}

	// Check for boot animation script
	bootPath := filepath.Join(m.configDir, "_boot.lua")
	if _, err := os.Stat(bootPath); err == nil {
		m.bootScriptPath = bootPath
		// Run boot animation in background while loading
		go m.runBootAnimation()
	}

	// Scan for all .lua files recursively
	var scriptPaths []string
	err := filepath.Walk(m.configDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".lua" && filepath.Base(path) != "_boot.lua" {
			scriptPaths = append(scriptPaths, path)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan config directory: %w", err)
	}

	fmt.Printf("[*] Found %d scripts to load...\n", len(scriptPaths))

	// Load each script
	loaded := 0
	for _, scriptPath := range scriptPaths {
		runner, err := NewScriptRunner(scriptPath, m.device, m.configDir)
		if err != nil {
			fmt.Printf("[!] Failed to load %s: %v\n", filepath.Base(scriptPath), err)
			continue
		}

		// Set refresh callback
		runner.SetRefreshCallback(m.requestRefresh)

		m.mu.Lock()
		m.runners[scriptPath] = runner
		m.mu.Unlock()

		loaded++

		// Start background worker if defined
		if runner.HasBackground() {
			fmt.Printf("[*] Starting background worker: %s\n", runner.ScriptName)
			runner.StartBackground(m.ctx)
		}
	}

	fmt.Printf("[*] Loaded %d/%d scripts\n", loaded, len(scriptPaths))

	// Clear loading indicator
	if m.device != nil {
		m.device.Clear()
	}

	return nil
}

// runBootAnimation runs the optional _boot.lua animation script.
func (m *ScriptManager) runBootAnimation() {
	if m.bootScriptPath == "" {
		return
	}

	runner, err := NewScriptRunner(m.bootScriptPath, m.device, m.configDir)
	if err != nil {
		fmt.Printf("[!] Boot animation failed: %v\n", err)
		return
	}
	defer runner.Close()

	// Call the boot function if defined
	fn := runner.L.GetGlobal("boot")
	if fn.Type() != lua.LTFunction {
		return
	}

	runner.L.Push(fn)
	if err := runner.L.PCall(0, 0, nil); err != nil {
		fmt.Printf("[!] Boot animation error: %v\n", err)
	}
}

// StartPassiveLoop starts the 15fps passive update loop.
func (m *ScriptManager) StartPassiveLoop() {
	m.mu.Lock()
	if m.passiveRunning {
		m.mu.Unlock()
		return
	}
	m.passiveRunning = true
	m.mu.Unlock()

	go m.passiveLoop()
}

// passiveLoop runs passive functions at 15fps.
func (m *ScriptManager) passiveLoop() {
	ticker := time.NewTicker(PassiveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			m.mu.Lock()
			m.passiveRunning = false
			m.mu.Unlock()
			return
		case <-ticker.C:
			m.runPassiveUpdate()
		}
	}
}

// runPassiveUpdate calls passive() on all visible scripts.
func (m *ScriptManager) runPassiveUpdate() {
	m.mu.RLock()
	visible := make(map[string]int)
	for k, v := range m.visibleScripts {
		visible[k] = v
	}
	callback := m.onKeyUpdate
	m.mu.RUnlock()

	if callback == nil {
		return
	}

	for scriptPath, keyIndex := range visible {
		m.mu.RLock()
		runner := m.runners[scriptPath]
		m.mu.RUnlock()

		if runner == nil || !runner.HasPassive() {
			continue
		}

		appearance, err := runner.RunPassive(keyIndex)
		if err != nil {
			// Silent fail for passive - don't spam logs at 15fps
			continue
		}

		if appearance != nil {
			callback(keyIndex, appearance)
		}
	}
}

// SetVisibleScripts updates which scripts are currently visible on the display.
// Map is scriptPath -> keyIndex
func (m *ScriptManager) SetVisibleScripts(scripts map[string]int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.visibleScripts = make(map[string]int)
	for k, v := range scripts {
		m.visibleScripts[k] = v
	}
}

// GetRunner returns the runner for a script path.
func (m *ScriptManager) GetRunner(scriptPath string) *ScriptRunner {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.runners[scriptPath]
}

// TriggerScript executes the trigger function for a script.
func (m *ScriptManager) TriggerScript(scriptPath string) error {
	m.mu.RLock()
	runner := m.runners[scriptPath]
	m.mu.RUnlock()

	if runner == nil {
		return fmt.Errorf("script not loaded: %s", scriptPath)
	}

	return runner.RunTrigger()
}

// requestRefresh is called when a script wants a display refresh.
func (m *ScriptManager) requestRefresh() {
	// Run an immediate passive update
	go m.runPassiveUpdate()
}

// Shutdown stops all runners and cleans up.
func (m *ScriptManager) Shutdown() {
	m.mu.Lock()
	if m.cancel != nil {
		m.cancel()
	}

	// Close all runners
	for path, runner := range m.runners {
		runner.Close()
		delete(m.runners, path)
	}
	m.mu.Unlock()

	fmt.Println("[*] Script manager shutdown complete")
}
