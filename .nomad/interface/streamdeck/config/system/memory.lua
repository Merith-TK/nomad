-- memory.lua - Shows memory usage percentage
-- Demonstrates: system monitoring with passive updates

local system = require("system")
local shell = require("shell")

-- Passive: show current memory usage
function passive(key, state)
    -- Get memory usage via Windows WMIC
    local out, _, code = shell.exec("wmic OS get FreePhysicalMemory,TotalVisibleMemorySize /Value")
    if code == 0 then
        local free = out:match("FreePhysicalMemory=(%d+)")
        local total = out:match("TotalVisibleMemorySize=(%d+)")

        if free and total then
            free = tonumber(free) / 1024 / 1024 -- Convert to GB
            total = tonumber(total) / 1024 / 1024
            local used = total - free
            local percent = (used / total) * 100
            state.memory_used = used
            state.memory_total = total
            state.memory_percent = percent
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