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
| `kanso version` | Show the Kanso CLI version, commit hash, and Go runtime details |
| `kanso help` | Display command-line usage and available subcommands |

---

## Repository Structure

```text
kanso/
├── assets/          # Wallpapers, icons, and visual assets
├── build/           # Compiled CLI binaries (git ignored)
├── cmd/
│   └── kanso/       # Go CLI entrypoint and commands
├── configs/
│   └── hypr/        # Modular Lua configurations for Hyprland
│       ├── hyprland.lua  # Master entrypoint
│       ├── monitors.lua  # Display outputs & scaling
│       ├── autostart.lua # Desktop services & daemons
│       ├── input.lua     # Keyboard, touchpad & gestures
│       ├── keybinds.lua  # Keybindings and application dispatchers
│       ├── look.lua      # Layout, gaps, borders & animations
│       └── rules.lua     # Window rules and workspace assignments
├── packages/
│   ├── arch.txt     # Official repository dependencies
│   ├── aur.txt      # AUR packages (via yay)
│   └── build.txt    # Build prerequisites
├── themes/          # Styling and color themes
├── CONTRIBUTING.md  # Contribution guidelines
├── LICENSE          # MIT License
└── README.md
```

---

## License

This project is licensed under the [MIT License](LICENSE).
