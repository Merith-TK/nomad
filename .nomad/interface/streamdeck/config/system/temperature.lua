-- temperature.lua - Shows CPU temperature (if available)
-- Demonstrates: hardware monitoring

local system = require("system")
local shell = require("shell")

-- Passive: show CPU temperature
function passive(key, state)
    -- Only update every 10 seconds
    local now = os.time()
    if not state.last_update or (now - state.last_update) >= 10 then
        state.last_update = now

        -- Try to get temperature from thermal zones (Linux/Raspberry Pi)
        local temp = nil

        -- Try /sys/class/thermal/thermal_zone0/temp (common on Linux)
        local out, _, code = shell.exec("cat /sys/class/thermal/thermal_zone0/temp 2>/dev/null || echo '0'")
        if code == 0 then
            local raw_temp = tonumber(out:match("([%d]+)"))
            if raw_temp and raw_temp > 0 then
                temp = raw_temp / 1000 -- Convert from millidegrees to degrees
            end
        end

        -- If thermal zone failed, try vcgencmd (Raspberry Pi specific)
        if not temp then
            out, _, code = shell.exec("vcgencmd measure_temp 2>/dev/null | sed 's/temp=//' | sed 's/\\'C//'")
            if code == 0 then
                temp = tonumber(out:match("([%d%.]+)"))
            end
        end

        if temp then
            state.temperature = temp
        end
    end

    local temp_val = state.temperature or 0
    local color = {0, 255, 0} -- Green for normal temp

    if temp_val > 70 then
        color = {255, 0, 0} -- Red for hot
    elseif temp_val > 55 then
        color = {255, 165, 0} -- Orange for warm
    end

    local display_temp = temp_val > 0 and string.format("%.0f", temp_val) or "--"
    return {
        color = color,
        text = string.format("TEMP\n%s°C", display_temp),
        text_color = {255, 255, 255}
    }
end