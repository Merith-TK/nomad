package modules

import (
	"io/ioutil"
	"os"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
)

// FileModule provides file system operations for Lua scripts.
type FileModule struct{}

// NewFileModule creates a new file module.
func NewFileModule() *FileModule {
	return &FileModule{}
}

// Loader returns the Lua module loader function.
func (m *FileModule) Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"read":   m.fileRead,
		"write":  m.fileWrite,
		"append": m.fileAppend,
		"exists": m.fileExists,
		"mkdir":  m.fileMkdir,
		"list":   m.fileList,
		"remove": m.fileRemove,
		"size":   m.fileSize,
		"is_dir": m.fileIsDir,
	})
	L.Push(mod)
	return 1
}

func (m *FileModule) fileRead(L *lua.LState) int {
	path := L.CheckString(1)

	// Security: restrict to config directory and subdirectories
	configDir := L.GetGlobal("CONFIG_DIR").String()
	if configDir != "" {
		if !filepath.HasPrefix(filepath.Clean(path), filepath.Clean(configDir)) {
			L.Push(lua.LNil)
			L.Push(lua.LString("Access denied: can only read files within config directory"))
			return 2
		}
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LString(string(data)))
	L.Push(lua.LNil)
	return 2
}

func (m *FileModule) fileWrite(L *lua.LState) int {
	path := L.CheckString(1)
	content := L.CheckString(2)

	// Security: restrict to config directory and subdirectories
	configDir := L.GetGlobal("CONFIG_DIR").String()
	if configDir != "" {
		if !filepath.HasPrefix(filepath.Clean(path), filepath.Clean(configDir)) {
			L.Push(lua.LFalse)
			L.Push(lua.LString("Access denied: can only write files within config directory"))
			return 2
		}
	}

	err := ioutil.WriteFile(path, []byte(content), 0644)
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

func (m *FileModule) fileAppend(L *lua.LState) int {
	path := L.CheckString(1)
	content := L.CheckString(2)

	// Security: restrict to config directory and subdirectories
	configDir := L.GetGlobal("CONFIG_DIR").String()
	if configDir != "" {
		if !filepath.HasPrefix(filepath.Clean(path), filepath.Clean(configDir)) {
			L.Push(lua.LFalse)
			L.Push(lua.LString("Access denied: can only write files within config directory"))
			return 2
		}
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

func (m *FileModule) fileExists(L *lua.LState) int {
	path := L.CheckString(1)

	// Security: restrict to config directory and subdirectories
	configDir := L.GetGlobal("CONFIG_DIR").String()
	if configDir != "" {
		if !filepath.HasPrefix(filepath.Clean(path), filepath.Clean(configDir)) {
			L.Push(lua.LFalse)
			return 1
		}
	}

	_, err := os.Stat(path)
	L.Push(lua.LBool(err == nil))
	return 1
}

func (m *FileModule) fileMkdir(L *lua.LState) int {
	path := L.CheckString(1)

	// Security: restrict to config directory and subdirectories
	configDir := L.GetGlobal("CONFIG_DIR").String()
	if configDir != "" {
		if !filepath.HasPrefix(filepath.Clean(path), filepath.Clean(configDir)) {
			L.Push(lua.LFalse)
			L.Push(lua.LString("Access denied: can only create directories within config directory"))
			return 2
		}
	}

	err := os.MkdirAll(path, 0755)
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

func (m *FileModule) fileList(L *lua.LState) int {
	path := L.CheckString(1)

	// Security: restrict to config directory and subdirectories
	configDir := L.GetGlobal("CONFIG_DIR").String()
	if configDir != "" {
		if !filepath.HasPrefix(filepath.Clean(path), filepath.Clean(configDir)) {
			L.Push(lua.LNil)
			L.Push(lua.LString("Access denied: can only list files within config directory"))
			return 2
		}
	}

	entries, err := ioutil.ReadDir(path)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	tbl := L.NewTable()
	for i, entry := range entries {
		entryTbl := L.NewTable()
		entryTbl.RawSetString("name", lua.LString(entry.Name()))
		entryTbl.RawSetString("is_dir", lua.LBool(entry.IsDir()))
		entryTbl.RawSetString("size", lua.LNumber(entry.Size()))
		tbl.RawSetInt(i+1, entryTbl)
	}

	L.Push(tbl)
	L.Push(lua.LNil)
	return 2
}

func (m *FileModule) fileRemove(L *lua.LState) int {
	path := L.CheckString(1)

	// Security: restrict to config directory and subdirectories
	configDir := L.GetGlobal("CONFIG_DIR").String()
	if configDir != "" {
		if !filepath.HasPrefix(filepath.Clean(path), filepath.Clean(configDir)) {
			L.Push(lua.LFalse)
			L.Push(lua.LString("Access denied: can only remove files within config directory"))
			return 2
		}
	}

	err := os.Remove(path)
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

func (m *FileModule) fileSize(L *lua.LState) int {
	path := L.CheckString(1)

	// Security: restrict to config directory and subdirectories
	configDir := L.GetGlobal("CONFIG_DIR").String()
	if configDir != "" {
		if !filepath.HasPrefix(filepath.Clean(path), filepath.Clean(configDir)) {
			L.Push(lua.LNumber(-1))
			return 1
		}
	}

	info, err := os.Stat(path)
	if err != nil {
		L.Push(lua.LNumber(-1))
		return 1
	}

	L.Push(lua.LNumber(info.Size()))
	return 1
}

func (m *FileModule) fileIsDir(L *lua.LState) int {
	path := L.CheckString(1)

	// Security: restrict to config directory and subdirectories
	configDir := L.GetGlobal("CONFIG_DIR").String()
	if configDir != "" {
		if !filepath.HasPrefix(filepath.Clean(path), filepath.Clean(configDir)) {
			L.Push(lua.LFalse)
			return 1
		}
	}

	info, err := os.Stat(path)
	if err != nil {
		L.Push(lua.LFalse)
		return 1
	}

	L.Push(lua.LBool(info.IsDir()))
	return 1
}
