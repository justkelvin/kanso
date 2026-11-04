# Kanso (簡素)

> Minimal, opinionated Arch Linux & Hyprland manager powered by a lightweight Go CLI and native Lua compositor configuration.

---

## Overview

**Kanso** is designed for a lean, distraction-free laptop workflow starting from a bare Arch Linux installation (booting directly into TTY without a display manager). It eliminates desktop environment bloat while providing a fully functional, keyboard-driven Wayland desktop.

### Core Stack

* **Compositor**: [Hyprland](https://hyprland.org) (configured via native Lua, modularized under `~/.config/hypr/`)
* **Terminal**: [Kitty](https://sw.kovidgoyal.net/kitty/) (GPU-accelerated, lightweight)
* **Launcher**: [Rofi](https://github.com/davatorium/rofi) (with native Wayland support)
* **Status Bar**: [Waybar](https://github.com/Alexays/Waybar) (laptop status: battery, audio, backlight, network)
* **Notifications**: [Mako](https://github.com/emersion/mako) (lightweight Wayland notification daemon)
* **Audio**: PipeWire & WirePlumber (`wpctl` bindings for media keys)
* **Hardware & Power**: `brightnessctl`, `hypridle`, `hyprlock`, `hyprpaper`
* **Management CLI**: `kanso` (written in Go with minimal dependencies)

---

## Installation

Boot into your fresh minimal Arch Linux installation, log in at the TTY prompt (`tty1`), and run:

### 1. Install Prerequisites

```bash
sudo pacman -S --needed git go
```

### 2. Clone the Repository

```bash
git clone https://github.com/justkelvin/kanso.git
cd kanso
```

### 3. Run the Kanso CLI

Run directly with Go:

```bash
go run ./cmd/kanso
```

Or build the standalone binary:

```bash
go build -o build/kanso ./cmd/kanso
./build/kanso doctor
```

---

## CLI Commands

The `kanso` CLI provides system management, diagnostics, and deployment tools:

| Command | Description |
| :--- | :--- |
| `kanso` | Launch the interactive TUI installer and configuration wizard |
| `kanso doctor` | Run diagnostic health checks on packages, services, and Wayland readiness |
| `kanso status` | Display the current status of installed desktop packages and configs |
| `kanso wallpaper` | Manage and pick desktop wallpapers (`pick`, `next`, `prev`, `list`, `set`) |
| `kanso theme` | Manage themes and Wallbash dynamic colors (`list`, `get`, `set`, `pick`, `mode`, `mode-menu`, `wallbash`) |
| `kanso caffeine` | Control idle inhibitor state (`toggle`, `on`, `off`, `status`) |
| `kanso version` | Show the Kanso CLI version, commit hash, and Go runtime details |
| `kanso help` | Display command-line usage and available subcommands |

---

## Dynamic Theming & Wallbash

Kanso includes a native, lightweight theming and color engine written in Go with zero shell script overhead. It unifies presets and dynamic wallpaper-driven colors across Kitty, Rofi, Waybar, Hyprland, GTK, Qt, and Kvantum.

### 1. Keybindings

* <kbd>SUPER</kbd> + <kbd>SHIFT</kbd> + <kbd>T</kbd>: Opens the visual Rofi theme picker grid (`kanso theme pick`).
* <kbd>SUPER</kbd> + <kbd>SHIFT</kbd> + <kbd>R</kbd>: Opens the Rofi Wallbash mode selector (`kanso theme mode-menu`).
* <kbd>SUPER</kbd> + <kbd>W</kbd>: Opens the visual Rofi wallpaper picker (`kanso wallpaper pick`).

### 2. Wallbash Modes

| Mode | Value | Behavior |
| :--- | :---: | :--- |
| **theme** | `0` | Locked to the selected theme's own presets. Changing wallpapers will not recolor applications. |
| **auto** | `1` | Automatically samples the wallpaper luminance; dynamically extracts a matching dark or light palette via K-means clustering. |
| **dark** | `2` | Extracts a forced dark palette from the active wallpaper with high contrast text. |
| **light** | `3` | Extracts a forced light palette from the active wallpaper with crisp dark text. |

### 3. Included Themes

* **Tokyo Night**: Classic deep blue-violet neon palette with subtle gradients.
* **Catppuccin Mocha**: Soothing dark pastel aesthetic.
* **Catppuccin Latte**: Bright, crisp light theme with pastel accents.
* **Graphite Mono**: Minimalist grayscale monochrome design with static `.dcol` overrides.

Each theme provides its own Hyprland layout rules, Kitty terminal colors, Rofi styling, Waybar CSS variables, Kvantum Qt widget skins, wallpapers, and GTK icon/cursor settings.

---

## Post-Installation

Kanso intentionally avoids modifying sensitive system security files (such as PAM authentication modules) automatically. After completing `kanso install`, perform the following one-time steps on your fresh installation:

### 1. Keyring Auto-Unlock via PAM (`/etc/pam.d/login`)

Kanso installs `gnome-keyring` and `libsecret` so web browsers (Google Chrome, Chromium, Brave), code editors (VS Code, Zed), and Git can securely store credentials and tokens without plaintext fallbacks.

To allow your TTY1 login password to automatically unlock the `login` keyring on boot (preventing manual password prompts when opening applications), edit `/etc/pam.d/login`:

```bash
sudo nano /etc/pam.d/login
```

Add `pam_gnome_keyring.so` to the `auth` and `session` sections:

```pam
#%PAM-1.0

auth       requisite    pam_nologin.so
auth       include      system-local-login
auth       optional     pam_gnome_keyring.so

account    include      system-local-login

session    include      system-local-login
session    optional     pam_gnome_keyring.so auto_start

password   include      system-local-login
```

*(Optional)* To keep your keyring password in sync whenever you change your user password using `passwd`, add the following to `/etc/pam.d/passwd`:

```pam
password   optional     pam_gnome_keyring.so
```

### 2. User Hardware Groups (`video`, `input`)

Ensure your user account belongs to the `video` and `input` groups for backlight control (`brightnessctl`) and Wayland input handling:

```bash
sudo usermod -aG video,input $USER
```

> **Note:** Group changes take effect on your next login session. Run `kanso doctor` to verify group membership.

### 3. Git Credential Helper (Optional)

Configure Git to securely persist GitHub and remote repository credentials in GNOME Keyring via `libsecret`:

```bash
git config --global credential.helper libsecret
```

---

## Repository Structure

```text
kanso/
├── assets/          # Wallpapers, icons, and visual assets
├── build/           # Compiled CLI binaries (git ignored)
├── cmd/
│   └── kanso/       # Go CLI entrypoint and commands
├── configs/
│   ├── bash/        # Login profile and shell environment (.bashrc, .bash_profile)
│   ├── dolphin/     # Minimal Dolphin file manager configuration (dolphinrc)
│   ├── electron/    # Wayland runtime flags for Electron and VS Code
│   ├── hypr/        # Modular Lua configurations for Hyprland
│   │   ├── hyprland.lua  # Master entrypoint
│   │   ├── monitors.lua  # Display outputs & scaling
│   │   ├── autostart.lua # Desktop services & daemons
│   │   ├── input.lua     # Keyboard, touchpad & gestures
│   │   ├── keybinds.lua  # Keybindings and application dispatchers
│   │   ├── look.lua      # Layout, gaps, borders & animations
│   │   └── rules.lua     # Window rules and workspace assignments
│   ├── kde/         # Tokyo Night color scheme and Qt globals (kdeglobals)
│   ├── kitty/       # Terminal emulator styling and Tokyo Night theme
│   ├── mako/        # Notification daemon styling
│   ├── rofi/        # Tokyo Night application launcher & wallpaper grid
│   ├── satty/       # Wayland screenshot annotator config (config.toml)
│   ├── wallbash/    # Dynamic color templates (.dcol) for Kitty, Waybar, Rofi, Hyprland
│   ├── waybar/      # Status bar modules and styling
│   └── xdg-desktop-portal/ # Portal routing overrides (hyprland-portals.conf)
├── packages/
│   ├── arch.txt     # Official repository dependencies
│   ├── aur.txt      # AUR packages (via yay)
│   └── build.txt    # Build prerequisites
├── themes/          # Theme presets (Catppuccin Mocha/Latte, Tokyo Night, Graphite Mono)
├── CONTRIBUTING.md  # Contribution guidelines
├── LICENSE          # MIT License
└── README.md
```

---

## License

This project is licensed under the [MIT License](LICENSE).
