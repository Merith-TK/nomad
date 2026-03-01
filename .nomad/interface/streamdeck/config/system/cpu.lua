-- cpu.lua - Shows CPU usage percentage

local shell = require("shell")
local time  = require("time")

local script = {}

function script.passive(key, state)
    local now = time.now()
    if not state.last_update or (now - state.last_update) >= 5 then
        state.last_update = now
        local out, _, code = shell.exec("top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1}'")
        if code == 0 then
            local cpu = tonumber(out:match("([%d%.]+)"))
            if cpu then state.cpu = cpu end
        end
    end

    local cpu   = state.cpu or 0
    local color = {0, 255, 0}
    if cpu > 80 then
        color = {255, 0, 0}
    elseif cpu > 60 then
        color = {255, 165, 0}
    end

    return { color = color, text = string.format("CPU\n%.0f%%", cpu), text_color = {255, 255, 255} }
end

return script
