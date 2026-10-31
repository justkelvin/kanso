# Kanso

A personal Arch Linux / Hyprland setup with a minimal Go CLI, native compositor configuration, and minimal Bash.

## Installation

Boot into your fresh minimal Arch Linux installation and log into the TTY (`tty1`).

### 1. Install prerequisites

Install `git` and `go` using `pacman`:

```bash
sudo pacman -S --needed git go
```

### 2. Clone the repository

```bash
git clone https://github.com/your-username/kanso.git
cd kanso
```

### 3. Run the installer CLI

```bash
go run ./cmd/kanso
```
