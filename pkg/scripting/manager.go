// Package scripting provides Lua script execution and management for Stream Deck integration.
//
// This package enables programmable Stream Deck functionality through Lua scripts,
// providing modules for system interaction, HTTP requests, shell commands, and
// Stream Deck control. Scripts can define background workers, passive updates,
// and trigger actions.
//
// Key components:
// - ScriptManager: Coordinates multiple script runners and passive updates
// - ScriptRunner: Manages individual Lua script lifecycle
// - Modules: Preloaded Lua modules for various system interactions
// - Image handling: Caching and loading of button images
//
// Contributors can extend functionality by:
// - Adding new Lua modules in the modules/ subdirectory
// - Implementing custom script runners
// - Extending the image loading system
// - Adding new script lifecycle hooks
package scripting

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/merith-tk/nomad/pkg/streamdeck"
	lua "github.com/yuin/gopher-lua"
)

const (
	// DefaultPassiveFPS is the default rate at which passive functions are called.
	DefaultPassiveFPS = 10
)

// ScriptManager coordinates all script runners and the passive loop.
type ScriptManager struct {
	mu sync.RWMutex

	device     *streamdeck.Device
	configDir  string
	passiveFPS int

	// All loaded script runners, keyed by script path
	runners map[string]*ScriptRunner

	// Context for lifecycle management
	ctx    context.Context
	cancel context.CancelFunc

	// Passive loop
	passiveRunning bool
	visibleScripts map[string]int // script path -> key index (currently visible)
	refreshPending bool           // flag for coalesced refresh requests

	// Passive update batching
	lastPassiveUpdate time.Time
	passiveBatch      map[string]*KeyAppearance // batched updates

	// Boot animation
	bootScriptPath string

	// Callback when passive wants to update a key
	onKeyUpdate func(keyIndex int, appearance *KeyAppearance)

	// T1 / T2 toggle-key scripts – set by the app on every navigation
	t1Script string
	t1Key    int
	t2Script string
	t2Key    int
}

// NewScriptManager creates a new script manager.
func NewScriptManager(dev *streamdeck.Device, configDir string, passiveFPS int) *ScriptManager {
	if passiveFPS <= 0 {
		passiveFPS = DefaultPassiveFPS
	}
	return &ScriptManager{
		device:         dev,
		configDir:      configDir,
		passiveFPS:     passiveFPS,
		runners:        make(map[string]*ScriptRunner),
		visibleScripts: make(map[string]int),
		passiveBatch:   make(map[string]*KeyAppearance),
	}
}

// SetKeyUpdateCallback sets the callback for passive key updates.
func (m *ScriptManager) SetKeyUpdateCallback(cb func(keyIndex int, appearance *KeyAppearance)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onKeyUpdate = cb
}

// Boot scans the config directory and loads all scripts.
// Runs boot animation if _boot.lua exists, then loads all scripts.
func (m *ScriptManager) Boot(ctx context.Context) error {
	m.mu.Lock()
	m.ctx, m.cancel = context.WithCancel(ctx)
	m.mu.Unlock()

	// Check for boot animation script - runs synchronously
	bootPath := filepath.Join(m.configDir, "_boot.lua")
	if _, err := os.Stat(bootPath); err == nil {
		m.bootScriptPath = bootPath
		// Run boot animation synchronously (blocks until complete)
		m.runBootAnimation()
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

	// Call the boot function from the module table
	if runner.module == nil {
		return
	}
	fn := runner.module.RawGetString("boot")
	if fn.Type() != lua.LTFunction {
		return
	}

	runner.L.Push(fn)
	if err := runner.L.PCall(0, 0, nil); err != nil {
		fmt.Printf("[!] Boot animation error: %v\n", err)
	}
}

// StartPassiveLoop starts the passive update loop at the configured FPS.
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

// passiveLoop runs passive functions at the configured FPS.
func (m *ScriptManager) passiveLoop() {
	fps := m.passiveFPS
	if fps <= 0 {
		fps = DefaultPassiveFPS
	}
	interval := time.Second / time.Duration(fps)

	ticker := time.NewTicker(interval)
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
			m.runTogglePassive() // always runs, even when no content scripts are visible

			// Process batched updates (limit to prevent blocking)
			m.processBatchedUpdates(5) // Process up to 5 updates per tick
		}
	}
}

// runPassiveUpdate calls passive() on all visible scripts concurrently.
func (m *ScriptManager) runPassiveUpdate() {
	m.mu.RLock()
	visible := make(map[string]int)
	for k, v := range m.visibleScripts {
		visible[k] = v
	}
	m.mu.RUnlock()

	if len(visible) == 0 {
		return
	}

	var wg sync.WaitGroup
	for scriptPath, keyIndex := range visible {
		wg.Add(1)
		go func(scriptPath string, keyIndex int) {
			defer wg.Done()

			m.mu.RLock()
			runner := m.runners[scriptPath]
			m.mu.RUnlock()

			if runner == nil || !runner.HasPassive() {
				return
			}

			appearance, err := runner.RunPassive(keyIndex)
			if err != nil {
				return
			}

			if appearance != nil {
				// Batch the update instead of calling callback immediately
				m.batchUpdate(scriptPath, appearance)
			}
		}(scriptPath, keyIndex)
	}
	wg.Wait()
}

