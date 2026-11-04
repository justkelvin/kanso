-- Window Rules: Set floating state for dialogs, audio mixer, system utilities, and popups

hl.config({
    windowrulev2 = {
        -- Global Safety: prevent windows from requesting maximize
        "suppressevent maximize, class:.*",

        -- Security, Password Managers & Auth
        "float, class:^(Bitwarden)$",
        "float, class:^(org\\.keepassxc\\.KeePassXC)$",
        "float, class:^(hyprpolkitagent)$",
        "float, class:^(pinentry-.*)$",

        -- Audio, Bluetooth & Network System Controls
        "float, class:^(blueman-manager)$",
        "float, class:^(pavucontrol(-qt)?|org\\.pulseaudio\\.pavucontrol)$",
        "float, class:^(nm-(applet|connection-editor))$",

        -- Desktop Portals & Screen Share Pickers
        "float, class:^(hyprland-share-picker)$",
        "float, class:^([Xx]dg-desktop-portal-gtk)$",
        "float, class:^(org\\.freedesktop\\.impl\\.portal\\.desktop\\.(hyprland|gtk))$",

        -- Theme & Display Configuration
        "float, class:^(kvantummanager)$",
        "float, class:^(qt[56]ct)$",
        "float, class:^(nwg-(look|displays))$",

        -- Screenshot Annotation & Archive Utilities
        "float, class:^(com\\.gabm\\.satty)$",
        "float, class:^(org\\.kde\\.ark)$",

        -- Generic Dialog Classes
        "float, class:^(.*dialog.*)$",

        -- File Choosers, Action Dialogs & Prompts (Float + Centered)
        "float, center, title:^(Open File|Save As.*|Choose Files|File Upload.*|Confirm to replace files|File Already Exists.*|Authentication Required)$",
        "float, center, title:^(Progress Dialog.*|Copying.*|File Operation Progress|Add Folder to Workspace|Choose wallpaper.*|Library.*|Volume Control)$",
        "float, center, title:^(.*dialog.*)$",

        -- Picture-in-Picture (Float, Pin, Bottom-Right, 25% Monitor Size)
        "float, title:^([Pp]icture[-\\s]?[Ii]n[-\\s]?[Pp]icture.*)$",
        "pin, title:^([Pp]icture[-\\s]?[Ii]n[-\\s]?[Pp]icture.*)$",
        "size (monitor_w*0.25) (monitor_h*0.25), title:^([Pp]icture[-\\s]?[Ii]n[-\\s]?[Pp]icture.*)$",
        "move (monitor_w*0.73) (monitor_h*0.72), title:^([Pp]icture[-\\s]?[Ii]n[-\\s]?[Pp]icture.*)$",

        -- XWayland Video Bridge (Screen Sharing helper — hidden 1x1 buffer)
        "opacity 0.0 override, class:^(xwaylandvideobridge)$",
        "noanim, class:^(xwaylandvideobridge)$",
        "noinitialfocus, class:^(xwaylandvideobridge)$",
        "maxsize 1 1, class:^(xwaylandvideobridge)$",
        "noblur, class:^(xwaylandvideobridge)$",
    },
})
