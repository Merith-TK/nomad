-- memory.lua - Shows memory usage percentage

local shell = require("shell")
local time  = require("time")
local system = require("system")

local script = {}

function script.passive(key, state)
    local now = time.now()
    if not state.last_update then
        state.last_update = now
        return { color = {0, 255, 0}, text = "MEM\n--%", text_color = {255, 255, 255} }
    elseif (now - state.last_update) >= 5 then
        state.last_update = now
        if system.os() == "windows" then
            -- PowerShell outputs a plain integer e.g. "71"
            local out, _, code = shell.exec("powershell -NoProfile -Command \"$os = Get-CimInstance Win32_OperatingSystem; [math]::Round(($os.TotalVisibleMemorySize - $os.FreePhysicalMemory) / $os.TotalVisibleMemorySize * 100)\"")
            if code == 0 then
                local pct = tonumber(out:match("([%d]+)"))
                if pct then state.memory_percent = pct end
            end
        else
            local out, _, code = shell.exec("free | grep Mem | awk '{printf \"%.0f\", $3/$2 * 100.0}'")
            if code == 0 then
                local pct = tonumber(out:match("([%d]+)"))
                if pct then state.memory_percent = pct end
            end
        end
    end

    local pct   = state.memory_percent or 0
    local color = {0, 255, 0}
    if pct > 90 then
        color = {255, 0, 0}
    elseif pct > 75 then
        color = {255, 165, 0}
    end

    return { color = color, text = string.format("MEM\n%.0f%%", pct), text_color = {255, 255, 255} }
end

return script
