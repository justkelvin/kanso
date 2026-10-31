package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a")).Bold(true)
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#e0af68")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#f7768e")).Bold(true)
	sectionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7aa2f7")).Bold(true).MarginTop(1)
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#565f89"))
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run diagnostics on system dependencies and Wayland readiness",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(titleStyle.Render("Kanso Doctor — System Diagnostics"))

		issues := 0

		// 1. Operating System
		fmt.Println(sectionStyle.Render("Operating System"))
		if isArchLinux() {
			fmt.Printf("  %s Base distribution is Arch Linux\n", successStyle.Render("✔"))
		} else {
			fmt.Printf("  %s Non-Arch Linux system detected. Kanso is optimized for Arch.\n", warnStyle.Render("!"))
			issues++
		}

		// 2. Core Desktop Tools
		fmt.Println(sectionStyle.Render("Core Desktop Tools"))
		essentialBinaries := []struct {
			name        string
			description string
		}{
			{"hyprland", "Wayland compositor"},
			{"kitty", "Default terminal emulator"},
			{"rofi", "Application launcher"},
			{"waybar", "Status bar"},
			{"mako", "Notification daemon"},
			{"pipewire", "Audio server"},
			{"wireplumber", "PipeWire session manager"},
			{"wpctl", "WirePlumber audio controller"},
			{"brightnessctl", "Backlight controller"},
			{"hyprlock", "Screen locker"},
			{"hypridle", "Idle management daemon"},
			{"hyprpaper", "Wallpaper daemon"},
			{"grim", "Screen capture utility"},
			{"slurp", "Wayland region selector"},
			{"wl-copy", "Wayland clipboard tool"},
			{"cliphist", "Clipboard history manager"},
		}

		for _, bin := range essentialBinaries {
			if path, err := exec.LookPath(bin.name); err == nil {
				fmt.Printf("  %s %-15s %s\n", successStyle.Render("✔"), bin.name, dimStyle.Render("("+path+")"))
			} else {
				fmt.Printf("  %s %-15s %s (missing: %s)\n", errorStyle.Render("✖"), bin.name, dimStyle.Render(bin.description), bin.name)
				issues++
			}
		}

		// 3. Portals and Polkit
		fmt.Println(sectionStyle.Render("Session Integration"))
		checkPortalAndPolkit(&issues)

		// 4. User Group Permissions
		fmt.Println(sectionStyle.Render("User Groups"))
		checkUserGroups(&issues)

		// 5. Desktop Configurations
		fmt.Println(sectionStyle.Render("Configuration Files"))
		checkConfigs(&issues)

		// Summary
		fmt.Println()
		if issues == 0 {
			fmt.Printf("%s All diagnostic checks passed! System is ready for Hyprland.\n", successStyle.Render("✔"))
		} else {
			fmt.Printf("%s Doctor found %d issue(s) or missing package(s). Run 'kanso' to configure.\n", warnStyle.Render("!"), issues)
		}
	},
}

func isArchLinux() bool {
	if _, err := os.Stat("/etc/arch-release"); err == nil {
		return true
	}
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") && strings.Contains(strings.ToLower(line), "arch") {
			return true
		}
	}
	return false
}

func checkPortalAndPolkit(issues *int) {
	// Check hyprpolkitagent
	if _, err := exec.LookPath("hyprpolkitagent"); err == nil {
		fmt.Printf("  %s %-25s %s\n", successStyle.Render("✔"), "hyprpolkitagent", dimStyle.Render("Polkit authentication agent found"))
	} else if _, err := os.Stat("/usr/lib/hyprpolkitagent/hyprpolkitagent"); err == nil {
		fmt.Printf("  %s %-25s %s\n", successStyle.Render("✔"), "hyprpolkitagent", dimStyle.Render("Found at /usr/lib/hyprpolkitagent"))
	} else {
		fmt.Printf("  %s %-25s %s\n", warnStyle.Render("!"), "hyprpolkitagent", dimStyle.Render("Missing official polkit agent"))
		*issues++
	}

	// Check xdg-desktop-portal-hyprland
	if _, err := os.Stat("/usr/lib/xdg-desktop-portal-hyprland"); err == nil {
		fmt.Printf("  %s %-25s %s\n", successStyle.Render("✔"), "xdg-desktop-portal-hyprland", dimStyle.Render("Found portal backend"))
	} else {
		fmt.Printf("  %s %-25s %s\n", warnStyle.Render("!"), "xdg-desktop-portal-hyprland", dimStyle.Render("Missing portal backend"))
		*issues++
	}
}

func checkUserGroups(issues *int) {
	u, err := user.Current()
	if err != nil {
		fmt.Printf("  %s Failed to query current user\n", warnStyle.Render("!"))
		return
	}

	gids, err := u.GroupIds()
	if err != nil {
		fmt.Printf("  %s Failed to list user groups\n", warnStyle.Render("!"))
		return
	}

	groupNames := make(map[string]bool)
	for _, gid := range gids {
		if g, err := user.LookupGroupId(gid); err == nil {
			groupNames[g.Name] = true
		}
	}

	for _, reqGroup := range []string{"video", "input"} {
		if groupNames[reqGroup] {
			fmt.Printf("  %s User is in '%s' group\n", successStyle.Render("✔"), reqGroup)
		} else {
			fmt.Printf("  %s User is not in '%s' group %s\n", warnStyle.Render("!"), reqGroup, dimStyle.Render("(run: sudo usermod -aG "+reqGroup+" "+u.Username+")"))
		}
	}
}

func checkConfigs(issues *int) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	hyprConfig := filepath.Join(home, ".config", "hypr", "hyprland.lua")
	if _, err := os.Stat(hyprConfig); err == nil {
		fmt.Printf("  %s %-25s %s\n", successStyle.Render("✔"), "hyprland.lua", dimStyle.Render("Deployed in ~/.config/hypr/"))
	} else {
		fmt.Printf("  %s %-25s %s\n", warnStyle.Render("!"), "hyprland.lua", dimStyle.Render("Not deployed yet"))
		*issues++
	}
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
