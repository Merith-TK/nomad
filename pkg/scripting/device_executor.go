package scripting

import (
	"fmt"

	"github.com/merith-tk/nomad/pkg/streamdeck"
	lua "github.com/yuin/gopher-lua"
)

// DeviceExecutor runs Lua scripts with StreamDeck API access.
type DeviceExecutor struct {
	ctx    ScriptContext
	device *streamdeck.Device
	sdAPI  *StreamDeckAPI
}

// NewDeviceExecutor creates an executor with StreamDeck support.
func NewDeviceExecutor(ctx ScriptContext, dev *streamdeck.Device) *DeviceExecutor {
	return &DeviceExecutor{
		ctx:    ctx,
		device: dev,
		sdAPI:  NewStreamDeckAPI(dev),
	}
}

// RunFile executes a Lua script with all APIs available.
func (e *DeviceExecutor) RunFile(path string) error {
	L := lua.NewState()
	defer L.Close()

	// Register base modules
	baseExec := &Executor{ctx: e.ctx}
	baseExec.registerModules(L)

	// Register StreamDeck module
	e.sdAPI.RegisterModule(L)

	// Set globals
	L.SetGlobal("SCRIPT_PATH", lua.LString(path))
	L.SetGlobal("CONFIG_DIR", lua.LString(e.ctx.ConfigDir))

	// Execute
	if err := L.DoFile(path); err != nil {
		return fmt.Errorf("lua error in %s: %w", path, err)
	}

	return nil
}

// RunString executes a Lua string with all APIs.
func (e *DeviceExecutor) RunString(script string) error {
	L := lua.NewState()
	defer L.Close()

	baseExec := &Executor{ctx: e.ctx}
	baseExec.registerModules(L)
	e.sdAPI.RegisterModule(L)

	if err := L.DoString(script); err != nil {
		return fmt.Errorf("lua error: %w", err)
	}

	return nil
}
