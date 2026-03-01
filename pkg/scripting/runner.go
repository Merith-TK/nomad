package scripting

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/merith-tk/nomad/pkg/scripting/modules"
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

	// Background coroutine support
	bgThread       *lua.LState // Coroutine for background function
	bgThreadCancel context.CancelFunc
	bgSleepUntil   time.Time      // When to resume from sleep
	bgFunc         *lua.LFunction // Cached background function

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
	// Create module instances
	shellMod := modules.NewShellModule()
	httpMod := modules.NewHTTPModule()
	systemMod := modules.NewSystemModule()
	sdMod := modules.NewStreamDeckModule(r.device)

	// Register modules
	r.L.PreloadModule("shell", shellMod.Loader)
	r.L.PreloadModule("http", httpMod.Loader)
	r.L.PreloadModule("system", systemMod.Loader)
	r.L.PreloadModule("streamdeck", sdMod.Loader)

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

// backgroundLoop runs the background function as a coroutine with restart logic.
func (r *ScriptRunner) backgroundLoop() {
	defer func() {
		r.mu.Lock()
		r.bgRunning = false
		if r.bgThreadCancel != nil {
			r.bgThreadCancel()
		}
		r.bgThread = nil
		r.bgFunc = nil
		r.mu.Unlock()
	}()

	for {
		select {
		case <-r.bgCtx.Done():
			return
		default:
		}

		// Run or resume background coroutine
		finished, sleepMs, err := r.runBackgroundCoroutine()

		if err != nil {
			fmt.Printf("[!] Background error in %s: %v\n", r.ScriptName, err)

			r.mu.Lock()
			r.bgRestarts++
			if r.bgThreadCancel != nil {
				r.bgThreadCancel()
			}
			r.bgThread = nil // Reset coroutine on error
			r.bgFunc = nil
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
			continue
		}

		if finished {
			// Coroutine finished normally, restart it
			r.mu.Lock()
			if r.bgThreadCancel != nil {
				r.bgThreadCancel()
			}
			r.bgThread = nil
			r.bgFunc = nil
			r.mu.Unlock()

			// Brief pause before restarting
			select {
			case <-r.bgCtx.Done():
				return
			case <-time.After(100 * time.Millisecond):
			}
			continue
		}

		// Coroutine yielded (sleep) - wait WITHOUT holding mutex
		if sleepMs > 0 {
			select {
			case <-r.bgCtx.Done():
				return
			case <-time.After(time.Duration(sleepMs) * time.Millisecond):
			}
		} else {
			// No sleep specified, brief yield to allow other operations
			select {
			case <-r.bgCtx.Done():
				return
			case <-time.After(10 * time.Millisecond):
			}
		}
	}
}

// runBackgroundCoroutine runs or resumes the background coroutine.
// Returns: (finished bool, sleepMs int, err error)
func (r *ScriptRunner) runBackgroundCoroutine() (bool, int, error) {
	r.mu.Lock()

	// Get background function
	fn := r.L.GetGlobal("background")
	if fn.Type() != lua.LTFunction {
		r.mu.Unlock()
		return true, 0, nil
	}
	bgFn := fn.(*lua.LFunction)

	// Create new coroutine if needed
	if r.bgThread == nil {
		r.bgThread, r.bgThreadCancel = r.L.NewThread()
		r.bgFunc = bgFn
	}

	// Prepare resume arguments
	var resumeArgs []lua.LValue
	if r.bgFunc != nil {
		// First resume - pass function and state
		resumeArgs = []lua.LValue{r.bgFunc, r.state}
		r.bgFunc = nil // Clear so subsequent resumes don't pass function again
	} else {
		// Subsequent resume - no function needed
		resumeArgs = []lua.LValue{nil}
	}

	r.mu.Unlock() // Release mutex during Lua execution

	// Resume the coroutine (this may take time)
	var status lua.ResumeState
	var err error
	var values []lua.LValue

	if len(resumeArgs) > 1 {
		// First resume - pass function and state
		status, err, values = r.L.Resume(r.bgThread, resumeArgs[0].(*lua.LFunction), resumeArgs[1])
	} else {
		// Subsequent resume - no function needed
		status, err, values = r.L.Resume(r.bgThread, nil)
	}

	r.mu.Lock() // Re-acquire mutex for state updates
	defer r.mu.Unlock()

	if err != nil {
		return false, 0, err
	}

	if status == lua.ResumeOK {
		// Coroutine finished
		return true, 0, nil
	}

	// Coroutine yielded - check if sleep duration was passed
	sleepMs := 0
	if len(values) > 0 {
		if n, ok := values[0].(lua.LNumber); ok {
			sleepMs = int(n)
		}
	}

	return false, sleepMs, nil
}

// callBackground executes the background function (legacy, used if no coroutine).
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
