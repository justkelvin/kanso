-- Input devices: Keyboard, laptop touchpad, and multi-finger gestures

hl.config({
    input = {
        kb_layout = "us",
        follow_mouse = 1,
        sensitivity = 0, -- -1.0 - 1.0, 0 means no modification.

        touchpad = {
            natural_scroll = true,
            tap_to_click = true,
            disable_while_typing = true,
            middle_button_emulation = true,
        },
    },
})

-- Touchpad Gestures
hl.gesture({
    fingers = 3,
    direction = "horizontal",
    action = "workspace",
})
