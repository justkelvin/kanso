# Contributing to Kanso

Thank you for your interest in contributing to **Kanso**! 

Kanso is built as a minimal, personal Arch Linux & Hyprland setup paired with a Go CLI manager. While it is tailored for specific workflows, suggestions, bug reports, and improvements are always welcome.

## Guidelines

1. **Simplicity First**: Kanso adheres to the Japanese concept of *Kanso* (簡素) — simplicity, elimination of clutter, and purity. Keep configurations and code minimal and focused on essentials.
2. **Standard Tooling**: We use Go for the CLI tool, native Lua (`hyprland.lua`) for Hyprland configuration, and official Arch Linux packages wherever possible.
3. **Reproducibility**: Changes to dependency lists (`packages/arch.txt`, `packages/aur.txt`) must be validated against official repositories to avoid broken packages or orphaned dependencies.

## Development Workflow

1. Fork or branch from `master`.
2. Ensure you have Go installed (`go version >= 1.22`).
3. Test your changes locally:
   ```bash
   go build -o build/kanso ./cmd/kanso
   ./build/kanso doctor
   ```
4. Follow conventional commit messages (e.g. `feat: ...`, `fix: ...`, `docs: ...`).
5. Open a Pull Request detailing the changes and reasons behind them.
