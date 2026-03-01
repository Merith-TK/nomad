package modules

import (
	"time"

	lua "github.com/yuin/gopher-lua"
)

// TimeModule provides time and date utilities for Lua scripts.
type TimeModule struct{}

// NewTimeModule creates a new time module.
func NewTimeModule() *TimeModule {
	return &TimeModule{}
}

// Loader returns the Lua module loader function.
func (m *TimeModule) Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"now":       m.timeNow,
		"format":    m.timeFormat,
		"parse":     m.timeParse,
		"sleep":     m.timeSleep,
		"timestamp": m.timeTimestamp,
		"date":      m.timeDate,
	})
	L.Push(mod)
	return 1
}

func (m *TimeModule) timeNow(L *lua.LState) int {
	now := time.Now()
	L.Push(lua.LNumber(now.Unix()))
	return 1
}

func (m *TimeModule) timeFormat(L *lua.LState) int {
	timestamp := L.CheckNumber(1)
	layout := L.CheckString(2)

	t := time.Unix(int64(timestamp), 0)
	formatted := t.Format(layout)

	L.Push(lua.LString(formatted))
	return 1
}

func (m *TimeModule) timeParse(L *lua.LState) int {
	layout := L.CheckString(1)
	value := L.CheckString(2)

	t, err := time.Parse(layout, value)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LNumber(t.Unix()))
	L.Push(lua.LNil)
	return 2
}

func (m *TimeModule) timeSleep(L *lua.LState) int {
	ms := L.CheckNumber(1)

	// Convert to duration and sleep
	duration := time.Duration(ms) * time.Millisecond
	time.Sleep(duration)

	return 0
}

func (m *TimeModule) timeTimestamp(L *lua.LState) int {
	L.Push(lua.LNumber(time.Now().Unix()))
	return 1
}

func (m *TimeModule) timeDate(L *lua.LState) int {
	timestamp := L.OptNumber(1, lua.LNumber(time.Now().Unix()))

	t := time.Unix(int64(timestamp), 0)

	// Return table with date components
	tbl := L.NewTable()
	tbl.RawSetString("year", lua.LNumber(t.Year()))
	tbl.RawSetString("month", lua.LNumber(t.Month()))
	tbl.RawSetString("day", lua.LNumber(t.Day()))
	tbl.RawSetString("hour", lua.LNumber(t.Hour()))
	tbl.RawSetString("minute", lua.LNumber(t.Minute()))
	tbl.RawSetString("second", lua.LNumber(t.Second()))
	tbl.RawSetString("weekday", lua.LNumber(t.Weekday()))
	tbl.RawSetString("yearday", lua.LNumber(t.YearDay()))

	L.Push(tbl)
	return 1
}
