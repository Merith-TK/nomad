-- uptime.lua - Shows system uptime

local shell = require("shell")
local time  = require("time")

local script = {}

function script.passive(key, state)
    local now = time.now()
    if not state.last_update or (now - state.last_update) >= 60 then
        state.last_update = now
        local out, _, code = shell.exec("uptime -p")
        if code == 0 then
            state.uptime_days  = tonumber(out:match("(%d+) day"))    or 0
            state.uptime_hours = tonumber(out:match("(%d+) hour"))   or 0
            state.uptime_mins  = tonumber(out:match("(%d+) minute")) or 0
        end
    end

    local d, h, m = state.uptime_days or 0, state.uptime_hours or 0, state.uptime_mins or 0
    local text = d > 0
        and string.format("UP\n%dd %dh", d, h)
        or  string.format("UP\n%dh %dm", h, m)

    return { color = {0, 100, 200}, text = text, text_color = {255, 255, 255} }
end

return script
