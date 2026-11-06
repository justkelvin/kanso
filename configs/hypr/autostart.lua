-- Autostart essential services, daemons, and desktop surfaces on startup

hl.on("hyprland.start", function()
    -- Environment synchronization for D-Bus and systemd user services
    hl.exec_cmd("dbus-update-activation-environment --systemd WAYLAND_DISPLAY XDG_CURRENT_DESKTOP QT_STYLE_OVERRIDE")
    hl.exec_cmd("systemctl --user import-environment WAYLAND_DISPLAY XDG_CURRENT_DESKTOP QT_STYLE_OVERRIDE")

    -- Official polkit authentication agent
    hl.exec_cmd("systemctl --user start hyprpolkitagent")

    -- Credential store daemon (Secret Service)
    hl.exec_cmd("gnome-keyring-daemon --start --components=secrets &")

    -- Desktop surfaces and daemons
    hl.exec_cmd("waybar &")
    hl.exec_cmd("hyprpaper &")
    hl.exec_cmd("hypridle &")
    hl.exec_cmd("mako &")

    -- Network & Bluetooth tray applets
    hl.exec_cmd("nm-applet --indicator &")
    hl.exec_cmd("blueman-applet &")

    -- Clipboard history watcher
    hl.exec_cmd("wl-paste --type text --watch cliphist store &")
    hl.exec_cmd("wl-paste --type image --watch cliphist store &")
end)
