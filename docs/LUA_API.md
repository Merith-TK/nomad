# NOMAD Stream Deck  Lua Scripting API

Scripts live under `.nomad/interface/streamdeck/config/` and are organised into
subdirectories that become navigation folders on the device.

---

## Table of Contents

1. [Script Structure](#1-script-structure)
2. [Lifecycle Functions](#2-lifecycle-functions)
3. [Shared State](#3-shared-state)
4. [Global Variables](#4-global-variables)
5. [Restart Policy](#5-restart-policy)
6. [Special Files](#6-special-files)
7. [Performance Rules](#7-performance-rules)
8. [Appearance Table](#8-appearance-table)
9. [Module Reference](#9-module-reference)
   - [shell](#shell)
   - [system](#system)
   - [http](#http)
   - [streamdeck](#streamdeck)
   - [file](#file)
   - [store](#store)
   - [time](#time)
   - [json](#json)
   - [log](#log)
   - [utils](#utils)
   - [strings](#strings)
   - [pkg_data](#pkg_data-packages-only)
10. [Packages](#10-packages)
11. [Quick Reference](#11-quick-reference)
12. [Examples](#12-examples)

---

## 1. Script Structure

Every script **must** return a Lua table. The table may contain any combination
of the three lifecycle functions. Any file that does not return a table is
rejected at load time.

```lua
local system = require("system")
local log    = require("log")

local script = {}

function script.background(state) ... end  -- optional
function script.passive(key, state) ... end  -- optional
function script.trigger(state) ... end  -- optional

return script
```

> **There is no "global functions" mode.** The file must `return script`.

---

## 2. Lifecycle Functions

### `script.background(state)`

Runs once as a managed Lua coroutine.  
Use `system.sleep(ms)` to **yield**  other calls execute during the sleep window.  
Do **not** use `time.sleep()` inside background; it blocks the goroutine entirely.

```lua
function script.background(state)
    while true do
        local out, _, code = shell.exec("some-command")
        if code == 0 then
            state.value = out
            system.refresh()   -- tell passive() to repaint immediately
        end
        system.sleep(2000)     -- yield for 2 s, then loop
    end
end
```

The coroutine restarts according to [`RESTART_POLICY`](#5-restart-policy) if it exits or errors.

---

### `script.passive(key, state)  table|nil`

Called at the passive FPS (default **30 fps**) while the key is visible.  
Return an [appearance table](#8-appearance-table) or `nil` (leave key unchanged).

```lua
function script.passive(key, state)
    return {
        color = {0, 128, 255},
        text  = state.label or "?",
    }
end
```

**Rules:**
- Must return quickly (< 1 ms target).
- Never call `shell.exec`, `http.get`, `file.read`, or any other I/O.
- Use `background()` to fetch data; store it in `state`; read it here.

---

### `script.trigger(state)`

Called once when the physical key is pressed.  
After changing state, call `system.refresh()` to force an immediate repaint.

```lua
function script.trigger(state)
    state.active = not state.active
    system.refresh()
end
```

Avoid long blocking operations in trigger. For slow work, set a flag in `state`
and handle it in `background()`, or use `shell.exec_async()`.

---

## 3. Shared State

The `state` table is created once per script and passed to every lifecycle call.  
Use it to share data between background, passive, and trigger.

```lua
-- background writes:
state.cpu = "42%"
system.refresh()

-- passive reads:
return { text = state.cpu or "" }
```

Module-local variables (`local x = 0`) are also persistent for the lifetime of
the Lua state, but are reset if the script runner is restarted (e.g. on error
when `RESTART_POLICY = "always"`). Prefer `state` for data that must survive restarts.

---

## 4. Global Variables

These are injected into every script's Lua environment before execution:

| Variable | Type | Description |
|---|---|---|
| `SCRIPT_PATH` | string | Absolute path to the script file |
| `SCRIPT_NAME` | string | Filename without extension |
| `CONFIG_DIR` | string | Absolute path to the config root |

---

## 5. Restart Policy

Controls what happens when `background()` exits or errors:

```lua
RESTART_POLICY = "always"   -- restart immediately on exit or error (default)
RESTART_POLICY = "once"     -- restart once, then stop
RESTART_POLICY = "never"    -- do not restart; background runs exactly once
```

Place this at the top of the file (outside the returned table).

---

## 6. Special Files

### `_boot.lua`

Runs once at startup **before** any other scripts load.  
Has full access to all modules.  
**Do not call `system.sleep()`**  use `time.sleep(ms)` instead (truly blocking; safe here).

```lua
local log  = require("log")
local time = require("time")

log.info("Boot starting")
time.sleep(500)   -- safe blocking sleep
log.info("Boot complete")
```

---

### `.directory.lua`  Toggle / Dual-mode Keys

A `.directory.lua` file gives the directory key itself two states: **t1** (default
appearance) and **t2** (pressed appearance). It supports two extra lifecycle functions:

| Function | Role |
|---|---|
| `script.passive(key, state)` | Appearance when **not** pressed (t1) |
| `script.t1_passive(key, state)` | Alias for passive  t1 appearance |
| `script.t1_trigger(state)` | Called on first press (enter directory) |
| `script.t2_passive(key, state)` | Appearance while **in** the directory (t2) |
| `script.t2_trigger(state)` | Called on second press (exit directory) |

`background()` is shared between both states.

---

## 7. Performance Rules

| Location | Rule |
|---|---|
| `passive()` | **No I/O at all.** Only read `state` and return a table. |
| `background()` | All shell, http, and file work goes here. Use `system.sleep()` to pace. |
| `trigger()` | May do quick I/O. For slow work, set a flag and handle in background. |
| `_boot.lua` | Use `time.sleep()` (blocking). `system.sleep()` does not yield here. |

The passive loop runs at ~30 fps. If `passive()` is slow, every key on the active
page stutters. Keep it under 1 ms.

---

## 8. Appearance Table

Returned by `passive()` (and its variants). All fields are optional.

```lua
return {
    color      = {r, g, b},       -- background fill, 0-255 per channel
    text       = "label\nline2",  -- key label; \n for multi-line
    text_color = {r, g, b},       -- label colour (default: white {255,255,255})
    image      = "icon.png",      -- relative path from config root; overrides color/text
}
```

If `nil` is returned the key is left exactly as it was on the previous tick.

---

## 9. Module Reference

Require modules at the top of the file:

```lua
local shell = require("shell")
local log   = require("log")
```

---

### `shell`

Run external commands.

```lua
-- Synchronous: returns output, stderr, exit-code
local out, err, code = shell.exec("command --flag")

-- Fire-and-forget (non-blocking)
local ok, err = shell.exec_async("notify-send 'hello'")

-- Open a file/URL with the default OS handler
shell.open("/path/to/file.pdf")
shell.open("https://example.com")

-- Open a new terminal window running a command
shell.terminal("htop")
```

> `shell.exec` in `passive()` will cause severe lag. Use `background()` instead.

---

### `system`

OS introspection and script control.

```lua
local os_name  = system.os()        -- "linux" | "windows" | "darwin"
local home     = system.env("HOME") -- environment variable value or nil
local hostname = system.hostname()  -- current hostname string

system.sleep(ms)   -- yield the background coroutine (background() ONLY)
system.refresh()   -- request an immediate passive() repaint
```

> `system.sleep()` **only works inside `background()`**.  
> In `trigger()`, `_boot.lua`, or top-level code, use `time.sleep()` instead.

---

### `http`

HTTP client.

```lua
-- Simple GET
local body, err = http.get("https://api.example.com/data")

-- POST with explicit content-type
-- NOTE: arg order is (url, contentType, body)
local body, err = http.post("https://api.example.com/v1", "application/json", '{"key":"val"}')

-- Full control
local body, err = http.request(
    "PUT",                                      -- method
    "https://api.example.com/resource/1",       -- url
    {["Authorization"] = "Bearer TOKEN"},       -- headers table
    '{"active":true}',                          -- body string
    5000                                        -- timeout ms (optional, default 10000)
)
```

---

### `streamdeck`

Direct device control. Prefer returning an appearance table from `passive()`
instead of calling these from background/trigger  the runtime handles painting.
These are useful for one-shot overrides (e.g. boot animations).

```lua
streamdeck.set_color(key, r, g, b)  -- paint a single key by index
streamdeck.set_brightness(percent)  -- 0-100
streamdeck.clear()                  -- blank all keys
streamdeck.clear_key(key)           -- blank one key
streamdeck.reset()                  -- hardware reset

local model = streamdeck.get_model()        -- e.g. "StreamDeck MK.2"
local count = streamdeck.get_keys()         -- total key count
local cols, rows = streamdeck.get_layout()  -- grid dimensions
```

---

### `file`

General-purpose file I/O. Paths may be absolute or relative to the process
working directory.

```lua
local content, err = file.read("path/to/file.txt")
local ok, err      = file.write("out.txt", "hello\n")
local ok, err      = file.append("log.txt", "line\n")
local exists       = file.exists("path")       -- bool
local ok, err      = file.mkdir("new/dir")
local list, err    = file.list("dir")          -- array of entry names
local ok, err      = file.remove("file.txt")
local size, err    = file.size("file.txt")     -- bytes
local isdir        = file.is_dir("path")       -- bool
```

---

### `store`

A **cross-script** shared key-value store backed by a thread-safe Go `sync.Map`.
Values written by one script are immediately visible to all other scripts.

```lua
store.set("obs.streaming", true)
store.set("volume.level",  75)
store.set("mode",          "gaming")

local level   = store.get("volume.level")   -- number|string|boolean|nil
local exists  = store.has("mode")           -- bool
store.delete("mode")

local keys = store.keys()                   -- sorted array of key strings
for _, k in ipairs(keys) do
    print(k, store.get(k))
end
```

**Supported value types:** `string`, `number`, `bool`.  
Tables and functions cannot be stored  serialize with `json.encode` first.

---

### `time`

Time utilities.

```lua
local ts     = time.now()               -- Unix timestamp (seconds) as integer
local ts_ms  = time.timestamp()         -- Unix timestamp in milliseconds

local d      = time.date()              -- table for current time
local d_past = time.date(ts)            -- table for a given timestamp
-- d fields: year, month, day, hour, min, sec, weekday, yearday

local s      = time.format(ts, "2006-01-02 15:04:05")  -- Go time layout
local t, err = time.parse("2006-01-02", "2024-07-04")  -- returns timestamp or nil+err

time.sleep(ms)   -- true blocking sleep  safe anywhere (passive, trigger, _boot)
```

> Use `time.sleep()` in `_boot.lua` and `trigger()`.  
> Use `system.sleep()` in `background()` (it yields the coroutine).

---

### `json`

```lua
local s, err   = json.encode({key = "value", n = 42})  -- Lua table  JSON string
local t, err   = json.decode('{"key":"value","n":42}') -- JSON string  Lua table
```

Returns `nil, err` on failure. Always check `err`.

---

### `log`

Structured logging to the NOMAD log output.

```lua
log.info("message")
log.warn("message")
log.error("message")
log.debug("message")           -- only visible at debug log level
log.printf("value=%d", 42)     -- fmt.Sprintf-style formatted message
log.print("a", "b", "c")       -- space-joined, newline appended
```

---

### `utils`

Table helpers.

```lua
local copy   = utils.deepcopy(t)            -- recursive deep copy
local found  = utils.contains(t, value)     -- bool  searches array values
local n      = utils.size(t)                -- count of all keys (array + hash)
local merged = utils.merge(base, override)  -- shallow merge; override wins
```

---

### `strings`

String utilities.

```lua
strings.split("a,b,c", ",")        -- {"a","b","c"}
strings.trim("  hello  ")          -- "hello"
strings.startswith("hello", "he")  -- true
strings.endswith("hello", "lo")    -- true
strings.contains("hello", "ell")   -- true
strings.replace("hello", "l", "r") -- "herro"
strings.upper("hello")             -- "HELLO"
strings.lower("HELLO")             -- "hello"
strings.capitalize("hello world")  -- "Hello world"
strings.titlecase("hello world")   -- "Hello World"
```

---

### `pkg_data` (packages only)

Available only inside package scripts. Provides a sandboxed filesystem rooted at
the package's own data directory  scripts outside packages cannot use this module.

```lua
-- All file.* functions are available, scoped to the package data dir:
local content, err = pkg_data.read("cache.json")
local ok, err      = pkg_data.write("cache.json", data)
local ok, err      = pkg_data.append("log.txt", line)
local exists       = pkg_data.exists("cache.json")
local isdir        = pkg_data.is_dir("subdir")
local list, err    = pkg_data.list(".")
local ok, err      = pkg_data.mkdir("subdir")
local ok, err      = pkg_data.remove("old.txt")

-- JSON helpers (read/write + automatic encode/decode):
local t, err       = pkg_data.json_read("config.json")   -- JSON file  Lua table
local ok, err      = pkg_data.json_write("config.json", t) -- Lua table  JSON file

-- Resolve an absolute path (for passing to shell.exec etc.):
local abs          = pkg_data.path()           -- package data root
local abs          = pkg_data.path("subdir/file.txt")
```

---

## 10. Packages

A package bundles a daemon script, Lua library modules, and optional config
under a single directory in `.nomad/packages/<package-name>/`.

### Directory layout

```
.nomad/packages/my-package/
    manifest.json
    daemon.lua          -- background worker (runs independently of any key)
    lib/
        helpers.lua     -- require("my-package/helpers")
    data/               -- written by pkg_data; gitignored
```

### `manifest.json`

```json
{
    "name": "my-package",
    "version": "1.0.0",
    "description": "What this package does",
    "author": "you",
    "entry": "daemon.lua"
}
```

### Communicating with key scripts

Packages and key scripts cannot share a Lua state, but they share the `store`
module. The daemon writes values and calls `system.refresh()`; key scripts read
them in `passive()`.

```lua
-- daemon.lua (package)
local store  = require("store")
local system = require("system")
while true do
    store.set("my-pkg.value", compute())
    system.refresh()
    system.sleep(1000)
end

-- some_key.lua
local store  = require("store")
local script = {}
function script.passive(key, state)
    return { text = tostring(store.get("my-pkg.value") or "") }
end
return script
```

---

## 11. Quick Reference

### Which sleep to use?

| Context | Use | Why |
|---|---|---|
| `background()` | `system.sleep(ms)` | Yields the coroutine; passive/trigger run during the gap |
| `trigger()` | `time.sleep(ms)` | Blocking is fine; trigger runs once on demand |
| `_boot.lua` | `time.sleep(ms)` | Not a coroutine; system.sleep() does nothing here |
| `passive()` | **Neither** | Never sleep in passive  it stalls the whole paint loop |

### State patterns

```lua
-- Pattern: background fetches, passive displays
function script.background(state)
    while true do
        local out, _, code = shell.exec("uptime -p")
        if code == 0 then state.label = out:gsub("\n",""); system.refresh() end
        system.sleep(5000)
    end
end
function script.passive(key, state)
    return { text = state.label or "" }
end
```

### Forcing a repaint

Call `system.refresh()` any time you change `state` in `background()` or `trigger()`.
The runtime will run `passive()` again immediately and push the new appearance.

---

## 12. Examples

### Launcher button

```lua
local shell  = require("shell")
local system = require("system")

local script = {}

function script.passive(key, state)
    return {
        image = "icons/firefox.png",
        text  = "Firefox",
    }
end

function script.trigger(state)
    shell.open("https://firefox.com")
end

return script
```

---

### System status monitor

```lua
local shell  = require("shell")
local system = require("system")

local script = {}

function script.background(state)
    while true do
        local out, _, code = shell.exec("top -bn1 | grep 'Cpu(s)' | awk '{print $2}'")
        if code == 0 then
            state.cpu = out:gsub("\n","") .. "%"
            system.refresh()
        end
        system.sleep(2000)
    end
end

function script.passive(key, state)
    local cpu = state.cpu or ""
    local hot = tonumber(cpu) and tonumber(cpu) > 80
    return {
        color = hot and {255, 60, 0} or {0, 60, 255},
        text  = "CPU\n" .. cpu,
    }
end

return script
```

---

### Toggle button

```lua
local system = require("system")

local script = {}

function script.passive(key, state)
    if state.active then
        return { color = {0, 200, 80}, text = "ON" }
    else
        return { color = {60, 60, 60}, text = "OFF" }
    end
end

function script.trigger(state)
    state.active = not state.active
    system.refresh()
end

return script
```

---

### Counter with cross-script store

```lua
local store  = require("store")
local system = require("system")

local script = {}
local KEY = "global.click_count"

function script.passive(key, state)
    local n = store.get(KEY) or 0
    return {
        color = {80, 0, 160},
        text  = tostring(n),
    }
end

function script.trigger(state)
    local n = (store.get(KEY) or 0) + 1
    store.set(KEY, n)
    system.refresh()
end

return script
```

---

### Custom image from the web (cached)

```lua
local http    = require("http")
local file    = require("file")
local system  = require("system")
local json    = require("json")

local script  = {}
local CACHE   = CONFIG_DIR .. "/cache/weather_icon.png"

function script.background(state)
    while true do
        if not file.exists(CACHE) then
            local body, err = http.get("https://example.com/icon.png")
            if not err then
                file.write(CACHE, body)
                state.icon = CACHE
                system.refresh()
            end
        else
            state.icon = CACHE
        end
        system.sleep(3600000)  -- refresh once per hour
    end
end

function script.passive(key, state)
    if state.icon then
        return { image = state.icon }
    end
    return { color = {40, 40, 40}, text = "" }
end

return script
```
