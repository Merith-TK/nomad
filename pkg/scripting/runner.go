package scripting

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/merith-tk/nomad/pkg/streamdeck"
	lua "github.com/yuin/gopher-lua"
)

// RestartPolicy defines how background workers handle errors.
type RestartPolicy int

const (
	RestartAlways RestartPolicy = iota // Always restart on error (default)
	RestartNever                       // Never restart, fail permanently
	RestartOnce                        // Restart once, then stop
)

// KeyAppearance defines how a key should look (returned by passive).
type KeyAppearance struct {
	Color     [3]int // RGB color (0-255)
	Text      string // Text to display
	TextColor [3]int // Text color RGB
	Image     string // Path to image file (future)
}

// ScriptRunner manages a single Lua script's lifecycle.
type ScriptRunner struct {
	mu sync.RWMutex

	// Script info
	ScriptPath string
	ScriptName string // Filename without .lua

	// Lua state (persistent for shared state)
	L     *lua.LState
	state *lua.LTable // Shared state table

	// Function availability
	hasBackground bool
	hasPassive    bool
	hasTrigger    bool

	// Background worker
	bgCtx         context.Context
	bgCancel      context.CancelFunc
	bgRunning     bool
	bgRestarts    int
	restartPolicy RestartPolicy

	// Device access
	device    *streamdeck.Device
	configDir string

	// Refresh callback (called when script wants display update)
	onRefresh func()
}

// NewScriptRunner creates a runner for a Lua script.
func NewScriptRunner(scriptPath string, dev *streamdeck.Device, configDir string) (*ScriptRunner, error) {
	r := &ScriptRunner{
		ScriptPath:    scriptPath,
		ScriptName:    filepath.Base(scriptPath[:len(scriptPath)-4]), // Remove .lua
		device:        dev,
		configDir:     configDir,
		restartPolicy: RestartAlways,
	}

	// Create Lua state
	r.L = lua.NewState()

	// Create shared state table
	r.state = r.L.NewTable()
	r.L.SetGlobal("state", r.state)

	// Register modules
	r.registerModules()

	// Load the script (defines functions)
	if err := r.L.DoFile(scriptPath); err != nil {
		r.L.Close()
		return nil, fmt.Errorf("failed to load script %s: %w", scriptPath, err)
	}

	// Check which functions are defined
	r.hasBackground = r.L.GetGlobal("background").Type() == lua.LTFunction
	r.hasPassive = r.L.GetGlobal("passive").Type() == lua.LTFunction
	r.hasTrigger = r.L.GetGlobal("trigger").Type() == lua.LTFunction

	// Check for restart policy setting
	policy := r.L.GetGlobal("RESTART_POLICY")
	if policy.Type() == lua.LTString {
		switch policy.String() {
		case "never":
			r.restartPolicy = RestartNever
		case "once":
			r.restartPolicy = RestartOnce
		case "always":
			r.restartPolicy = RestartAlways
		}
	}

	return r, nil
}

// registerModules adds all available modules to the Lua state.
func (r *ScriptRunner) registerModules() {
	// Shell module
	r.L.PreloadModule("shell", r.loaderShell)

	// HTTP module
	r.L.PreloadModule("http", r.loaderHTTP)

	// System module
	r.L.PreloadModule("system", r.loaderSystem)

	// StreamDeck module
	r.L.PreloadModule("streamdeck", r.loaderStreamDeck)

	// Set globals
	r.L.SetGlobal("SCRIPT_PATH", lua.LString(r.ScriptPath))
	r.L.SetGlobal("SCRIPT_NAME", lua.LString(r.ScriptName))
	r.L.SetGlobal("CONFIG_DIR", lua.LString(r.configDir))
}

// SetRefreshCallback sets the function called when script requests refresh.
func (r *ScriptRunner) SetRefreshCallback(cb func()) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onRefresh = cb
}

// requestRefresh triggers a display refresh from within a script.
func (r *ScriptRunner) requestRefresh() {
	r.mu.RLock()
	cb := r.onRefresh
	r.mu.RUnlock()

	if cb != nil {
		cb()
	}
}

// HasBackground returns true if script defines background().
func (r *ScriptRunner) HasBackground() bool {
	return r.hasBackground
}

// HasPassive returns true if script defines passive().
func (r *ScriptRunner) HasPassive() bool {
	return r.hasPassive
}

// HasTrigger returns true if script defines trigger().
func (r *ScriptRunner) HasTrigger() bool {
	return r.hasTrigger
}

// StartBackground starts the background worker goroutine.
func (r *ScriptRunner) StartBackground(parentCtx context.Context) {
	if !r.hasBackground {
		return
	}

	r.mu.Lock()
	if r.bgRunning {
		r.mu.Unlock()
		return
	}

	r.bgCtx, r.bgCancel = context.WithCancel(parentCtx)
	r.bgRunning = true
	r.mu.Unlock()

	go r.backgroundLoop()
}

