-- network.lua - Shows network connectivity status

local shell = require("shell")
local time  = require("time")
local system = require("system")

local script = {}

function script.passive(key, state)
    local now = time.now()
    if not state.last_update then
        state.last_update = now
        return { color = {255, 0, 0}, text = "NET\n--", text_color = {255, 255, 255} }
    elseif (now - state.last_update) >= 10 then
        state.last_update = now
        local cmd
        if system.os() == "windows" then
            cmd = "ping -n 1 -w 1000 8.8.8.8"
        else
            cmd = "ping -c 1 -W 1 8.8.8.8 > /dev/null 2>&1"
        end
        local _, _, code = shell.exec(cmd)
        state.network_online = (code == 0)
    end

    if state.network_online then
        return { color = {0, 255, 0}, text = "NET\nONLINE", text_color = {255, 255, 255} }
    else
        return { color = {255, 0, 0}, text = "NET\nOFFLINE", text_color = {255, 255, 255} }
    end
end

return script
