# NOMAD Stream Deck — Lua Scripting API

Scripts live under `.nomad/interface/streamdeck/config/` and are organised into
subdirectories that become navigation folders on the device.

---

## Script Structure

Every script **must** return a table (conventionally named `script`). The table
can contain any combination of three lifecycle functions:

```lua
local script = {}

--[[
  background(state)
  Runs as a coroutine, restarted according to RESTART_POLICY.
  Use system.sleep(ms) to yield back to Go — passive() and trigger()
  execute during the sleep window.
  Do NOT call system.sleep() from passive() or trigger(); use time.sleep() there.
]]
function script.background(state)
    while true do
        -- update state.something
        system.sleep(1000)  -- yield for 1 second
    end
end

--[[
  passive(key, state) -> table|nil
  Called at the passive FPS rate (default 2 fps) while the key is on-screen.
  Return an appearance table to update the key display, or nil to leave it unchanged.
  key   : zero-based key index (number)
  state : shared per-script state table
]]
function script.passive(key, state)
    return {
        color      = {255, 0, 0},       -- RGB background  (0-255 each)
        text       = "Hi",              -- label text (newlines allowed)
        text_color = {255, 255, 255},   -- RGB text colour (default: white)
        image      = "icon.png",        -- image path (relative) or https:// URL
    }
end

--[[
  trigger(state)
  Called once when the key is pressed.
  Avoid long blocking operations; use shell.exec_async() or background state flags.
]]
function script.trigger(state)
    -- do something
end

return script
```

> All three functions are optional — only define what your script needs.

---

## Shared State

The `state` table is created once per script and passed to every call.
Use it to share data across `background`, `passive`, and `trigger`.

```lua
function script.background(state)
    state.count = (state.count or 0) + 1
    system.sleep(1000)
end

function script.passive(key, state)
    return { text = tostring(state.count or 0), color = {30, 30, 80} }
end
```

---

## Global Variables

Set automatically for every script at load time:

| Variable | Type | Description |
|---|---|---|
| `SCRIPT_PATH` | string | Absolute path to the current `.lua` file |
| `SCRIPT_NAME` | string | Filename without the `.lua` extension |
| `CONFIG_DIR` | string | Absolute path to the config root directory |
| `state` | table | Alias for the shared state table |

---

## Restart Policy

Set `RESTART_POLICY` as a **top-level global** (not inside a function) to
control what happens when `background()` exits or crashes:

```lua
RESTART_POLICY = "always"  -- default: restart immediately on exit or error
RESTART_POLICY = "once"    -- restart once, then stop permanently
RESTART_POLICY = "never"   -- never restart; background runs exactly once
```

---

## Special Files

### `_boot.lua`

If present in the config root, runs once at startup before any other script
is loaded. Use for splash/boot animations.

The returned table must contain a `boot()` function.
Use `time.sleep()` (blocking) here — **not** `system.sleep()` (which requires a coroutine context).

```lua
local streamdeck = require("streamdeck")
local time = require("time")

local script = {}

function script.boot()
    streamdeck.clear()
    for i = 0, streamdeck.get_keys() - 1 do
        streamdeck.set_color(i, 0, 0, 200)
        time.sleep(50)
    end
    streamdeck.clear()
end

return script
```

---

## Modules

All modules are loaded with `require()` and use a zero-cost preload (no disk I/O).

---

### `shell` — Command Execution

```lua
local shell = require("shell")
```

| Function | Returns | Description |
|---|---|---|
| `shell.exec(cmd)` | `stdout, stderr, exitcode` | Run and **wait** for completion |
| `shell.exec_async(cmd)` | `ok, err` | Start in background, don't wait |
| `shell.open(target)` | — | Open file / URL with system default app |
| `shell.terminal(cmd)` | — | Open a new terminal window running `cmd` |

```lua
-- Blocking execution
local out, err, code = shell.exec("echo hello")

-- Background (fire-and-forget)
shell.exec_async("start /B myapp.exe")

-- Open URL or file
shell.open("https://github.com")

-- New terminal window
shell.terminal("ssh user@myserver")
```

---

### `system` — OS Utilities

```lua
local system = require("system")
```

| Function | Returns | Description |
|---|---|---|
| `system.os()` | string | `"windows"`, `"darwin"`, or `"linux"` |
| `system.env(name)` | string | Value of environment variable `name` |
| `system.hostname()` | string | Machine hostname |
| `system.sleep(ms)` | — | **Yield** the background coroutine for `ms` milliseconds. Only valid inside `background()`. |
| `system.refresh()` | — | Request an immediate passive redraw for this script |

