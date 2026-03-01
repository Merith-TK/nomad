package scripting

import (
	"fmt"
	"image/color"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// loaderShell provides shell command execution for ScriptRunner.
func (r *ScriptRunner) loaderShell(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"exec":       r.shellExec,
		"exec_async": r.shellExecAsync,
		"open":       r.shellOpen,
		"terminal":   r.shellTerminal,
	})
	L.Push(mod)
	return 1
}

func (r *ScriptRunner) shellExec(L *lua.LState) int {
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

func (r *ScriptRunner) shellExecAsync(L *lua.LState) int {
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

	go cmd.Wait()

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

func (r *ScriptRunner) shellOpen(L *lua.LState) int {
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

// shellTerminal opens a new terminal window and runs the command in it.
// On Windows: opens cmd.exe or Windows Terminal
// On macOS: opens Terminal.app
// On Linux: tries common terminal emulators
func (r *ScriptRunner) shellTerminal(L *lua.LState) int {
	cmdStr := L.CheckString(1)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// "start cmd /k <command>" opens a new cmd window and runs the command
		// Pass as single string to cmd /c to avoid argument parsing issues
		cmd = exec.Command("cmd", "/c", "start", cmdStr)
		fmt.Println("terminal run call:", cmdStr)
	case "darwin":
		// Use osascript to open Terminal and run command
		script := `tell application "Terminal" to do script "` + strings.ReplaceAll(cmdStr, `"`, `\"`) + `"`
		cmd = exec.Command("osascript", "-e", script)
	default:
		// Linux: try common terminal emulators
		terminals := [][]string{
			{"x-terminal-emulator", "-e"},
			{"gnome-terminal", "--"},
			{"konsole", "-e"},
			{"xfce4-terminal", "-e"},
			{"xterm", "-e"},
		}
		for _, term := range terminals {
			if _, err := exec.LookPath(term[0]); err == nil {
				args := append(term[1:], "sh", "-c", cmdStr)
				cmd = exec.Command(term[0], args...)
				break
			}
		}
		if cmd == nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString("no terminal emulator found"))
			return 2
		}
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

// loaderHTTP provides HTTP functionality.
func (r *ScriptRunner) loaderHTTP(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"get":     r.httpGet,
		"post":    r.httpPost,
		"request": r.httpRequest,
	})
	L.Push(mod)
	return 1
}

func (r *ScriptRunner) httpGet(L *lua.LState) int {
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

	L.Push(lua.LString(string(body)))
	L.Push(lua.LNumber(resp.StatusCode))
	return 2
}

func (r *ScriptRunner) httpPost(L *lua.LState) int {
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

func (r *ScriptRunner) httpRequest(L *lua.LState) int {
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

// loaderSystem provides system utilities.
func (r *ScriptRunner) loaderSystem(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"os":       r.systemOS,
		"env":      r.systemEnv,
		"sleep":    r.systemSleep,
		"hostname": r.systemHostname,
		"refresh":  r.systemRefresh, // Trigger display refresh
	})
	L.Push(mod)
	return 1
}

func (r *ScriptRunner) systemOS(L *lua.LState) int {
	L.Push(lua.LString(runtime.GOOS))
	return 1
}

func (r *ScriptRunner) systemEnv(L *lua.LState) int {
	key := L.CheckString(1)
	value := os.Getenv(key)
	if value == "" {
		L.Push(lua.LNil)
	} else {
		L.Push(lua.LString(value))
	}
	return 1
}

func (r *ScriptRunner) systemSleep(L *lua.LState) int {
	ms := L.CheckInt(1)
	time.Sleep(time.Duration(ms) * time.Millisecond)
	return 0
}

func (r *ScriptRunner) systemHostname(L *lua.LState) int {
	name, err := os.Hostname()
	if err != nil {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(lua.LString(name))
	return 1
}

func (r *ScriptRunner) systemRefresh(L *lua.LState) int {
	// Request a display refresh (non-blocking)
	go r.requestRefresh()
	return 0
}

// loaderStreamDeck provides StreamDeck control.
func (r *ScriptRunner) loaderStreamDeck(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"set_color":      r.sdSetColor,
		"set_brightness": r.sdSetBrightness,
		"clear":          r.sdClear,
		"clear_key":      r.sdClearKey,
		"reset":          r.sdReset,
		"get_model":      r.sdGetModel,
		"get_keys":       r.sdGetKeys,
		"get_layout":     r.sdGetLayout,
	})
	L.Push(mod)
	return 1
}

func (r *ScriptRunner) sdSetColor(L *lua.LState) int {
	if r.device == nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString("no device connected"))
		return 2
	}

	key := L.CheckInt(1)
	red := L.CheckInt(2)
	g := L.CheckInt(3)
	b := L.CheckInt(4)

	c := color.RGBA{R: uint8(red), G: uint8(g), B: uint8(b), A: 255}
	err := r.device.SetKeyColor(key, c)
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

func (r *ScriptRunner) sdSetBrightness(L *lua.LState) int {
	if r.device == nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString("no device connected"))
		return 2
	}

	percent := L.CheckInt(1)
	err := r.device.SetBrightness(percent)
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

func (r *ScriptRunner) sdClear(L *lua.LState) int {
	if r.device == nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString("no device connected"))
		return 2
	}

	err := r.device.Clear()
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

func (r *ScriptRunner) sdClearKey(L *lua.LState) int {
	if r.device == nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString("no device connected"))
		return 2
	}

	key := L.CheckInt(1)
	err := r.device.SetKeyColor(key, color.RGBA{0, 0, 0, 255})
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

func (r *ScriptRunner) sdReset(L *lua.LState) int {
	if r.device == nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString("no device connected"))
		return 2
	}

	err := r.device.Reset()
	if err != nil {
		L.Push(lua.LFalse)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(lua.LTrue)
	L.Push(lua.LNil)
	return 2
}

func (r *ScriptRunner) sdGetModel(L *lua.LState) int {
	if r.device == nil {
		L.Push(lua.LNil)
		return 1
	}

	L.Push(lua.LString(r.device.Model.Name))
	return 1
}

func (r *ScriptRunner) sdGetKeys(L *lua.LState) int {
	if r.device == nil {
		L.Push(lua.LNumber(0))
		return 1
	}

	L.Push(lua.LNumber(r.device.Model.Keys))
	return 1
}

func (r *ScriptRunner) sdGetLayout(L *lua.LState) int {
	if r.device == nil {
		L.Push(lua.LNumber(0))
		L.Push(lua.LNumber(0))
		return 2
	}

	L.Push(lua.LNumber(r.device.Model.Cols))
	L.Push(lua.LNumber(r.device.Model.Rows))
	return 2
}
