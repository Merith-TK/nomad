package modules

import (
	"fmt"
	"log"
	"os"

	lua "github.com/yuin/gopher-lua"
)

// LogModule provides logging utilities for Lua scripts.
type LogModule struct {
	logger *log.Logger
}

// NewLogModule creates a new log module.
func NewLogModule() *LogModule {
	return &LogModule{
		logger: log.New(os.Stdout, "[SCRIPT] ", log.LstdFlags),
	}
}

// Loader returns the Lua module loader function.
func (m *LogModule) Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"info":   m.logInfo,
		"warn":   m.logWarn,
		"error":  m.logError,
		"debug":  m.logDebug,
		"printf": m.logPrintf,
	})
	L.Push(mod)
	return 1
}

func (m *LogModule) logInfo(L *lua.LState) int {
	message := L.CheckString(1)
	m.logger.Println("[INFO]", message)
	return 0
}

func (m *LogModule) logWarn(L *lua.LState) int {
	message := L.CheckString(1)
	m.logger.Println("[WARN]", message)
	return 0
}

func (m *LogModule) logError(L *lua.LState) int {
	message := L.CheckString(1)
	m.logger.Println("[ERROR]", message)
	return 0
}

func (m *LogModule) logDebug(L *lua.LState) int {
	message := L.CheckString(1)
	m.logger.Println("[DEBUG]", message)
	return 0
}

func (m *LogModule) logPrintf(L *lua.LState) int {
	format := L.CheckString(1)

	// Collect remaining arguments
	args := make([]interface{}, L.GetTop()-1)
	for i := 2; i <= L.GetTop(); i++ {
		args[i-2] = luaValueToString(L.Get(i))
	}

	message := fmt.Sprintf(format, args...)
	m.logger.Println(message)
	return 0
}

// Helper function to convert Lua values to strings for logging
func luaValueToString(value lua.LValue) interface{} {
	switch v := value.(type) {
	case *lua.LNilType:
		return "nil"
	case lua.LBool:
		return bool(v)
	case lua.LNumber:
		return float64(v)
	case lua.LString:
		return string(v)
	case *lua.LTable:
		return "[table]"
	default:
		return v.String()
	}
}