// runPassiveUpdate calls passive() on all visible content-key scripts concurrently. adds an update to the batch queue.
func (m *ScriptManager) batchUpdate(scriptPath string, appearance *KeyAppearance) {
	m.mu.Lock()
	m.passiveBatch[scriptPath] = appearance
	m.mu.Unlock()
}

// processBatchedUpdates processes queued passive updates.
func (m *ScriptManager) processBatchedUpdates(maxUpdates int) {
	m.mu.Lock()
	batch := make(map[string]*KeyAppearance)
	for k, v := range m.passiveBatch {
		batch[k] = v
	}
	// Clear the batch
	m.passiveBatch = make(map[string]*KeyAppearance)
	callback := m.onKeyUpdate
	m.mu.Unlock()

	if callback == nil {
		return
	}

	// Process updates
	processed := 0
	for scriptPath, appearance := range batch {
		if processed >= maxUpdates {
			break
		}

		// Find the key index for this script
		m.mu.RLock()
		keyIndex, visible := m.visibleScripts[scriptPath]
		m.mu.RUnlock()

		if visible {
			callback(keyIndex, appearance)
			processed++
		}
	}

	// Re-queue remaining updates if we hit the limit
	if len(batch) > processed {
		m.mu.Lock()
		for scriptPath, appearance := range batch {
			if _, alreadyProcessed := m.passiveBatch[scriptPath]; !alreadyProcessed {
				m.passiveBatch[scriptPath] = appearance
			}
		}
		m.mu.Unlock()
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

// IsUsableScript returns true if the script has been loaded and defines at least
// one of background / passive / trigger. Used by the Navigator to filter the
// button list so that helper-only scripts are not shown as buttons.
func (m *ScriptManager) IsUsableScript(scriptPath string) bool {
	m.mu.RLock()
	runner := m.runners[scriptPath]
	m.mu.RUnlock()
	if runner == nil {
		return false
	}
	return runner.HasBackground() || runner.HasPassive() || runner.HasTrigger()
}

// SetToggleScripts registers the .directory.lua script (and physical key indices)
// that should drive the T1 and T2 reserved keys via t1_passive/t1_trigger etc.
// Pass an empty string for either path to fall back to default toggle behaviour.
func (m *ScriptManager) SetToggleScripts(t1Script string, t1Key int, t2Script string, t2Key int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.t1Script = t1Script
	m.t1Key = t1Key
	m.t2Script = t2Script
	m.t2Key = t2Key
}

// HasT1Script returns true when a script is driving the T1 key.
func (m *ScriptManager) HasT1Script() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.t1Script != ""
}

// HasT2Script returns true when a script is driving the T2 key.
func (m *ScriptManager) HasT2Script() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.t2Script != ""
}

// TriggerT1 calls t1_trigger on the registered T1 script, if any.
func (m *ScriptManager) TriggerT1() error {
	m.mu.RLock()
	runner := m.runners[m.t1Script]
	m.mu.RUnlock()
	if runner == nil {
		return nil
	}
	return runner.RunT1Trigger()
}

// TriggerT2 calls t2_trigger on the registered T2 script, if any.
func (m *ScriptManager) TriggerT2() error {
	m.mu.RLock()
	runner := m.runners[m.t2Script]
	m.mu.RUnlock()
	if runner == nil {
		return nil
	}
	return runner.RunT2Trigger()
}

// runTogglePassive runs t1_passive / t2_passive for the currently registered toggle scripts.
func (m *ScriptManager) runTogglePassive() {
	type toggleEntry struct {
		script string
		key    int
		isT1   bool
	}

	m.mu.RLock()
	entries := []toggleEntry{
		{m.t1Script, m.t1Key, true},
		{m.t2Script, m.t2Key, false},
	}
	cb := m.onKeyUpdate
	m.mu.RUnlock()

	for _, e := range entries {
		if e.script == "" || cb == nil {
			continue
		}
		m.mu.RLock()
		runner := m.runners[e.script]
		m.mu.RUnlock()
		if runner == nil {
			continue
		}
		var ap *KeyAppearance
		var err error
		if e.isT1 {
			ap, err = runner.RunT1Passive(e.key)
		} else {
			ap, err = runner.RunT2Passive(e.key)
		}
		if err != nil || ap == nil {
			continue
		}
		cb(e.key, ap)
	}
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

// RefreshScript immediately runs passive() for one script and pushes the result
// through the key-update callback. Use this after a trigger to update just the
// pressed button instead of redrawing the entire display.
func (m *ScriptManager) RefreshScript(scriptPath string) {
	m.mu.RLock()
	runner := m.runners[scriptPath]
	keyIndex, visible := m.visibleScripts[scriptPath]
	callback := m.onKeyUpdate
	m.mu.RUnlock()

	if runner == nil || !visible || callback == nil || !runner.HasPassive() {
		return
	}

	appearance, err := runner.RunPassive(keyIndex)
	if err != nil || appearance == nil {
		return
	}

	callback(keyIndex, appearance)
}

// requestRefresh is called when a script wants a display refresh.
// Sets a flag that will be picked up by the next passive loop tick.
func (m *ScriptManager) requestRefresh() {
	m.mu.Lock()
	m.refreshPending = true
	m.mu.Unlock()
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
