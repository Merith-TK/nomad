package modules

import (
	"image/color"

	lua "github.com/yuin/gopher-lua"
)

// StreamDeckModule provides StreamDeck control.
type StreamDeckModule struct {
	device interface{} // Would be *streamdeck.Device in real implementation
}

// NewStreamDeckModule creates a new StreamDeck module.
func NewStreamDeckModule(device interface{}) *StreamDeckModule {
	return &StreamDeckModule{device: device}
}

// Loader returns the Lua module loader function.
func (m *StreamDeckModule) Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"set_color":      m.sdSetColor,
		"set_brightness": m.sdSetBrightness,
		"clear":          m.sdClear,
		"clear_key":      m.sdClearKey,
		"reset":          m.sdReset,
		"get_model":      m.sdGetModel,
		"get_keys":       m.sdGetKeys,
		"get_layout":     m.sdGetLayout,
	})
	L.Push(mod)
	return 1
}

func (m *StreamDeckModule) sdSetColor(L *lua.LState) int {
	if m.device == nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString("no device connected"))
		return 2
	}

	key := L.CheckInt(1)
	red := L.CheckInt(2)
	g := L.CheckInt(3)
	b := L.CheckInt(4)

	// In real implementation, this would call device.SetKeyColor
	_ = key // Key index for the button
	_ = color.RGBA{R: uint8(red), G: uint8(g), B: uint8(b), A: 255}

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

func (m *StreamDeckModule) sdSetBrightness(L *lua.LState) int {
	if m.device == nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString("no device connected"))
		return 2
	}

	percent := L.CheckInt(1)

	// In real implementation, this would call device.SetBrightness
	_ = percent

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

func (m *StreamDeckModule) sdClear(L *lua.LState) int {
	if m.device == nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString("no device connected"))
		return 2
	}

	// In real implementation, this would call device.Clear

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

func (m *StreamDeckModule) sdClearKey(L *lua.LState) int {
	if m.device == nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString("no device connected"))
		return 2
	}

	// key := L.CheckInt(1) - Not needed for clear operation

	// In real implementation, this would call device.SetKeyColor with black

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

func (m *StreamDeckModule) sdReset(L *lua.LState) int {
	if m.device == nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString("no device connected"))
		return 2
	}

	// In real implementation, this would call device.Reset

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

func (m *StreamDeckModule) sdGetModel(L *lua.LState) int {
	if m.device == nil {
		L.Push(lua.LNil)
		return 1
	}

	// In real implementation, this would return device.Model.Name
	L.Push(lua.LString("Stream Deck"))
	return 1
}

func (m *StreamDeckModule) sdGetKeys(L *lua.LState) int {
	if m.device == nil {
		L.Push(lua.LNumber(0))
		return 1
	}

	// In real implementation, this would return device.Model.Keys
	L.Push(lua.LNumber(15))
	return 1
}

func (m *StreamDeckModule) sdGetLayout(L *lua.LState) int {
	if m.device == nil {
		L.Push(lua.LNumber(0))
		L.Push(lua.LNumber(0))
		return 2
	}

	// In real implementation, this would return device.Model.Cols, device.Model.Rows
	L.Push(lua.LNumber(5))
	L.Push(lua.LNumber(3))
	return 2
}