> **Important:** `system.sleep()` yields the Lua coroutine. Calling it outside
> `background()` (i.e. from `passive()`, `trigger()`, or `_boot.lua`) will
> panic. Use `time.sleep()` for blocking sleeps in those contexts.

```lua
function script.background(state)
    while true do
        state.running = (system.os() == "windows")
        system.sleep(2000)   -- yield; passive/trigger run during this gap
    end
end

function script.trigger(state)
    state.active = true
    system.refresh()         -- update display right away
end
```

---

### `http` — HTTP Requests

```lua
local http = require("http")
```

| Function | Returns | Description |
|---|---|---|
| `http.get(url)` | `body, err` | HTTP GET |
| `http.post(url, body, content_type)` | `body, err` | HTTP POST |
| `http.request(method, url, body, headers)` | `body, err` | Custom request |

```lua
local body, err = http.get("https://api.example.com/status")
if err then print("HTTP error: " .. err) end

local res, err = http.post(
    "https://api.example.com/event",
    '{"key":"value"}',
    "application/json"
)

local res, err = http.request("PUT", url, body, {
    ["Authorization"] = "Bearer token",
    ["Content-Type"]  = "application/json",
})
```

---

### `streamdeck` — Hardware Control

```lua
local deck = require("streamdeck")
```

| Function | Description |
|---|---|
| `deck.set_color(key, r, g, b)` | Set one key to a solid RGB colour |
| `deck.set_brightness(pct)` | Set display brightness 0–100 |
| `deck.clear()` | Set all keys to black |
| `deck.clear_key(key)` | Set one key to black |
| `deck.reset()` | Full device reset |
| `deck.get_model()` | Returns model name string |
| `deck.get_keys()` | Total key count |
| `deck.get_layout()` | Returns `cols, rows` |

```lua
-- Flash the pressed key red
function script.trigger(state)
    deck.set_color(0, 255, 0, 0)
end

-- Sweep a colour across all keys in boot
local cols, rows = deck.get_layout()
for col = 0, cols - 1 do
    for row = 0, rows - 1 do
        deck.set_color(row * cols + col, 0, 100, 255)
    end
end
```

---

### `file` — File I/O

All paths are restricted to the config directory.

```lua
local file = require("file")
```

| Function | Returns | Description |
|---|---|---|
| `file.read(path)` | `content, err` | Read file as string |
| `file.write(path, content)` | `ok, err` | Write string to file |
| `file.exists(path)` | bool | True if path exists |
| `file.list(dir)` | table of names | List directory contents |
| `file.isdir(path)` | bool | True if path is a directory |
| `file.size(path)` | number, err | File size in bytes |

```lua
local content, err = file.read("data.txt")
file.write("out.txt", "hello")
for _, name in ipairs(file.list(".")) do print(name) end
```

---

## Standard Library (lualib)

Pure-Go implementations — zero disk I/O on `require()`.

---

### `time` — Time & Date

```lua
local time = require("time")
```

| Function | Returns | Description |
|---|---|---|
| `time.now()` | number | Current Unix timestamp (seconds) |
| `time.timestamp()` | number | Alias for `time.now()` |
| `time.date([ts])` | table | Decompose timestamp into fields |
| `time.format(ts, layout)` | string | Format timestamp with Go layout |
| `time.parse(layout, str)` | number, err | Parse string to timestamp |
| `time.sleep(ms)` | — | **Blocking** sleep (safe anywhere, unlike `system.sleep`) |

```lua
local now = time.now()
local d   = time.date(now)
-- d.year, d.month, d.day, d.hour, d.minute, d.second, d.weekday, d.yearday

local text = string.format("%02d:%02d", d.hour, d.minute)

-- Throttle updates
if not state.last or (now - state.last) >= 5 then
    state.last = now
    -- expensive work here
end

-- Blocking sleep (inside passive, trigger, or _boot.lua)
time.sleep(200)
```

---

### `json` — JSON Encode / Decode

```lua
local json = require("json")
```

| Function | Returns | Description |
|---|---|---|
| `json.encode(value)` | string, err | Lua value → JSON string |
| `json.decode(str)` | value, err | JSON string → Lua value |

