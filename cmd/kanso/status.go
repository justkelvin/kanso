package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show deployment status of configs, services, and session",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(titleStyle.Render("Kanso Status Overview"))

		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error resolving user home directory: %v\n", err)
			return
		}

		// 1. Session Environment
		fmt.Println(sectionStyle.Render("Session Environment"))
		sessionType := os.Getenv("XDG_SESSION_TYPE")
		currentDesktop := os.Getenv("XDG_CURRENT_DESKTOP")
		waylandDisplay := os.Getenv("WAYLAND_DISPLAY")

		if sessionType == "" {
			sessionType = "none (TTY or console)"
		}
		if currentDesktop == "" {
			currentDesktop = "none"
		}
		if waylandDisplay == "" {
			waylandDisplay = "none"
		}

		fmt.Printf("  %s %-20s %s\n", dimStyle.Render("•"), "Session Type:", sessionType)
		fmt.Printf("  %s %-20s %s\n", dimStyle.Render("•"), "Current Desktop:", currentDesktop)
		fmt.Printf("  %s %-20s %s\n", dimStyle.Render("•"), "Wayland Display:", waylandDisplay)

		// 2. Configuration Files
		fmt.Println(sectionStyle.Render("Configuration Files"))
		configs := []struct {
			name string
			path string
		}{
			{"Hyprland Config", filepath.Join(home, ".config", "hypr", "hyprland.lua")},
			{"Hypridle Config", filepath.Join(home, ".config", "hypr", "hypridle.conf")},
			{"Hyprlock Config", filepath.Join(home, ".config", "hypr", "hyprlock.conf")},
			{"Hyprpaper Config", filepath.Join(home, ".config", "hypr", "hyprpaper.conf")},
			{"Waybar Config", filepath.Join(home, ".config", "waybar", "config.jsonc")},
			{"Kitty Config", filepath.Join(home, ".config", "kitty", "kitty.conf")},
			{"Rofi Config", filepath.Join(home, ".config", "rofi", "config.rasi")},
			{"Mako Config", filepath.Join(home, ".config", "mako", "config")},
		}

		deployedCount := 0
		for _, cfg := range configs {
			if _, err := os.Stat(cfg.path); err == nil {
				fmt.Printf("  %s %-20s %s\n", successStyle.Render("✔"), cfg.name, dimStyle.Render(cfg.path))
				deployedCount++
			} else {
				fmt.Printf("  %s %-20s %s\n", warnStyle.Render("○"), cfg.name, dimStyle.Render("Not installed"))
			}
		}

		// 3. User Services
		fmt.Println(sectionStyle.Render("User Services"))
		services := []string{"pipewire.service", "wireplumber.service", "hyprpolkitagent.service"}
		for _, svc := range services {
			out, err := exec.Command("systemctl", "--user", "is-active", svc).Output()
			statusStr := strings.TrimSpace(string(out))
			if err == nil && statusStr == "active" {
				fmt.Printf("  %s %-25s %s\n", successStyle.Render("✔"), svc, successStyle.Render("active"))
			} else {
				if statusStr == "" {
					statusStr = "inactive"
				}
				fmt.Printf("  %s %-25s %s\n", dimStyle.Render("•"), svc, dimStyle.Render(statusStr))
			}
		}

		// Summary
		fmt.Println()
		fmt.Printf("Configurations deployed: %d/%d\n", deployedCount, len(configs))
		if deployedCount < len(configs) {
			fmt.Printf("Run %s to deploy missing configurations.\n", successStyle.Render("kanso install"))
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
