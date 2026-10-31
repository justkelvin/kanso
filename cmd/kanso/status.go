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
		PrintBanner()

		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error resolving user home directory: %v\n", err)
			return
		}

		// 1. Session Info Card
		sessionType := os.Getenv("XDG_SESSION_TYPE")
		currentDesktop := os.Getenv("XDG_CURRENT_DESKTOP")
		waylandDisplay := os.Getenv("WAYLAND_DISPLAY")

		if sessionType == "" {
			sessionType = "tty"
		}
		if currentDesktop == "" {
			currentDesktop = "none"
		}
		if waylandDisplay == "" {
			waylandDisplay = "none"
		}

		sessionBadge := BadgeGreen("RUNNING")
		if waylandDisplay == "none" {
			sessionBadge = BadgeYellow("TTY (IDLE)")
		}

		sessionCard := RenderCard("Session info", []KeyVal{
			{Key: "Desktop", Value: currentDesktop},
			{Key: "Session Type", Value: sessionType},
			{Key: "Wayland Display", Value: waylandDisplay},
			{Key: "State", Badge: sessionBadge},
		})
		fmt.Println(sessionCard)

		// 2. Configuration Files Card
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
		var configRows []KeyVal
		for _, cfg := range configs {
			if _, err := os.Stat(cfg.path); err == nil {
				configRows = append(configRows, KeyVal{
					Key:   cfg.name,
					Value: cfg.path,
					Badge: BadgeGreen("DEPLOYED"),
				})
				deployedCount++
			} else {
				configRows = append(configRows, KeyVal{
					Key:   cfg.name,
					Value: "missing",
					Badge: BadgeMuted("NOT INSTALLED"),
				})
			}
		}

		configCard := RenderCard(fmt.Sprintf("Configurations (%d/%d)", deployedCount, len(configs)), configRows)
		fmt.Println(configCard)

		// 3. User Services Card
		services := []string{"pipewire.service", "wireplumber.service", "hyprpolkitagent.service"}
		var serviceRows []KeyVal
		for _, svc := range services {
			out, err := exec.Command("systemctl", "--user", "is-active", svc).Output()
			statusStr := strings.TrimSpace(string(out))
			if err == nil && statusStr == "active" {
				serviceRows = append(serviceRows, KeyVal{
					Key:   svc,
					Badge: BadgeGreen("ACTIVE"),
				})
			} else {
				serviceRows = append(serviceRows, KeyVal{
					Key:   svc,
					Badge: BadgeMuted("INACTIVE"),
				})
			}
		}

		serviceCard := RenderCard("User services", serviceRows)
		fmt.Println(serviceCard)

		if deployedCount < len(configs) {
			fmt.Printf("Run %s to deploy missing configurations.\n\n", BadgeGreen("kanso install"))
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
