-- shutdown.lua - System shutdown with confirmation animation
-- Demonstrates: trigger-only script (no background/passive needed)
--
-- WARNING: Actual shutdown command commented out for safety

local streamdeck = require("streamdeck")
local system = require("system")

-- No background worker needed for this script
-- No passive needed - uses default appearance

-- Trigger: flash red warning, then shutdown
function trigger(state)
    -- Count presses - require double-press to confirm
    state.presses = (state.presses or 0) + 1
    
    if state.presses == 1 then
        -- First press: flash warning
        print("Shutdown: press again within 3 seconds to confirm")
        
        -- Flash keys red 2 times as warning
        for i = 1, 2 do
            local keys = streamdeck.get_keys()
            for k = 0, keys - 1 do
                streamdeck.set_color(k, 255, 0, 0)
            end
            system.sleep(150)
            streamdeck.clear()
            system.sleep(150)
        end
        
        -- Reset counter after 3 seconds
        -- (In a real impl, you'd use a timer. Here we just note the limitation.)
        -- For now, state persists so next press will trigger.
        
    else
        -- Second press: confirmed
        print("Shutdown confirmed!")
        state.presses = 0
        
        -- Flash green to confirm
        local keys = streamdeck.get_keys()
        for k = 0, keys - 1 do
            streamdeck.set_color(k, 0, 255, 0)
        end
        system.sleep(500)
        streamdeck.clear()
        
        -- Actual shutdown - UNCOMMENT TO ENABLE
        -- local shell = require("shell")
        -- shell.exec("shutdown /s /t 60 /c \"Shutdown initiated from Stream Deck\"")
    end
    
    system.refresh()
end
