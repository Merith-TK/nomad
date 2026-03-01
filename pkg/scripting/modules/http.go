package modules

import (
	"io"
	"net/http"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// HTTPModule provides HTTP functionality.
type HTTPModule struct{}

// NewHTTPModule creates a new HTTP module.
func NewHTTPModule() *HTTPModule {
	return &HTTPModule{}
}

// Loader returns the Lua module loader function.
func (m *HTTPModule) Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"get":     m.httpGet,
		"post":    m.httpPost,
		"request": m.httpRequest,
	})
	L.Push(mod)
	return 1
}

func (m *HTTPModule) httpGet(L *lua.LState) int {
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

func (m *HTTPModule) httpPost(L *lua.LState) int {
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

func (m *HTTPModule) httpRequest(L *lua.LState) int {
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