// backgroundLoop runs the background function with restart logic.
func (r *ScriptRunner) backgroundLoop() {
	defer func() {
		r.mu.Lock()
		r.bgRunning = false
		r.mu.Unlock()
	}()

	for {
		select {
		case <-r.bgCtx.Done():
			return
		default:
		}

		// Call background(state) - should be quick, Go handles the loop
		err := r.callBackground()

		// Pause between calls - 500ms gives passive plenty of time to run
		select {
		case <-r.bgCtx.Done():
			return
		case <-time.After(500 * time.Millisecond):
		}

		if err != nil {
			fmt.Printf("[!] Background error in %s: %v\n", r.ScriptName, err)

			r.mu.Lock()
			r.bgRestarts++
			policy := r.restartPolicy
			restarts := r.bgRestarts
			r.mu.Unlock()

			// Check restart policy
			switch policy {
			case RestartNever:
				fmt.Printf("[!] %s: restart policy is 'never', stopping background\n", r.ScriptName)
				return
			case RestartOnce:
				if restarts > 1 {
					fmt.Printf("[!] %s: restart policy is 'once', max restarts reached\n", r.ScriptName)
					return
				}
				fmt.Printf("[*] %s: restarting background (attempt %d)\n", r.ScriptName, restarts)
			case RestartAlways:
				fmt.Printf("[*] %s: restarting background (attempt %d)\n", r.ScriptName, restarts)
			}

			// Brief delay before restart
			select {
			case <-r.bgCtx.Done():
				return
			case <-time.After(1 * time.Second):
			}
		}
	}
}

// callBackground executes the background function.
func (r *ScriptRunner) callBackground() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	fn := r.L.GetGlobal("background")
	if fn.Type() != lua.LTFunction {
		return nil
	}

	r.L.Push(fn)
	r.L.Push(r.state)

	if err := r.L.PCall(1, 0, nil); err != nil {
		return err
	}

	return nil
}

// StopBackground stops the background worker.
func (r *ScriptRunner) StopBackground() {
	r.mu.Lock()
	if r.bgCancel != nil {
		r.bgCancel()
	}
	r.mu.Unlock()
}

// RunPassive calls passive(key, state) and returns appearance.
// Uses TryLock to avoid blocking if background is running.
func (r *ScriptRunner) RunPassive(keyIndex int) (*KeyAppearance, error) {
	// Try to acquire lock - if background is holding it, skip this update
	if !r.mu.TryLock() {
		return nil, nil // Skip, try again next tick
	}
	defer r.mu.Unlock()

	if !r.hasPassive {
		return nil, nil
	}

	fn := r.L.GetGlobal("passive")
	if fn.Type() != lua.LTFunction {
		return nil, nil
	}

	r.L.Push(fn)
	r.L.Push(lua.LNumber(keyIndex))
	r.L.Push(r.state)

	if err := r.L.PCall(2, 1, nil); err != nil {
		return nil, err
	}

	// Get return value
	ret := r.L.Get(-1)
	r.L.Pop(1)

	if ret.Type() == lua.LTNil {
		return nil, nil
	}

	if ret.Type() != lua.LTTable {
		return nil, nil
	}

	tbl := ret.(*lua.LTable)
	appearance := &KeyAppearance{}

	// Parse color: {r, g, b} or table with .color field
	if colorVal := r.L.GetField(tbl, "color"); colorVal.Type() == lua.LTTable {
		colorTbl := colorVal.(*lua.LTable)
		appearance.Color[0] = int(lua.LVAsNumber(r.L.RawGetInt(colorTbl, 1)))
		appearance.Color[1] = int(lua.LVAsNumber(r.L.RawGetInt(colorTbl, 2)))
		appearance.Color[2] = int(lua.LVAsNumber(r.L.RawGetInt(colorTbl, 3)))
	}

	// Parse text
	if textVal := r.L.GetField(tbl, "text"); textVal.Type() == lua.LTString {
		appearance.Text = textVal.String()
	}

	// Parse text_color
	if tcVal := r.L.GetField(tbl, "text_color"); tcVal.Type() == lua.LTTable {
		tcTbl := tcVal.(*lua.LTable)
		appearance.TextColor[0] = int(lua.LVAsNumber(r.L.RawGetInt(tcTbl, 1)))
		appearance.TextColor[1] = int(lua.LVAsNumber(r.L.RawGetInt(tcTbl, 2)))
		appearance.TextColor[2] = int(lua.LVAsNumber(r.L.RawGetInt(tcTbl, 3)))
	} else {
		// Default text color: white
		appearance.TextColor = [3]int{255, 255, 255}
	}

	// Parse image path - resolve relative to script directory
	if imgVal := r.L.GetField(tbl, "image"); imgVal.Type() == lua.LTString {
		imgPath := imgVal.String()
		// If it's a URL, keep as-is
		if strings.HasPrefix(imgPath, "http://") || strings.HasPrefix(imgPath, "https://") {
			appearance.Image = imgPath
		} else if !filepath.IsAbs(imgPath) {
			// Relative path - resolve relative to script's directory
			scriptDir := filepath.Dir(r.ScriptPath)
			appearance.Image = filepath.Join(scriptDir, imgPath)
		} else {
			appearance.Image = imgPath
		}
	}

	return appearance, nil
}

// RunTrigger calls trigger(state).
func (r *ScriptRunner) RunTrigger() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.hasTrigger {
		return nil
	}

	fn := r.L.GetGlobal("trigger")
	if fn.Type() != lua.LTFunction {
		return nil
	}

	r.L.Push(fn)
	r.L.Push(r.state)

	if err := r.L.PCall(1, 0, nil); err != nil {
		return err
	}

	return nil
}

// Close shuts down the runner and releases resources.
func (r *ScriptRunner) Close() {
	r.StopBackground()

	r.mu.Lock()
	if r.L != nil {
		r.L.Close()
		r.L = nil
	}
	r.mu.Unlock()
}
