-- disk.lua - Shows disk usage for root filesystem

local shell = require("shell")
local time  = require("time")
local system = require("system")

local script = {}

function script.passive(key, state)
    local now = time.now()
    if not state.last_update then
        state.last_update = now
        return { color = {0, 255, 0}, text = "DISK\n--%", text_color = {255, 255, 255} }
    elseif (now - state.last_update) >= 30 then
        state.last_update = now
        if system.os() == "windows" then
            -- PowerShell outputs a plain integer e.g. "91"
            local out, _, code = shell.exec("powershell -NoProfile -Command \"$d = Get-PSDrive C; [math]::Round($d.Used / ($d.Used + $d.Free) * 100)\"")
            if code == 0 then
                local pct = tonumber(out:match("([%d]+)"))
                if pct then state.disk_percent = pct end
            end
        else
            local out, _, code = shell.exec("df / | tail -1 | awk '{print $5}' | sed 's/%//'")
            if code == 0 then
                local pct = tonumber(out:match("([%d]+)"))
                if pct then state.disk_percent = pct end
            end
        end
    end

    local pct   = state.disk_percent or 0
    local color = {0, 255, 0}
    if pct > 95 then
        color = {255, 0, 0}
    elseif pct > 85 then
        color = {255, 165, 0}
    end

    return { color = color, text = string.format("DISK\n%.0f%%", pct), text_color = {255, 255, 255} }
end

return script
