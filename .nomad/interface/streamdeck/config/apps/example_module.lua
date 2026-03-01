-- example_module.lua - Demonstrates the module-based architecture

local system = require("system")
local log    = require("log")

local script = {}

local click_count = 0
local color       = {255, 0, 255} -- Magenta

function script.background(state)
    while true do
        log.debug("background tick, clicks=" .. click_count)
        system.sleep(5000)
    end
end

function script.passive(key, state)
    return {
        color      = color,
        text       = tostring(click_count),
        text_color = {255, 255, 255},
    }
end

function script.trigger(state)
    click_count = click_count + 1
    color = click_count % 2 == 0 and {0, 255, 0} or {255, 0, 255}
    log.info("triggered, count=" .. click_count)
end

return script
