// Package scripting provides Lua script execution for the nomad system.
package scripting

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// ScriptContext holds context passed to Lua scripts.
type ScriptContext struct {
	ScriptPath string                 // Path to the script being executed
	ConfigDir  string                 // Base config directory
	Extra      map[string]interface{} // Additional context data
}

// Executor runs Lua scripts with the nomad API available.
type Executor struct {
	ctx ScriptContext
}

// NewExecutor creates a new Lua script executor.
func NewExecutor(ctx ScriptContext) *Executor {
	return &Executor{ctx: ctx}
}

// RunFile executes a Lua script file.
func (e *Executor) RunFile(path string) error {
	L := lua.NewState()
	defer L.Close()

	// Register modules
	e.registerModules(L)

	// Set script context globals
	L.SetGlobal("SCRIPT_PATH", lua.LString(path))
	L.SetGlobal("CONFIG_DIR", lua.LString(e.ctx.ConfigDir))

	// Execute the script
	if err := L.DoFile(path); err != nil {
		return fmt.Errorf("lua error: %w", err)
	}

	return nil
}

// RunString executes a Lua script string.
func (e *Executor) RunString(script string) error {
	L := lua.NewState()
	defer L.Close()

	e.registerModules(L)

	if err := L.DoString(script); err != nil {
		return fmt.Errorf("lua error: %w", err)
	}

	return nil
}

// registerModules adds all nomad-specific modules to the Lua state.
func (e *Executor) registerModules(L *lua.LState) {
	// Shell module - for executing system commands
	L.PreloadModule("shell", e.loaderShell)

	// HTTP module - for web requests
	L.PreloadModule("http", e.loaderHTTP)

	// System module - OS info and utilities
	L.PreloadModule("system", e.loaderSystem)

	// TODO: StreamDeck module will be added when device reference is available
}

// loaderShell provides shell command execution.
func (e *Executor) loaderShell(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"exec":       e.shellExec,
		"exec_async": e.shellExecAsync,
		"open":       e.shellOpen,
	})
	L.Push(mod)
	return 1
}

// shellExec runs a command and returns (stdout, stderr, exit_code).
func (e *Executor) shellExec(L *lua.LState) int {
	cmdStr := L.CheckString(1)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", cmdStr)
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}

	stdout, err := cmd.Output()
	exitCode := 0
	stderrStr := ""

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			stderrStr = string(exitErr.Stderr)
		} else {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			L.Push(lua.LNumber(-1))
			return 3
		}
	}

	L.Push(lua.LString(string(stdout)))
	L.Push(lua.LString(stderrStr))
	L.Push(lua.LNumber(exitCode))
	return 3
}

// shellExecAsync runs a command in the background (fire and forget).
func (e *Executor) shellExecAsync(L *lua.LState) int {
	cmdStr := L.CheckString(1)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", cmdStr)
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}

	err := cmd.Start()
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Don't wait for completion
	go cmd.Wait()

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

// shellOpen opens a file/URL with the system default application.
func (e *Executor) shellOpen(L *lua.LState) int {
	target := L.CheckString(1)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", target)
	case "darwin":
		cmd = exec.Command("open", target)
	default:
		cmd = exec.Command("xdg-open", target)
	}

	err := cmd.Start()
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	go cmd.Wait()

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

// loaderHTTP provides HTTP request functionality.
func (e *Executor) loaderHTTP(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"get":     e.httpGet,
		"post":    e.httpPost,
		"request": e.httpRequest,
	})
	L.Push(mod)
	return 1
}

// httpGet performs a simple GET request.
func (e *Executor) httpGet(L *lua.LState) int {
	url := L.CheckString(1)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Return body, status_code
	L.Push(lua.LString(string(body)))
	L.Push(lua.LNumber(resp.StatusCode))
	return 2
}

// httpPost performs a POST request with body.
func (e *Executor) httpPost(L *lua.LState) int {
	url := L.CheckString(1)
	contentType := L.CheckString(2)
	body := L.CheckString(3)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(url, contentType, strings.NewReader(body))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LString(string(respBody)))
	L.Push(lua.LNumber(resp.StatusCode))
	return 2
}

// httpRequest performs a custom HTTP request.
// Usage: http.request(method, url, headers_table, body)
func (e *Executor) httpRequest(L *lua.LState) int {
	method := L.CheckString(1)
	url := L.CheckString(2)
	headers := L.OptTable(3, nil)
	body := L.OptString(4, "")

	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Add headers if provided
	if headers != nil {
		headers.ForEach(func(key, value lua.LValue) {
			req.Header.Set(key.String(), value.String())
		})
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LString(string(respBody)))
	L.Push(lua.LNumber(resp.StatusCode))
	return 2
}

// loaderSystem provides system information and utilities.
func (e *Executor) loaderSystem(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"os":       e.systemOS,
		"env":      e.systemEnv,
		"sleep":    e.systemSleep,
		"hostname": e.systemHostname,
	})
	L.Push(mod)
	return 1
}

// systemOS returns the current OS name.
func (e *Executor) systemOS(L *lua.LState) int {
	L.Push(lua.LString(runtime.GOOS))
	return 1
}

// systemEnv gets an environment variable.
func (e *Executor) systemEnv(L *lua.LState) int {
	key := L.CheckString(1)
	value := os.Getenv(key)
	if value == "" {
		L.Push(lua.LNil)
	} else {
		L.Push(lua.LString(value))
	}
	return 1
}

// systemSleep pauses execution for N milliseconds.
func (e *Executor) systemSleep(L *lua.LState) int {
	ms := L.CheckInt(1)
	time.Sleep(time.Duration(ms) * time.Millisecond)
	return 0
}

// systemHostname returns the system hostname.
func (e *Executor) systemHostname(L *lua.LState) int {
	name, err := os.Hostname()
	if err != nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(lua.LString(name))
	return 1
}
