-- disk.lua - Shows disk usage for root filesystem
-- Demonstrates: system monitoring with passive updates

local system = require("system")
local shell = require("shell")

-- Passive: show current disk usage
function passive(key, state)
    -- Only update every 30 seconds to reduce load (disk I/O is expensive)
    local now = os.time()
    if not state.last_update or (now - state.last_update) >= 30 then
        state.last_update = now

        -- Get disk usage via df command (Linux)
        local out, _, code = shell.exec("df / | tail -1 | awk '{print $5}' | sed 's/%//'")
        if code == 0 then
            local percent = tonumber(out:match("([%d]+)"))
            if percent then
                state.disk_percent = percent
            end
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