```lua
local str, err = json.encode({ name = "alice", score = 42 })
local tbl, err = json.decode('{"ok":true}')
if tbl.ok then print("all good") end
```

---

### `log` — Levelled Logging

```lua
local log = require("log")
```

| Function | Description |
|---|---|
| `log.info(msg)` | Info-level message |
| `log.warn(msg)` | Warning |
| `log.error(msg)` | Error |
| `log.debug(msg)` | Debug (verbose) |
| `log.printf(fmt, ...)` | Printf-style with Go format verbs |
| `log.print(...)` | Space-separated values |

---

### `utils` — Table Utilities

```lua
local utils = require("utils")
```

| Function | Returns | Description |
|---|---|---|
| `utils.deepcopy(t)` | table | Recursive copy of table `t` |
| `utils.contains(t, v)` | bool | True if value `v` is in `t` |
| `utils.size(t)` | number | Count of all entries (including non-sequence keys) |
| `utils.merge(t1, t2)` | table | Deep copy of `t1` with `t2` merged in (`t2` wins) |

---

### `strings` — String Utilities

```lua
local strings = require("strings")
```

| Function | Returns | Description |
|---|---|---|
| `strings.split(str, sep)` | table | Split by separator |
| `strings.trim(str)` | string | Strip leading/trailing whitespace |
| `strings.startswith(str, prefix)` | bool | Prefix check |
| `strings.endswith(str, suffix)` | bool | Suffix check |
| `strings.contains(str, sub)` | bool | Substring check |
| `strings.replace(str, old, new[, n])` | string | Replace (`n=-1` = all) |
| `strings.upper(str)` | string | Uppercase |
| `strings.lower(str)` | string | Lowercase |
| `strings.capitalize(str)` | string | First letter uppercased |
| `strings.titlecase(str)` | string | Title-case each word |

---

## Complete Examples

### Simple Launcher

```lua
-- notepad.lua
local shell = require("shell")

local script = {}

function script.passive(key, state)
    return { color = {60, 60, 100}, text = "NOTE", text_color = {255, 255, 255} }
end

function script.trigger(state)
    shell.open("notepad.exe")
end

return script
```

### Status Monitor

```lua
-- cpu.lua
local shell  = require("shell")
local system = require("system")
local time   = require("time")

local script = {}

function script.background(state)
    while true do
        local out, _, code = shell.exec(
            "top -bn1 | grep 'Cpu(s)' | awk '{print 100-$8}'"
        )
        if code == 0 then
            state.cpu = tonumber(out:match("([%d%.]+)")) or 0
        end
        system.sleep(2000)
    end
end

function script.passive(key, state)
    local cpu = state.cpu or 0
    local r   = math.floor(cpu * 2.55)
    local g   = math.floor(255 - cpu * 2.55)
    return { color = {r, g, 0}, text = string.format("%.0f%%", cpu), text_color = {255,255,255} }
end

return script
```

### Toggle with Confirmation

```lua
-- mute.lua
local system = require("system")
local shell  = require("shell")

local script = {}

function script.passive(key, state)
    if state.muted then
        return { color = {180, 50, 50}, text = "MUTE" }
    else
        return { color = {50, 100, 50}, text = "LIVE" }
    end
end

function script.trigger(state)
    state.muted = not state.muted
    -- toggle system mute on Linux:
    shell.exec("pactl set-sink-mute @DEFAULT_SINK@ toggle")
    system.refresh()
end

return script
```

### Custom Image

```lua
-- app.lua  (place icon.png next to this file)
local shell = require("shell")

local script = {}

function script.passive(key, state)
    return { image = "icon.png" }   -- relative to script location
end

function script.trigger(state)
    shell.open("myapp.exe")
end

return script
```

---

## Quick-Reference Cheatsheet

```
Context       | Blocking sleep  | Notes
--------------|-----------------|----------------------------------
background()  | system.sleep(ms)| yields coroutine; safe
passive()     | time.sleep(ms)  | blocking; keep < a few ms
trigger()     | time.sleep(ms)  | blocking; keep short
_boot.lua     | time.sleep(ms)  | no coroutine context
```

- **State is per-script** — two buttons running the same `.lua` file share nothing.
- **passive() is called at the passive FPS** (default 2 fps). Keep it cheap — no I/O, no heavy computation.
- **trigger() blocks the event loop** while running. For long tasks, set a flag in state and handle it in background().
- **`RESTART_POLICY`** must be a top-level global (not inside any function).


## Shared State

