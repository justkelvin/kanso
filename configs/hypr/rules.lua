-- Window Rules: Set floating state for dialogs, audio mixer, and system tools

hl.config({
    windowrulev2 = {
        "float, class:^(pavucontrol)$",
        "float, class:^(blueman-manager)$",
        "float, class:^(nm-connection-editor)$",
        "float, title:^(Open File)$",
        "float, title:^(Save File)$",
        "float, title:^(Confirm to replace files)$",
        "float, title:^(File Operation Progress)$",
        "suppressevent maximize, class:.*",
    },
})
