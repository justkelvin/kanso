-- Kanso Hyprland Window Rules
-- Native Lua window_rule declarations for Hyprland 0.57+

-- 1. Floating Utility Classes (Audio Mixer, Network, Bluetooth, Auth, Theme Managers, Dialogs)
hl.window_rule({
    name = "kanso_floating_classes",
    match = {
        class = "^(pavucontrol(-qt)?|org\\.pulseaudio\\.pavucontrol|nm-(applet|connection-editor)|blueman-manager|Bitwarden|org\\.keepassxc\\.KeePassXC|seahorse|gcr-prompter|hyprpolkitagent|pinentry-.*|hyprland-share-picker|[Xx]dg-desktop-portal-gtk|org\\.freedesktop\\.impl\\.portal\\.desktop\\.(hyprland|gtk)|kvantummanager|qt[56]ct|nwg-(look|displays)|com\\.gabm\\.satty|org\\.kde\\.ark|.*dialog.*)$",
    },
    float = true,
})

-- 2. Floating & Centered Action Modals & File Choosers
hl.window_rule({
    name = "kanso_floating_titles",
    match = {
        title = "^(Open File|Save As.*|Choose Files|File Upload.*|Confirm to replace files|File Already Exists.*|Authentication Required|Progress Dialog.*|Copying.*|File Operation Progress|Add Folder to Workspace|Choose wallpaper.*|Library.*|Volume Control|.*dialog.*)$",
    },
    float = true,
    center = true,
})

-- 3. Picture-in-Picture (Float, Pin, Bottom-Right)
hl.window_rule({
    name = "kanso_pip",
    match = {
        title = "^([Pp]icture[-\\s]?[Ii]n[-\\s]?[Pp]icture.*)$",
    },
    float = true,
    pin = true,
    size = "(monitor_w*0.25) (monitor_h*0.25)",
    move = "(monitor_w*0.73) (monitor_h*0.72)",
})

-- 4. XWayland Video Bridge (Screen Sharing helper — hidden 1x1 buffer)
hl.window_rule({
    name = "kanso_xwayland_video_bridge",
    match = {
        class = "xwaylandvideobridge",
    },
    no_initial_focus = true,
    no_anim = true,
    no_blur = true,
    max_size = { 1, 1 },
    opacity = 0.0,
})

-- 5. Global suppression of maximize requests
hl.window_rule({
    name = "kanso_suppress_maximize",
    match = {
        class = ".*",
    },
    suppress_event = "maximize",
})
