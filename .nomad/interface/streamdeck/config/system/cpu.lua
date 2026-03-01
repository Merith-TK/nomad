-- cpu.lua - Shows CPU usage percentage
-- Demonstrates: system monitoring with passive updates

local system = require("system")
local shell = require("shell")

-- Passive: show current CPU usage
function passive(key, state)
    -- Only update every 5 seconds to reduce load
    local now = os.time()
    if not state.last_update or (now - state.last_update) >= 5 then
        state.last_update = now

        -- Get CPU usage via top command (Linux)
        local out, _, code = shell.exec("top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1}'")
        if code == 0 then
            local cpu = tonumber(out:match("([%d%.]+)"))
            if cpu then
                state.cpu = cpu
            end
        end
    end

    local cpu = state.cpu or 0
    local color = {0, 255, 0} -- Green for low usage

    if cpu > 80 then
        color = {255, 0, 0} -- Red for high usage
    elseif cpu > 60 then
        color = {255, 165, 0} -- Orange for medium usage
    end

    return {
        color = color,
        text = string.format("CPU\n%.0f%%", cpu),
        text_color = {255, 255, 255}
    }
end