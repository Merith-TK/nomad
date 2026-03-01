-- temperature.lua - Shows CPU temperature (if available)

local shell = require("shell")
local time  = require("time")

local script = {}

function script.passive(key, state)
    local now = time.now()
    if not state.last_update or (now - state.last_update) >= 10 then
        state.last_update = now

        -- Try /sys/class/thermal (Linux / Raspberry Pi)
        local out, _, code = shell.exec("cat /sys/class/thermal/thermal_zone0/temp 2>/dev/null || echo '0'")
        if code == 0 then
            local raw = tonumber(out:match("([%d]+)"))
            if raw and raw > 0 then
                state.temperature = raw / 1000
            end
        end

        -- Fallback: vcgencmd (Raspberry Pi)
        if not state.temperature then
            out, _, code = shell.exec("vcgencmd measure_temp 2>/dev/null | sed 's/temp=//' | sed \"s/'C//\"")
            if code == 0 then
                state.temperature = tonumber(out:match("([%d%.]+)"))
            end
        end
    end

    local temp  = state.temperature or 0
    local color = {0, 255, 0}
    if temp > 70 then
        color = {255, 0, 0}
    elseif temp > 55 then
        color = {255, 165, 0}
    end

    local display = temp > 0 and string.format("%.0f", temp) or "--"
    return { color = color, text = string.format("TEMP\n%s°C", display), text_color = {255, 255, 255} }
end

return script
