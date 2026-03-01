-- network.lua - Shows network connectivity status
-- Demonstrates: network monitoring with passive updates

local system = require("system")
local shell = require("shell")

-- Passive: show network status
function passive(key, state)
    -- Check if we can reach a common internet host
    local out, _, code = shell.exec("ping -n 1 -w 1000 8.8.8.8 >nul 2>&1 && echo online || echo offline")

    local is_online = false
    if code == 0 and out:find("online") then
        is_online = true
    end

    state.network_online = is_online

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