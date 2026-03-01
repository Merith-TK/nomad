-- uptime.lua - Shows system uptime
-- Demonstrates: system information display

local system = require("system")
local shell = require("shell")

-- Passive: show system uptime
function passive(key, state)
    -- Only update every 60 seconds (uptime doesn't change often)
    local now = os.time()
    if not state.last_update or (now - state.last_update) >= 60 then
        state.last_update = now

        -- Get uptime via uptime command (Linux)
        local out, _, code = shell.exec("uptime -p")
        if code == 0 then
            -- Parse output like "up 1 day, 2 hours, 30 minutes"
            local days = out:match("(%d+) day")
            local hours = out:match("(%d+) hour")
            local mins = out:match("(%d+) minute")

            state.uptime_days = days and tonumber(days) or 0
            state.uptime_hours = hours and tonumber(hours) or 0
            state.uptime_mins = mins and tonumber(mins) or 0
        end
    end

    local days = state.uptime_days or 0
    local hours = state.uptime_hours or 0

    local display_text = ""
    if days > 0 then
        display_text = string.format("UP\n%dd %dh", days, hours)
    else
        display_text = string.format("UP\n%dh %dm", hours, state.uptime_mins or 0)
    end

    return {
        color = {0, 100, 200}, -- Blue
        text = display_text,
        text_color = {255, 255, 255}
    }
end