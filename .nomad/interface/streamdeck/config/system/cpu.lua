-- cpu.lua - Shows CPU usage percentage
-- Demonstrates: system monitoring with passive updates

local system = require("system")
local shell = require("shell")

-- Passive: show current CPU usage
function passive(key, state)
    -- Get CPU usage via Windows performance counter
    local out, _, code = shell.exec("powershell -Command \"Get-Counter '\\Processor(_Total)\\% Processor Time' -SampleInterval 1 -MaxSamples 1 | Select-Object -ExpandProperty CounterSamples | Select-Object -ExpandProperty CookedValue\"")
    if code == 0 then
        local cpu = tonumber(out:match("([%d%.]+)"))
        if cpu then
            state.cpu = cpu
        end
    end

    local cpu = state.cpu or 0
    local color = {0, 255, 0} -- Green for low usage

    if cpu > 80 then
        color = {255, 0, 0} -- Red for high usage
    elseif cpu > 60 then
        color = {255, 165, 0} -- Orange for medium usage
    end

    return {
        color = color,
        text = string.format("CPU\n%.0f%%", cpu),
        text_color = {255, 255, 255}
    }
end