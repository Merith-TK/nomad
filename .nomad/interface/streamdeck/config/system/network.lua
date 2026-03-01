-- network.lua - Shows network connectivity status

local shell = require("shell")
local time  = require("time")

local script = {}

function script.passive(key, state)
    local now = time.now()
    if not state.last_update or (now - state.last_update) >= 10 then
        state.last_update = now
        local out, _, code = shell.exec("timeout 2 ping -c 1 -W 1 8.8.8.8 >/dev/null 2>&1 && echo online || echo offline")
        state.network_online = (code == 0 and out:find("online") ~= nil)
    end

    if state.network_online then
        return { color = {0, 255, 0}, text = "NET\nONLINE", text_color = {255, 255, 255} }
    else
        return { color = {255, 0, 0}, text = "NET\nOFFLINE", text_color = {255, 255, 255} }
    end
end

return script
