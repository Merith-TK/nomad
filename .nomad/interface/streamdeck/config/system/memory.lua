-- memory.lua - Shows memory usage percentage
-- Demonstrates: system monitoring with passive updates

local system = require("system")
local shell = require("shell")

-- Passive: show current memory usage
function passive(key, state)
    -- Only update every 5 seconds to reduce load
    local now = os.time()
    if not state.last_update or (now - state.last_update) >= 5 then
        state.last_update = now

        -- Get memory usage via free command (Linux)
        local out, _, code = shell.exec("free | grep Mem | awk '{printf \"%.0f\", $3/$2 * 100.0}'")
        if code == 0 then
            local percent = tonumber(out:match("([%d]+)"))
            if percent then
                state.memory_percent = percent
            end
        end
    end

    local percent = state.memory_percent or 0
    local color = {0, 255, 0} -- Green for low usage

    if percent > 90 then
        color = {255, 0, 0} -- Red for high usage
    elseif percent > 75 then
        color = {255, 165, 0} -- Orange for medium usage
    end

    return {
        color = color,
        text = string.format("MEM\n%.0f%%", percent),
        text_color = {255, 255, 255}
    }
end