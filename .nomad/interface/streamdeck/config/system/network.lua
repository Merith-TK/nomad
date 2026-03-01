-- network.lua - Shows network connectivity status
-- Demonstrates: network monitoring with passive updates

local system = require("system")
local shell = require("shell")

-- Passive: show network status
function passive(key, state)
    -- Only update every 10 seconds to reduce load
    local now = os.time()
    if not state.last_update or (now - state.last_update) >= 10 then
        state.last_update = now

        -- Check if we can reach a common internet host (Linux)
        local out, _, code = shell.exec("timeout 2 ping -c 1 -W 1 8.8.8.8 >/dev/null 2>&1 && echo online || echo offline")

        local is_online = false
        if code == 0 and out:find("online") then
            is_online = true
        end

        state.network_online = is_online
    end

    local is_online = state.network_online
    local color = {255, 0, 0} -- Red for offline
    local status = "OFFLINE"

    if is_online then
        color = {0, 255, 0} -- Green for online
        status = "ONLINE"
    end

    return {
        color = color,
        text = string.format("NET\n%s", status),
        text_color = {255, 255, 255}
    }
end