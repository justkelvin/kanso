-- Kanso Hyprland Native Lua Theme: Catppuccin Latte
-- Warm light aesthetic with pastel border accents

local theme = {
    general = {
        gaps_in = 3,
        gaps_out = 8,
        border_size = 2,
        col = {
            active_border = { colors = { "rgba(dc8a78ff)", "rgba(8839efff)" }, angle = 45 },
            inactive_border = { colors = { "rgba(7287fdcc)", "rgba(179299cc)" }, angle = 45 },
        },
    },
    group = {
        col = {
            border_active = { colors = { "rgba(dc8a78ff)", "rgba(8839efff)" }, angle = 45 },
            border_inactive = { colors = { "rgba(7287fdcc)", "rgba(179299cc)" }, angle = 45 },
            border_locked_active = { colors = { "rgba(dc8a78ff)", "rgba(8839efff)" }, angle = 45 },
            border_locked_inactive = { colors = { "rgba(7287fdcc)", "rgba(179299cc)" }, angle = 45 },
        },
    },
    decoration = {
        rounding = 10,
        shadow = {
            enabled = false,
        },
        blur = {
            enabled = true,
            size = 6,
            passes = 3,
        },
    },
}

if hl and hl.config then
    hl.config(theme)
    if hl.layer_rule then
        hl.layer_rule({ match = { namespace = "waybar" }, blur = true })
    end
end

return theme
