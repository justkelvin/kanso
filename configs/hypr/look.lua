-- Window layout, spacing, borders, colors, and decorations

hl.config({
    general = {
        gaps_in = 4,
        gaps_out = 8,
        border_size = 2,
        col = {
            active_border = "rgba(7aa2f7ee)",     -- Tokyo night accent / clean blue
            inactive_border = "rgba(24283baa)",   -- Subtle dark muted border
        },
        layout = "dwindle",
        resize_on_border = true,
        extend_border_grab_area = 15,
        hover_icon_on_border = true,
        float_force_onscreen = 2,
    },

    decoration = {
        rounding = 8,
        active_opacity = 1.0,
        inactive_opacity = 0.95,

        shadow = {
            enabled = false, -- Keep minimal & battery-friendly on laptops
        },

        blur = {
            enabled = true,
            size = 4,
            passes = 2,
            vibrancy = 0.1696,
        },
    },

    dwindle = {
        preserve_split = true,
    },

    misc = {
        disable_hyprland_logo = true,
        disable_splash_rendering = true,
        force_default_wallpaper = 0,
        background_color = "0x1a1b26",
        new_float_force_onscreen = 2,
    },
})
