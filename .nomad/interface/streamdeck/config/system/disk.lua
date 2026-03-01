-- disk.lua - Shows disk usage for C: drive
-- Demonstrates: system monitoring with passive updates

local system = require("system")
local shell = require("shell")

-- Passive: show current disk usage
function passive(key, state)
    -- Get disk usage via Windows WMIC
    local out, _, code = shell.exec("wmic logicaldisk where name=\"C:\" get FreeSpace,Size /Value")
    if code == 0 then
        local free = out:match("FreeSpace=(%d+)")
        local total = out:match("Size=(%d+)")

        if free and total then
            free = tonumber(free) / 1024 / 1024 / 1024 -- Convert to GB
            total = tonumber(total) / 1024 / 1024 / 1024
            local used = total - free
            local percent = (used / total) * 100
            state.disk_used = used
            state.disk_total = total
            state.disk_percent = percent
        end
    end

    local percent = state.disk_percent or 0
    local color = {0, 255, 0} -- Green for low usage

    if percent > 95 then
        color = {255, 0, 0} -- Red for high usage
    elseif percent > 85 then
        color = {255, 165, 0} -- Orange for medium usage
    end

    return {
        color = color,
        text = string.format("DISK\n%.0f%%", percent),
        text_color = {255, 255, 255}
    }
end