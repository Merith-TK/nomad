-- cpu.lua - Shows CPU usage percentage

local shell = require("shell")
local time  = require("time")
local system = require("system")

local script = {}

function script.passive(key, state)
    local now = time.now()
    if not state.last_update then
        -- First call: return default without running command
        state.last_update = now
        return { color = {0, 255, 0}, text = "CPU\n--%", text_color = {255, 255, 255} }
    elseif (now - state.last_update) >= 5 then
        state.last_update = now
        local out, _, code
        if system.os() == "windows" then
            -- CIM returns a plain integer e.g. "39"
            out, _, code = shell.exec("powershell -NoProfile -Command (Get-CimInstance Win32_Processor).LoadPercentage")
        else
            out, _, code = shell.exec("top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1}'")
        end
        if code == 0 then
            local cpu = tonumber(out:match("(%d+)"))
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
