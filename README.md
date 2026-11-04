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
| `kanso caffeine` | Control idle inhibitor state (`toggle`, `on`, `off`, `status`) |
| `kanso version` | Show the Kanso CLI version, commit hash, and Go runtime details |
| `kanso help` | Display command-line usage and available subcommands |

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
