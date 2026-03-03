-- uptime.lua - Shows system uptime

local shell = require("shell")
local time  = require("time")
local system = require("system")

local script = {}

function script.passive(key, state)
    local now = time.now()
    if not state.last_update then
        state.last_update = now
        return { color = {0, 100, 200}, text = "UP\n--h --m", text_color = {255, 255, 255} }
    elseif (now - state.last_update) >= 60 then
        state.last_update = now
        if system.os() == "windows" then
            -- PowerShell outputs total seconds as a plain number
            local out, _, code = shell.exec("powershell -NoProfile -Command \"((Get-Date) - (gcim Win32_OperatingSystem).LastBootUpTime).TotalSeconds\"")
            if code == 0 then
                local secs = tonumber(out:match("([%d%.]+)")) or 0
                state.uptime_days  = math.floor(secs / 86400)
                state.uptime_hours = math.floor((secs % 86400) / 3600)
                state.uptime_mins  = math.floor((secs % 3600) / 60)
            end
        else
            local out, _, code = shell.exec("uptime -p")
            if code == 0 then
                state.uptime_days  = tonumber(out:match("(%d+) day"))    or 0
                state.uptime_hours = tonumber(out:match("(%d+) hour"))   or 0
                state.uptime_mins  = tonumber(out:match("(%d+) minute")) or 0
            end
        end
    end

    local d, h, m = state.uptime_days or 0, state.uptime_hours or 0, state.uptime_mins or 0
    local text = d > 0
        and string.format("UP\n%dd %dh", d, h)
        or  string.format("UP\n%dh %dm", h, m)

    return { color = {0, 100, 200}, text = text, text_color = {255, 255, 255} }
end

return script
