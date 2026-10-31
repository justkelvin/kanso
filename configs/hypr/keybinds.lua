-- Keybindings: Application launchers, window navigation, and hardware media keys

local mainMod = "SUPER"

-- Core Applications
hl.bind(mainMod .. " + RETURN", hl.dsp.exec_cmd("kitty"))
hl.bind(mainMod .. " + SPACE",  hl.dsp.exec_cmd("rofi -show drun"))
hl.bind(mainMod .. " + E",      hl.dsp.exec_cmd("dolphin"))
hl.bind(mainMod .. " + V",      hl.dsp.exec_cmd("cliphist list | rofi -dmenu | cliphist decode | wl-copy"))

-- Window Management
hl.bind(mainMod .. " + Q",      hl.dsp.window.close())
hl.bind(mainMod .. " + M",      hl.dsp.dispatch("exit"))
hl.bind(mainMod .. " + F",      hl.dsp.window.fullscreen())
hl.bind(mainMod .. " + T",      hl.dsp.window.toggle_floating())

-- Focus Navigation (Vim & Arrow keys)
hl.bind(mainMod .. " + H",      hl.dsp.window.move_focus("l"))
hl.bind(mainMod .. " + L",      hl.dsp.window.move_focus("r"))
hl.bind(mainMod .. " + K",      hl.dsp.window.move_focus("u"))
hl.bind(mainMod .. " + J",      hl.dsp.window.move_focus("d"))
hl.bind(mainMod .. " + Left",   hl.dsp.window.move_focus("l"))
hl.bind(mainMod .. " + Right",  hl.dsp.window.move_focus("r"))
hl.bind(mainMod .. " + Up",     hl.dsp.window.move_focus("u"))
hl.bind(mainMod .. " + Down",   hl.dsp.window.move_focus("d"))

-- Workspace Switching (1 - 9)
for i = 1, 9 do
    hl.bind(mainMod .. " + " .. i, hl.dsp.workspace.switch(tostring(i)))
    hl.bind(mainMod .. " + SHIFT + " .. i, hl.dsp.window.move_to_workspace(tostring(i)))
end

-- Laptop Audio Keys (WirePlumber / PipeWire)
hl.bind("XF86AudioRaiseVolume",  hl.dsp.exec_cmd("wpctl set-volume -l 1.5 @DEFAULT_AUDIO_SINK@ 5%+"))
hl.bind("XF86AudioLowerVolume",  hl.dsp.exec_cmd("wpctl set-volume @DEFAULT_AUDIO_SINK@ 5%-"))
hl.bind("XF86AudioMute",         hl.dsp.exec_cmd("wpctl set-mute @DEFAULT_AUDIO_SINK@ toggle"))
hl.bind("XF86AudioMicMute",      hl.dsp.exec_cmd("wpctl set-mute @DEFAULT_AUDIO_SOURCE@ toggle"))

-- Laptop Backlight Keys (brightnessctl)
hl.bind("XF86MonBrightnessUp",   hl.dsp.exec_cmd("brightnessctl set 5%+"))
hl.bind("XF86MonBrightnessDown", hl.dsp.exec_cmd("brightnessctl set 5%-"))

-- Media Player Keys (playerctl)
hl.bind("XF86AudioPlay",         hl.dsp.exec_cmd("playerctl play-pause"))
hl.bind("XF86AudioNext",         hl.dsp.exec_cmd("playerctl next"))
hl.bind("XF86AudioPrev",         hl.dsp.exec_cmd("playerctl previous"))

-- Screenshot Utility (Grim + Slurp directly to clipboard)
hl.bind(mainMod .. " + SHIFT + S", hl.dsp.exec_cmd("grim -g \"$(slurp)\" - | wl-copy"))
