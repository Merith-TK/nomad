package modules

import (
	"os"
	"runtime"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// SystemModule provides system utilities.
type SystemModule struct{}

// NewSystemModule creates a new system module.
func NewSystemModule() *SystemModule {
	return &SystemModule{}
}

// Loader returns the Lua module loader function.
func (m *SystemModule) Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"os":       m.systemOS,
		"env":      m.systemEnv,
		"sleep":    m.systemSleep,
		"hostname": m.systemHostname,
		"refresh":  m.systemRefresh, // Trigger display refresh
	})
	L.Push(mod)
	return 1
}

func (m *SystemModule) systemOS(L *lua.LState) int {
	L.Push(lua.LString(runtime.GOOS))
	return 1
}

func (m *SystemModule) systemEnv(L *lua.LState) int {
	key := L.CheckString(1)
	value := os.Getenv(key)
	if value == "" {
		L.Push(lua.LNil)
	} else {
		L.Push(lua.LString(value))
	}
	return 1
}

func (m *SystemModule) systemSleep(L *lua.LState) int {
	ms := L.CheckInt(1)

	// Check if we're in a coroutine (background thread)
	// If so, yield to let Go handle the sleep without blocking
	// Note: Don't use mutex here - we may already be holding it from runBackgroundCoroutine
	// Direct pointer compare is safe since bgThread is only set while holding mutex
	// This is a simplified version - in practice, we'd need access to the runner's state
	isCoroutine := false // TODO: Pass runner context or find another way

	if isCoroutine {
		// Yield with sleep duration - Go will wait and resume
		return L.Yield(lua.LNumber(ms))
	}

	// Not in coroutine (trigger/passive) - do a brief sleep
	// Keep it short to avoid blocking
	if ms > 100 {
		ms = 100 // Cap at 100ms for non-coroutine calls
	}
	time.Sleep(time.Duration(ms) * time.Millisecond)
	return 0
}

func (m *SystemModule) systemHostname(L *lua.LState) int {
	name, err := os.Hostname()
	if err != nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(lua.LString(name))
	return 1
}

func (m *SystemModule) systemRefresh(L *lua.LState) int {
	// Request a display refresh (non-blocking)
	// This would need access to the runner's refresh callback
	// For now, this is a placeholder
	return 0
}
