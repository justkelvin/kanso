-- Kanso Hyprland Native Lua Theme: Catppuccin Mocha
-- Pastel accents with rounded window geometry and subtle gradients

local theme = {
    general = {
        gaps_in = 3,
        gaps_out = 8,
        border_size = 2,
        col = {
            active_border = { colors = { "rgba(ca9ee6ff)", "rgba(f2d5cfff)" }, angle = 45 },
            inactive_border = { colors = { "rgba(b4befecc)", "rgba(6c7086cc)" }, angle = 45 },
        },
    },
    group = {
        col = {
            border_active = { colors = { "rgba(ca9ee6ff)", "rgba(f2d5cfff)" }, angle = 45 },
            border_inactive = { colors = { "rgba(b4befecc)", "rgba(6c7086cc)" }, angle = 45 },
            border_locked_active = { colors = { "rgba(ca9ee6ff)", "rgba(f2d5cfff)" }, angle = 45 },
            border_locked_inactive = { colors = { "rgba(b4befecc)", "rgba(6c7086cc)" }, angle = 45 },
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
