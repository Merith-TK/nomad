-- shutdown.lua - System shutdown with confirmation animation
-- Demonstrates: trigger with visual feedback, double-press confirmation
--
-- WARNING: Actual shutdown command commented out for safety

local streamdeck = require("streamdeck")
local system = require("system")
local shell = require("shell")

-- Track confirmation state
-- presses resets after 3 seconds (handled by timeout in state)

-- Passive: show button state
function passive(key, state)
    if state.confirming then
        -- Waiting for confirmation - show red warning
        return {
            color = {200, 0, 0},
            text = "SURE?",
            text_color = {255, 255, 255}
        }
    else
        -- Normal state
        return {
            color = {100, 30, 30},
            text = "OFF",
            text_color = {200, 200, 200}
        }
    end
end

-- Background: reset confirmation after timeout
function background(state)
    while true do
        if state.confirming then
            -- Check if confirmation timed out (3 seconds)
            local now = os.time()
            if now - (state.confirm_time or 0) > 3 then
                state.confirming = false
                print("Shutdown: confirmation timed out")
            end
        end
        system.sleep(500)
    end
end

-- Trigger: handle double-press confirmation
function trigger(state)
    if state.confirming then
        -- Second press within timeout - confirmed!
        state.confirming = false
        print("Shutdown confirmed!")
        
        -- Actual shutdown - UNCOMMENT TO ENABLE
        -- shell.exec("shutdown /s /t 60 /c \"Shutdown initiated from Stream Deck\"")
        
    else
        -- First press - enter confirmation mode
        state.confirming = true
        state.confirm_time = os.time()
        print("Shutdown: press again within 3 seconds to confirm")
    end
end