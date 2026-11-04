-- Keybindings: Application launchers, window navigation, and hardware media keys

local mainMod = "SUPER"
local home = os.getenv("HOME")
local screenshot = home .. "/.config/hypr/scripts/screenshot.sh"
local powermenu = home .. "/.config/hypr/scripts/powermenu.sh"

-- Core Applications
hl.bind(mainMod .. " + RETURN",         hl.dsp.exec_cmd("kitty"))
hl.bind(mainMod .. " + SHIFT + RETURN", hl.dsp.exec_cmd("kitty"))
hl.bind(mainMod .. " + O",              hl.dsp.exec_cmd("rofi -show drun"))
hl.bind(mainMod .. " + E",              hl.dsp.exec_cmd("dolphin"))
hl.bind(mainMod .. " + V",              hl.dsp.exec_cmd("cliphist list | rofi -dmenu | cliphist decode | wl-copy"))
hl.bind(mainMod .. " + W",              hl.dsp.exec_cmd("kanso wallpaper pick"))
hl.bind(mainMod .. " + I",              hl.dsp.exec_cmd("kanso caffeine toggle"))

-- Session & Window Management
hl.bind(mainMod .. " + Q",              hl.dsp.window.close())
hl.bind(mainMod .. " + SHIFT + C",      hl.dsp.window.close())
hl.bind(mainMod .. " + SHIFT + L",      hl.dsp.exec_cmd("hyprlock"))
hl.bind(mainMod .. " + F",              hl.dsp.window.fullscreen())
hl.bind(mainMod .. " + T",              hl.dsp.window.float())
hl.bind("CTRL + ALT + Delete",          hl.dsp.exec_cmd(powermenu))

-- Mouse Window Controls (SUPER + LMB drag to move, SUPER + RMB drag to resize)
hl.bind(mainMod .. " + mouse:272",      hl.dsp.window.drag(),   { mouse = true })
hl.bind(mainMod .. " + mouse:273",      hl.dsp.window.resize(), { mouse = true })

-- Focus Navigation (Vim & Arrow keys)
hl.bind(mainMod .. " + H",              hl.dsp.focus({ direction = "left" }))
hl.bind(mainMod .. " + L",              hl.dsp.focus({ direction = "right" }))
hl.bind(mainMod .. " + K",              hl.dsp.focus({ direction = "up" }))
hl.bind(mainMod .. " + J",              hl.dsp.focus({ direction = "down" }))
hl.bind(mainMod .. " + Left",           hl.dsp.focus({ direction = "left" }))
hl.bind(mainMod .. " + Right",          hl.dsp.focus({ direction = "right" }))
hl.bind(mainMod .. " + Up",             hl.dsp.focus({ direction = "up" }))
hl.bind(mainMod .. " + Down",           hl.dsp.focus({ direction = "down" }))

-- Workspace Switching (1 - 9)
for i = 1, 9 do
    local key = tostring(i)
    hl.bind(mainMod .. " + " .. key,         hl.dsp.focus({ workspace = i }))
    hl.bind(mainMod .. " + SHIFT + " .. key, hl.dsp.window.move({ workspace = i }))
end

-- Laptop Audio Keys (WirePlumber / PipeWire)
hl.bind("XF86AudioRaiseVolume",         hl.dsp.exec_cmd("wpctl set-volume -l 1.5 @DEFAULT_AUDIO_SINK@ 5%+"))
hl.bind("XF86AudioLowerVolume",         hl.dsp.exec_cmd("wpctl set-volume @DEFAULT_AUDIO_SINK@ 5%-"))
hl.bind("XF86AudioMute",                hl.dsp.exec_cmd("wpctl set-mute @DEFAULT_AUDIO_SINK@ toggle"))
hl.bind("XF86AudioMicMute",             hl.dsp.exec_cmd("wpctl set-mute @DEFAULT_AUDIO_SOURCE@ toggle"))

-- Laptop Backlight Keys (brightnessctl)
hl.bind("XF86MonBrightnessUp",          hl.dsp.exec_cmd("brightnessctl set 5%+"))
hl.bind("XF86MonBrightnessDown",        hl.dsp.exec_cmd("brightnessctl set 5%-"))

-- Media Player Keys (playerctl)
hl.bind("XF86AudioPlay",                hl.dsp.exec_cmd("playerctl play-pause"))
hl.bind("XF86AudioNext",                hl.dsp.exec_cmd("playerctl next"))
hl.bind("XF86AudioPrev",                hl.dsp.exec_cmd("playerctl previous"))

-- Screenshots
-- SUPER + P: Select area/window, copy + save, open annotation
hl.bind(mainMod .. " + P",              hl.dsp.exec_cmd(screenshot .. " area"))

-- SUPER + CTRL + P: Same, but freezes the screen first
hl.bind(mainMod .. " + CTRL + P",       hl.dsp.exec_cmd(screenshot .. " area-freeze"))

-- Print: All monitor outputs, copy + save
hl.bind("Print",                        hl.dsp.exec_cmd(screenshot .. " all"))
