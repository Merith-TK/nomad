-- _boot.lua - Boot animation shown while scripts load
-- Place in config root directory to customize boot sequence
--
-- The boot() function is called once when the interface starts up.
-- This script runs BEFORE other scripts are loaded.

local streamdeck = require("streamdeck")
local system = require("system")

-- Boot animation: sweep a color across the keys
function boot()
    local cols, rows = streamdeck.get_layout()
    local keys = streamdeck.get_keys()
    
    -- Clear all keys
    streamdeck.clear()
    
    -- Sweep blue from left to right
    for col = 0, cols - 1 do
        for row = 0, rows - 1 do
            local key = row * cols + col
            -- Blue gradient based on column
            local intensity = math.floor(100 + (col / cols) * 155)
            streamdeck.set_color(key, 0, 0, intensity)
        end
        system.sleep(100)
    end
    
    -- Brief pause
    system.sleep(200)
    
    -- Fade out
    for brightness = 100, 0, -20 do
        streamdeck.set_brightness(brightness)
        system.sleep(50)
    end
    
    -- Clear and restore brightness
    streamdeck.clear()
    streamdeck.set_brightness(75)
end
