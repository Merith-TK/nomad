-- nano.lua - Opens nano text editor

local shell  = require("shell")
local system = require("system")

RESTART_POLICY = "always"

local script = {}

function script.background(state)
    while true do
        if system.os() == "windows" then
            local out, _, code = shell.exec("tasklist /FI \"IMAGENAME eq nano.exe\" /NH 2>nul")
            state.running = (code == 0 and out:find("nano.exe") ~= nil)
        else
            state.running = false
        end
        system.sleep(2000)
    end
end

function script.passive(key, state)
    if state.running then
        return { color = {50, 180, 50}, text = "NP*", text_color = {255, 255, 255} }
    else
        return { color = {80, 80, 80},  text = "NP",  text_color = {200, 200, 200} }
    end
end

function script.trigger(state)
    if system.os() ~= "windows" then
        print("This script only works on Windows")
        return
    end
    shell.terminal("nano")
    system.refresh()
end

return script
