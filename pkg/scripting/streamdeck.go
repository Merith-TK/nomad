package scripting

import (
	"image/color"

	"github.com/merith-tk/nomad/pkg/streamdeck"
	lua "github.com/yuin/gopher-lua"
)

// StreamDeckAPI provides Lua bindings for StreamDeck control.
type StreamDeckAPI struct {
	device *streamdeck.Device
}

// NewStreamDeckAPI creates a StreamDeck API for Lua.
func NewStreamDeckAPI(dev *streamdeck.Device) *StreamDeckAPI {
	return &StreamDeckAPI{device: dev}
}

// RegisterModule adds the streamdeck module to the Lua state.
func (api *StreamDeckAPI) RegisterModule(L *lua.LState) {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"set_color":      api.setColor,
		"set_brightness": api.setBrightness,
		"clear":          api.clear,
		"clear_key":      api.clearKey,
		"reset":          api.reset,
		"get_model":      api.getModel,
		"get_keys":       api.getKeys,
		"get_layout":     api.getLayout,
	})
	L.PreloadModule("streamdeck", func(L *lua.LState) int {
		L.Push(mod)
		return 1
	})
}

// setColor sets a key's color.
// Usage: streamdeck.set_color(key_index, r, g, b)
func (api *StreamDeckAPI) setColor(L *lua.LState) int {
	if api.device == nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString("no device connected"))
		return 2
	}

	key := L.CheckInt(1)
	r := L.CheckInt(2)
	g := L.CheckInt(3)
	b := L.CheckInt(4)

	c := color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
	err := api.device.SetKeyColor(key, c)
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

// setBrightness sets the device brightness (0-100).
// Usage: streamdeck.set_brightness(percent)
func (api *StreamDeckAPI) setBrightness(L *lua.LState) int {
	if api.device == nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString("no device connected"))
		return 2
	}

	percent := L.CheckInt(1)
	err := api.device.SetBrightness(percent)
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

// clear clears all keys.
func (api *StreamDeckAPI) clear(L *lua.LState) int {
	if api.device == nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString("no device connected"))
		return 2
	}

	err := api.device.Clear()
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

// clearKey clears a single key.
// Usage: streamdeck.clear_key(key_index)
func (api *StreamDeckAPI) clearKey(L *lua.LState) int {
	if api.device == nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString("no device connected"))
		return 2
	}

	key := L.CheckInt(1)
	err := api.device.SetKeyColor(key, color.RGBA{0, 0, 0, 255})
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

// reset resets the device.
func (api *StreamDeckAPI) reset(L *lua.LState) int {
	if api.device == nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString("no device connected"))
		return 2
	}

	err := api.device.Reset()
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

// getModel returns the device model name.
func (api *StreamDeckAPI) getModel(L *lua.LState) int {
	if api.device == nil {
		L.Push(lua.LNil)
		return 1
	}

	L.Push(lua.LString(api.device.Model.Name))
	return 1
}

// getKeys returns the number of keys on the device.
func (api *StreamDeckAPI) getKeys(L *lua.LState) int {
	if api.device == nil {
		L.Push(lua.LNumber(0))
		return 1
	}

	L.Push(lua.LNumber(api.device.Model.Keys))
	return 1
}

// getLayout returns cols, rows of the device.
func (api *StreamDeckAPI) getLayout(L *lua.LState) int {
	if api.device == nil {
		L.Push(lua.LNumber(0))
		L.Push(lua.LNumber(0))
		return 2
	}

	L.Push(lua.LNumber(api.device.Model.Cols))
	L.Push(lua.LNumber(api.device.Model.Rows))
	return 2
}
