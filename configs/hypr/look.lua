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
})
