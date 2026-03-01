package modules

import (
	"encoding/json"
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// JSONModule provides JSON encoding/decoding for Lua scripts.
type JSONModule struct{}

// NewJSONModule creates a new JSON module.
func NewJSONModule() *JSONModule {
	return &JSONModule{}
}

// Loader returns the Lua module loader function.
func (m *JSONModule) Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"encode": m.jsonEncode,
		"decode": m.jsonDecode,
	})
	L.Push(mod)
	return 1
}

func (m *JSONModule) jsonEncode(L *lua.LState) int {
	value := L.Get(1)

	// Convert Lua value to Go value
	goValue := luaValueToGo(value)

	// Encode to JSON
	jsonBytes, err := json.Marshal(goValue)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("JSON encode error: %v", err)))
		return 2
	}

	L.Push(lua.LString(string(jsonBytes)))
	L.Push(lua.LNil)
	return 2
}

func (m *JSONModule) jsonDecode(L *lua.LState) int {
	jsonStr := L.CheckString(1)

	var result interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("JSON decode error: %v", err)))
		return 2
	}

	// Convert Go value back to Lua value
	luaValue := goValueToLua(L, result)
	L.Push(luaValue)
	L.Push(lua.LNil)
	return 2
}

// Helper functions for Lua/Go value conversion
func luaValueToGo(value lua.LValue) interface{} {
	switch v := value.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return bool(v)
	case lua.LNumber:
		return float64(v)
	case lua.LString:
		return string(v)
	case *lua.LTable:
		// Check if it's an array or object
		isArray := true
		maxKey := 0
		v.ForEach(func(key, val lua.LValue) {
			if key.Type() == lua.LTNumber {
				if int(key.(lua.LNumber)) > maxKey {
					maxKey = int(key.(lua.LNumber))
				}
			} else {
				isArray = false
			}
		})

		if isArray && maxKey > 0 {
			// Array
			arr := make([]interface{}, maxKey)
			for i := 1; i <= maxKey; i++ {
				arr[i-1] = luaValueToGo(v.RawGetInt(i))
			}
			return arr
		} else {
			// Object
			obj := make(map[string]interface{})
			v.ForEach(func(key, val lua.LValue) {
				if key.Type() == lua.LTString {
					obj[key.String()] = luaValueToGo(val)
				}
			})
			return obj
		}
	default:
		return nil
	}
}

func goValueToLua(L *lua.LState, value interface{}) lua.LValue {
	switch v := value.(type) {
	case nil:
		return lua.LNil
	case bool:
		return lua.LBool(v)
	case float64:
		return lua.LNumber(v)
	case int:
		return lua.LNumber(v)
	case string:
		return lua.LString(v)
	case []interface{}:
		tbl := L.NewTable()
		for i, item := range v {
			tbl.RawSetInt(i+1, goValueToLua(L, item))
		}
		return tbl
	case map[string]interface{}:
		tbl := L.NewTable()
		for key, val := range v {
			tbl.RawSetString(key, goValueToLua(L, val))
		}
		return tbl
	default:
		return lua.LString(fmt.Sprintf("%v", v))
	}
}
