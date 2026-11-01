-- Kanso Hyprland Animations Configuration
-- Defines timing curves and transition behaviors via the native Lua API

hl.config({
    animations = {
        enabled = true,
    },
})

-- Timing Curves (Bezier)
hl.curve("ease_out", {
    type = "bezier",
    points = { { 0.05, 0.9 }, { 0.1, 1.05 } },
})

hl.curve("ease_in_out", {
    type = "bezier",
    points = { { 0.65, 0.05 }, { 0.36, 1 } },
})

-- Animation Scopes
hl.animation({ leaf = "global", enabled = true, speed = 4, bezier = "ease_out" })
hl.animation({ leaf = "windows", enabled = true, speed = 4, bezier = "ease_out" })
hl.animation({ leaf = "windowsIn", enabled = true, speed = 4, bezier = "ease_out", style = "popin 80%" })
hl.animation({ leaf = "windowsOut", enabled = true, speed = 4, bezier = "ease_out", style = "popin 80%" })
hl.animation({ leaf = "border", enabled = true, speed = 5, bezier = "ease_out" })
hl.animation({ leaf = "fade", enabled = true, speed = 4, bezier = "ease_in_out" })
hl.animation({ leaf = "workspaces", enabled = true, speed = 4, bezier = "ease_out", style = "slide" })
