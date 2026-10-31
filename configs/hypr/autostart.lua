-- Autostart essential services, daemons, and desktop surfaces on startup

hl.on("hyprland.start", function()
    -- Official polkit authentication agent
    hl.exec_cmd("systemctl --user start hyprpolkitagent")

    -- Desktop surfaces and daemons
    hl.exec_cmd("waybar &")
    hl.exec_cmd("hyprpaper &")
    hl.exec_cmd("hypridle &")
    hl.exec_cmd("mako &")

    -- Network tray applet
    hl.exec_cmd("nm-applet --indicator &")

    -- Clipboard history watcher
    hl.exec_cmd("wl-paste --type text --watch cliphist store &")
    hl.exec_cmd("wl-paste --type image --watch cliphist store &")
end)
