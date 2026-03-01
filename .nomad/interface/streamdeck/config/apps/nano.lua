-- nano.lua - Opens nano on Windows
-- Demonstrates: background polling, passive display, trigger action
--
-- Available modules: shell, http, system, streamdeck
-- Available globals: state (shared table), SCRIPT_NAME, SCRIPT_PATH, CONFIG_DIR

local shell = require("shell")
local system = require("system")

-- Restart policy: "always" (default), "never", or "once"
RESTART_POLICY = "always"

-- Background worker: polls every 2 seconds to check if nano is running
function background(state)
    while true do
        if system.os() == "windows" then
            local out, _, code = shell.exec("tasklist /FI \"IMAGENAME eq nano.exe\" /NH 2>nul")
            state.running = (code == 0 and out:find("nano.exe") ~= nil)
        else
            state.running = false
        end
        system.sleep(2000)
    end
end

-- Passive: customize icon appearance based on state
-- Called at ~15fps when this script's button is visible
function passive(key, state)
    if state.running then
        -- Green background when nano is running
        return {
            color = {50, 180, 50},
            text = "NP*",
            text_color = {255, 255, 255}
        }
    else
        -- Gray background when not running
        return {
            color = {80, 80, 80},
            text = "NP",
            text_color = {200, 200, 200}
        }
    end
end

-- Trigger: called when button is pressed
function trigger(state)
    if system.os() ~= "windows" then
        print("This script only works on Windows")
        return
    end

    shell.open("nano.exe")
    -- After opening, force a refresh to update state faster
    system.refresh()
end