# Stream Deck Lua Scripting API

Scripts are placed in `.nomad/interface/streamdeck/config/` and organized into folders that become navigation pages.

## Script Structure

Each `.lua` script can define up to three functions:

```lua
-- Called every ~500ms by Go runtime
-- Must return quickly - no while loops or blocking calls!
-- Use state to track timing for slower polling
function background(state)
    state.count = (state.count or 0) + 1
    if state.count < 4 then return end  -- Only run every 4th call (~2 seconds)
    state.count = 0
    -- Update state.something
end

-- Called at 10fps when button is visible on screen
-- Return appearance table or nil
function passive(key, state)
    return {
        color = {255, 0, 0},      -- RGB background color
        text = "Hello",           -- Text overlay
        text_color = {255,255,255}, -- Text color (default: white)
        image = "icon.png"        -- Image path (relative to script) or URL
    }
end

-- Called when button is pressed
function trigger(state)
    -- Do something
end
```

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
Pause execution. **Warning: blocks all script functions during sleep.**
Avoid using in background() - use state-based timing instead.

```lua
system.sleep(100)  -- Sleep 100ms (use sparingly)
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

-- Background: called every 500ms, use state for slower polling
function background(state)
    state.poll = (state.poll or 0) + 1
    if state.poll < 4 then return end  -- Every ~2 seconds
    state.poll = 0
    
    if system.os() == "windows" then
        local out = shell.exec('wmic cpu get loadpercentage /value')
        local load = out:match("LoadPercentage=(%d+)")
        state.cpu = tonumber(load) or 0
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

1. **Background must return quickly** - Go calls it every 500ms, no while loops!
2. **Use state-based timing** - track call counts in state for slower polling
3. **Use `shell.terminal()` for interactive apps** - `shell.exec()` blocks
4. **State persists** - use it to communicate between functions
5. **Passive is called frequently** - keep it fast, no I/O
6. **Trigger can block briefly** - but keep it reasonable
7. **Restart policy** - controls what happens if background crashes
