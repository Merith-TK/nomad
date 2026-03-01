-- uptime.lua - Shows system uptime
-- Demonstrates: system information display

local system = require("system")
local shell = require("shell")

-- Passive: show system uptime
function passive(key, state)
    -- Get uptime via Windows systeminfo
    local out, _, code = shell.exec("powershell -Command \"(Get-Date) - (Get-CimInstance Win32_OperatingSystem).LastBootUpTime | Format-Table -HideTableHeaders\"")
    if code == 0 then
        -- Parse the output (format: dd.hh:mm:ss)
        local days, hours, mins, secs = out:match("(%d+)%s+(%d+):(%d+):(%d+)")
        if days and hours and mins then
            state.uptime_days = tonumber(days)
            state.uptime_hours = tonumber(hours)
            state.uptime_mins = tonumber(mins)
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