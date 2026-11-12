-- Window layout, spacing, borders, colors, and decorations

hl.config({
    general = {
        layout = "dwindle",
        resize_on_border = true,
        extend_border_grab_area = 15,
        hover_icon_on_border = true,
    },

    decoration = {
        active_opacity = 1.0,
        inactive_opacity = 0.95,
    },

    dwindle = {
        preserve_split = true,
    },

    misc = {
        disable_hyprland_logo = true,
        disable_splash_rendering = true,
        force_default_wallpaper = 0,
        float_force_onscreen = 2,
        new_float_force_onscreen = 2,
    },
})