The `state` table persists across all function calls for a script. Use it to share data between background, passive, and trigger.

```lua
function background(state)
    state.count = (state.count or 0) + 1
end

function passive(key, state)
    return { text = tostring(state.count or 0), color = {50,50,50} }
end
```

## Dual Architecture Support

Scripts support two architectures for maximum flexibility:

### Legacy Mode (Global Functions)
Define functions globally as shown in the examples above. This mode is fully backward compatible.

### Module Mode (Return Table)
Return a table containing your functions for better organization:

```lua
local utils = require("utils")

local function myTrigger(state)
    -- Do something
end

local function myPassive(key, state)
    return {
        color = {0, 255, 0},
        text = "Ready",
        text_color = {0, 0, 0}
    }
end

local function myBackground(state)
    while true do
        -- Background work
        system.sleep(1000)
    end
end

-- Return module table
return {
    trigger = myTrigger,
    passive = myPassive,
    background = myBackground
}
```

Both modes are automatically detected. Module mode is recommended for new scripts.

## Standard Library (lualib)

NOMAD provides built-in stdlib modules implemented in Go — no disk I/O on `require()`:

### `utils` - Table utilities
```lua
local utils = require("utils")

utils.deepcopy(t)           -- deep copy of table t
utils.contains(t, value)    -- true if value exists in t
utils.size(t)               -- count of all entries (including non-sequence)
utils.merge(t1, t2)         -- deep copy of t1 with t2 keys merged in (t2 wins)
```

### `strings` - String utilities
```lua
local strings = require("strings")

strings.split(str, sep)          -- split str by sep, returns table
strings.trim(str)                -- strip leading/trailing whitespace
strings.startswith(str, prefix)  -- true if str begins with prefix
strings.endswith(str, suffix)    -- true if str ends with suffix
strings.contains(str, substr)    -- true if substr found in str
strings.replace(str, old, new [, n]) -- replace n occurrences (-1 = all)
strings.upper(str)               -- uppercase
strings.lower(str)               -- lowercase
strings.capitalize(str)          -- first letter uppercased
strings.titlecase(str)           -- title case each word
```

## Global Variables

| Variable | Description |
|----------|-------------|
| `SCRIPT_PATH` | Absolute path to the current script |
| `SCRIPT_NAME` | Script filename without `.lua` extension |
| `CONFIG_DIR` | Absolute path to the config directory |

## Restart Policy

Set `RESTART_POLICY` global to control background worker restart behavior:

```lua
RESTART_POLICY = "always"  -- Always restart on error (default)
RESTART_POLICY = "never"   -- Never restart, fail permanently  
RESTART_POLICY = "once"    -- Restart once, then stop
```

---

## Modules

### `shell` - Command Execution

```lua
local shell = require("shell")
```

#### `shell.exec(command)`
Execute command and wait for result. **Blocks until complete.**

```lua
local stdout, stderr, exitcode = shell.exec("echo hello")
-- stdout: "hello\n"
-- stderr: ""
-- exitcode: 0
```

#### `shell.exec_async(command)`
Start command in background, don't wait for result.

```lua
local ok, err = shell.exec_async("start notepad")
-- ok: true/false
-- err: error message or nil
```

#### `shell.open(target)`
Open file/URL with system default application.

```lua
shell.open("https://github.com")  -- Opens in browser
shell.open("C:\\file.txt")        -- Opens in default text editor
shell.open("explorer.exe")        -- Opens Windows Explorer
```

#### `shell.terminal(command)`
Open a **new terminal window** and run command. Use for interactive CLI apps.

```lua
shell.terminal("nano")           -- Opens cmd with nano
shell.terminal("python")         -- Opens cmd with Python REPL
shell.terminal("ssh user@host")  -- Opens SSH session
```

---

### `system` - System Utilities

```lua
local system = require("system")
```

#### `system.os()`
Returns OS name: `"windows"`, `"darwin"`, or `"linux"`.

```lua
if system.os() == "windows" then
    -- Windows-specific code
end
```

#### `system.sleep(milliseconds)`
Pause execution. **In background(), this yields to Go** - passive/trigger can run during sleep.
In trigger/passive, sleep is capped at 100ms to avoid blocking.

```lua
-- In background(), sleep yields and resumes after duration
function background(state)
    while true do
        -- do work
        system.sleep(2000)  -- Yields for 2 seconds
    end
end
```

