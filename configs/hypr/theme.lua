-- Kanso Hyprland Active Theme
-- Dynamically overwritten when switching themes via 'kanso theme set'

local theme = {
    general = {
        gaps_in = 3,
        gaps_out = 8,
        border_size = 2,
        col = {
            active_border = { colors = { "rgba(bb9af7ff)", "rgba(b4f9f8ff)" }, angle = 45 },
            inactive_border = { colors = { "rgba(565f89cc)", "rgba(9aa5cecc)" }, angle = 45 },
        },
    },
    group = {
        col = {
            border_active = { colors = { "rgba(bb9af7ff)", "rgba(b4f9f8ff)" }, angle = 45 },
            border_inactive = { colors = { "rgba(565f89cc)", "rgba(9aa5cecc)" }, angle = 45 },
            border_locked_active = { colors = { "rgba(bb9af7ff)", "rgba(b4f9f8ff)" }, angle = 45 },
            border_locked_inactive = { colors = { "rgba(565f89cc)", "rgba(9aa5cecc)" }, angle = 45 },
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
