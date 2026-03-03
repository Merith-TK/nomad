-- .directory.lua for the media folder
--
-- Drives the folder button (passive) and the two reserved T1/T2 keys:
--   T1  – play / pause toggle
--   T2  – skip to next track
--
-- Requires playerctl to be installed:
--   https://github.com/altdesktop/playerctl
--
-- State is shared across all three functions so we only shell out once per
-- passive tick and reuse the result for both the folder label and T1.

local shell  = require("shell")
local script = {}

-- ── helpers ──────────────────────────────────────────────────────────────────

local function playerctl(args)
    local result = shell.run("playerctl " .. args .. " 2>/dev/null")
    if result and result ~= "" then
        return result:match("^%s*(.-)%s*$") -- trim whitespace
    end
    return nil
end

local function refresh_state(state)
    state.status = playerctl("status") or "Stopped"
    state.artist = playerctl("metadata artist") or ""
    state.title  = playerctl("metadata title")  or "No media"
    state.updated = true
end

-- ── folder button (passive) ───────────────────────────────────────────────────
-- Shows a scrolling "Artist – Title" marquee, or "Stopped" when idle.

function script.passive(key, state)
    refresh_state(state)

    local playing = state.status == "Playing"
    local label

    if not playing then
        label = "MEDIA"
    else
        -- Build a short marquee: trim to 8 chars, advance offset every ~2 ticks
        local full = state.title
        if state.artist ~= "" then
            full = state.artist .. " - " .. full
        end
        if #full <= 8 then
            label = full
        else
            state._offset = ((state._offset or 0) + 1) % #full
            local s = full:sub(state._offset + 1) .. "  " .. full
            label = s:sub(1, 8)
        end
    end

    local bg = playing and {10, 60, 10} or {30, 20, 40}
    return { color = bg, text = label, text_color = {200, 255, 200} }
end

-- ── T1 – play / pause ─────────────────────────────────────────────────────────

function script.t1_passive(key, state)
    local playing = (state.status == "Playing")
    if playing then
        return {
            color      = {0, 100, 0},
            text       = "[||",   -- pause
            text_color = {150, 255, 150},
        }
    else
        return {
            color      = {60, 60, 0},
            text       = "[>]",   -- play
            text_color = {255, 255, 100},
        }
    end
end

function script.t1_trigger(state)
    playerctl("play-pause")
    -- Force state refresh on next passive tick
    state.status = nil
end

-- ── T2 – next track ───────────────────────────────────────────────────────────

function script.t2_passive(key, state)
    return {
        color      = {0, 40, 80},
        text       = ">|>",
        text_color = {100, 180, 255},
    }
end

function script.t2_trigger(state)
    playerctl("next")
    state.status = nil
end

return script