#### `system.refresh()`
Request immediate display refresh. Useful after changing state in trigger.

```lua
function trigger(state)
    state.active = true
    system.refresh()  -- Update display immediately
end
```

#### `system.env(name)`
Get environment variable.

```lua
local home = system.env("USERPROFILE")  -- Windows
local home = system.env("HOME")         -- Linux/macOS
```

#### `system.hostname()`
Get system hostname.

```lua
local name = system.hostname()
```

---

### `http` - HTTP Requests

```lua
local http = require("http")
```

#### `http.get(url)`
Perform HTTP GET request.

```lua
local body, err = http.get("https://api.example.com/data")
if body then
    print(body)
else
    print("Error: " .. err)
end
```

#### `http.post(url, body, content_type)`
Perform HTTP POST request.

```lua
local response, err = http.post(
    "https://api.example.com/data",
    '{"key": "value"}',
    "application/json"
)
```

#### `http.request(method, url, body, headers)`
Perform custom HTTP request.

```lua
local response, err = http.request(
    "PUT",
    "https://api.example.com/resource",
    '{"updated": true}',
    {
        ["Content-Type"] = "application/json",
        ["Authorization"] = "Bearer token123"
    }
)
```

---

### `streamdeck` - Device Control

```lua
local deck = require("streamdeck")
```

#### `deck.brightness(level)`
Set display brightness (0-100).

```lua
deck.brightness(75)
```

#### `deck.key_count()`
Get total number of keys on device.

```lua
local keys = deck.key_count()  -- e.g., 15 for MK.2
```

#### `deck.cols()` / `deck.rows()`
Get device dimensions.

```lua
local cols = deck.cols()  -- e.g., 5
local rows = deck.rows()  -- e.g., 3
```

---

## Special Files

### `_boot.lua`
If present in config root, runs once at startup before other scripts load. Use for boot animations.

```lua
-- _boot.lua
local deck = require("streamdeck")

function boot()
    -- Called once at startup
    for i = 0, deck.key_count() - 1 do
        deck.set_color(i, math.random(255), math.random(255), math.random(255))
        system.sleep(50)
    end
end
```

---

## Example Scripts

### Simple Launcher
```lua
-- notepad.lua
local shell = require("shell")

function passive(key, state)
    return {
        color = {60, 60, 100},
        text = "Note",
        text_color = {255, 255, 255}
    }
end

function trigger(state)
    shell.open("notepad.exe")
end
```

### Status Monitor
```lua
-- cpu.lua
local shell = require("shell")
local system = require("system")

-- Background: coroutine with sleep
function background(state)
    while true do
        if system.os() == "windows" then
            local out = shell.exec('wmic cpu get loadpercentage /value')
            local load = out:match("LoadPercentage=(%d+)")
            state.cpu = tonumber(load) or 0
        end
        system.sleep(1000)  -- Update every second
    end
end

function passive(key, state)
    local cpu = state.cpu or 0
    local r = math.floor(cpu * 2.55)
    local g = math.floor(255 - cpu * 2.55)
    
    return {
        color = {r, g, 0},
        text = cpu .. "%",
        text_color = {255, 255, 255}
    }
end
```

### Toggle with State
```lua
-- mute.lua
local shell = require("shell")

function passive(key, state)
    if state.muted then
        return { color = {180, 50, 50}, text = "MUTE" }
    else
        return { color = {50, 100, 50}, text = "🔊" }
    end
end

function trigger(state)
    state.muted = not state.muted
    -- Actually toggle system mute here
    shell.exec('nircmd.exe mutesysvolume 2')
end
```

### Custom Icon
```lua
-- app.lua
-- Place icon.png next to app.lua in the same folder

function passive(key, state)
    return {
        image = "icon.png"  -- Relative to script location
    }
end

function trigger(state)
    shell.open("myapp.exe")
end
```

### Remote Icon
```lua
-- weather.lua

function passive(key, state)
    -- Fetch icon from URL (cached automatically)
    return {
        image = "https://example.com/weather-icon.png"
    }
end
```

---

## Tips

1. **Use `while true` + `system.sleep()` in background** - sleep yields to Go, doesn't block
2. **Use `shell.terminal()` for interactive apps** - `shell.exec()` blocks
3. **State persists** - use it to communicate between functions
4. **Passive is called frequently** - keep it fast, no I/O
5. **Trigger can block briefly** - but keep it reasonable
6. **Restart policy** - controls what happens if background crashes
