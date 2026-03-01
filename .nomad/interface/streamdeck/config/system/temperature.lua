-- temperature.lua - Shows CPU temperature (if available)
-- Demonstrates: hardware monitoring

local system = require("system")
local shell = require("shell")

-- Passive: show CPU temperature
function passive(key, state)
    -- Try to get temperature via OpenHardwareMonitor or similar
    -- This may not work on all systems
    local temp = nil

    -- Try WMIC first (may not be available)
    local out, _, code = shell.exec("wmic /namespace:\\\\root\\wmi PATH MSAcpi_ThermalZoneTemperature get CurrentTemperature /value 2>nul")
    if code == 0 then
        local raw_temp = out:match("CurrentTemperature=(%d+)")
        if raw_temp then
            temp = (tonumber(raw_temp) - 2732) / 10 -- Convert from tenths of Kelvin
        end
    end

    -- If WMIC failed, try PowerShell with CIM
    if not temp then
        out, _, code = shell.exec("powershell -Command \"Get-CimInstance -Namespace root/WMI -ClassName MSAcpi_ThermalZoneTemperature | Select-Object -ExpandProperty CurrentTemperature | ForEach-Object { ($_ - 2732) / 10 }\" 2>nul")
        if code == 0 then
            temp = tonumber(out:match("([%d%.]+)"))
        end
    end

    if temp then
        state.temperature = temp
    end

    local temp_val = state.temperature or 0
    local color = {0, 255, 0} -- Green for normal temp

    if temp_val > 80 then
        color = {255, 0, 0} -- Red for hot
    elseif temp_val > 65 then
        color = {255, 165, 0} -- Orange for warm
    end

    local display_temp = temp_val > 0 and temp_val or "--"
    return {
        color = color,
        text = string.format("TEMP\n%s°C", display_temp),
        text_color = {255, 255, 255}
    }
end