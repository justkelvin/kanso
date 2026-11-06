# Kanso Minimal Bash Login Profile
# Executed by login shells (TTY1 login)

# 1. Source standard ~/.bashrc if it exists
if [ -f "$HOME/.bashrc" ]; then
    . "$HOME/.bashrc"
fi

# 2. Wayland & Desktop Session Environment Variables
export XDG_CURRENT_DESKTOP=Hyprland
export XDG_SESSION_TYPE=wayland
export XDG_SESSION_DESKTOP=Hyprland

# Toolkit Wayland backends & styling
export GDK_BACKEND="wayland,x11,*"
export QT_QPA_PLATFORM="wayland;xcb"
export QT_STYLE_OVERRIDE=kvantum
export QT_WAYLAND_DISABLE_WINDOWDECORATION=1
export QT_AUTO_SCREEN_SCALE_FACTOR=1
export CLUTTER_BACKEND=wayland
export SDL_VIDEODRIVER=wayland

# Browser & Electron Wayland integration
export MOZ_ENABLE_WAYLAND=1
export ELECTRON_OZONE_PLATFORM_HINT=auto

# Java non-reparenting WM fix
export _JAVA_AWT_WM_NONREPARENTING=1

# Default Editor and Terminal
export EDITOR=nano
export VISUAL=nano
export TERMINAL=kitty

# Credential store & SSH agent integration (if GNOME Keyring SSH socket is active)
if [ -z "$SSH_AUTH_SOCK" ] && [ -S "${XDG_RUNTIME_DIR:-/run/user/$(id -u)}/gcr/ssh" ]; then
    export SSH_AUTH_SOCK="${XDG_RUNTIME_DIR:-/run/user/$(id -u)}/gcr/ssh"
fi

# 3. Autostart Hyprland on tty1 login using the official start-hyprland wrapper
# Boots directly into Hyprland without requiring a display manager.
if [ -z "$DISPLAY" ] && [ "$XDG_VTNR" = "1" ]; then
    exec start-hyprland
fi
