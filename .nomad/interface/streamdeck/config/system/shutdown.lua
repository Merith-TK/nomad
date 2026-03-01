-- shutdown.lua - System shutdown with double-press confirmation
--
-- WARNING: Actual shutdown command commented out for safety

local system = require("system")
local shell  = require("shell")

local script = {}

function script.passive(key, state)
    if state.confirming then
        return { color = {200, 0, 0}, text = "SURE?", text_color = {255, 255, 255} }
    else
        return { color = {100, 30, 30}, text = "OFF", text_color = {200, 200, 200} }
    end
end

function script.background(state)
    while true do
        if state.confirming then
            local now = os.time()
            if now - (state.confirm_time or 0) > 3 then
                state.confirming = false
                print("Shutdown: confirmation timed out")
            end
        end
        system.sleep(500)
    end
end

function script.trigger(state)
    if state.confirming then
        state.confirming = false
        print("Shutdown confirmed!")
        -- Uncomment to enable:
        -- shell.exec("shutdown /s /t 60 /c \"Shutdown initiated from Stream Deck\"")
    else
        state.confirming   = true
        state.confirm_time = os.time()
        print("Shutdown: press again within 3 seconds to confirm")
    end
end

return script
