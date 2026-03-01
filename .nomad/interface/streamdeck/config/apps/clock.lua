-- clock.lua - Displays current time on the button
-- Demonstrates: passive-only script with animated display

local system = require("system")

-- No background needed - passive updates at 15fps
-- No trigger needed - this is display-only

-- Passive: show current time
function passive(key, state)
    -- Update state every second
    local current = os.time()
    if state.last_time ~= current then
        state.last_time = current
        local t = os.date("*t", current)
        state.hour = t.hour
        state.min = t.min
        state.sec = t.sec
    end
    
    -- Format time
    local time_str = string.format("%02d:%02d", state.hour or 0, state.min or 0)
    
    -- Blink colon every second
    if (state.sec or 0) % 2 == 0 then
        time_str = string.format("%02d %02d", state.hour or 0, state.min or 0)
    end
    
    return {
        color = {20, 20, 60},
        text = time_str,
        text_color = {100, 200, 255}
    }
end