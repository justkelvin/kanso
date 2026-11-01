package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run diagnostics on system dependencies and Wayland readiness",
	Run: func(cmd *cobra.Command, args []string) {
		PrintBanner()

		issues := 0

		// 1. Operating System
		fmt.Println(cMuted.Render("  OPERATING SYSTEM"))
		if isArchLinux() {
			RenderCheckItem(true, "Base distribution is Arch Linux")
		} else {
			RenderCheckItem(false, "Non-Arch Linux system detected", "Kanso is optimized for Arch")
			issues++
		}
		fmt.Println()

		// 2. Core Desktop Tools
		fmt.Println(cMuted.Render("  DESKTOP BINARIES"))
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
				RenderCheckItem(true, fmt.Sprintf("%-16s %s", bin.name, cMuted.Render(path)))
			} else {
				RenderCheckItem(false, fmt.Sprintf("%-16s %s", bin.name, cMuted.Render("missing: "+bin.description)))
				issues++
			}
		}
		fmt.Println()

		// 3. Portals and Polkit
		fmt.Println(cMuted.Render("  SESSION INTEGRATION"))
		checkPortalAndPolkit(&issues)
		fmt.Println()

		// 4. User Group Permissions
		fmt.Println(cMuted.Render("  USER GROUPS"))
		checkUserGroups(&issues)
		fmt.Println()

		// 5. Compositor & IPC Readiness
		fmt.Println(cMuted.Render("  COMPOSITOR & IPC"))
		checkHyprlandIPC()
		fmt.Println()

		// Summary Card
		stateBadge := BadgeGreen("PASSED")
		stateDesc := "System is ready for Hyprland"
		if issues > 0 {
			stateBadge = BadgeYellow(fmt.Sprintf("%d WARNING(S)", issues))
			stateDesc = "Run 'kanso install' to resolve dependencies"
		}

		summaryCard := RenderCard("Diagnostic Summary", []KeyVal{
			{Key: "Health Status", Badge: stateBadge},
			{Key: "Action", Value: stateDesc},
			{Key: "Compositor", Value: "Hyprland (Native Lua)"},
		})
		fmt.Println(summaryCard)
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
	if _, err := exec.LookPath("hyprpolkitagent"); err == nil {
		RenderCheckItem(true, "hyprpolkitagent", "polkit authentication agent")
	} else if _, err := os.Stat("/usr/lib/hyprpolkitagent/hyprpolkitagent"); err == nil {
		RenderCheckItem(true, "hyprpolkitagent", "found at /usr/lib/hyprpolkitagent")
	} else {
		RenderCheckItem(false, "hyprpolkitagent", "missing official polkit agent")
		*issues++
	}

	if _, err := os.Stat("/usr/lib/xdg-desktop-portal-hyprland"); err == nil {
		RenderCheckItem(true, "xdg-desktop-portal-hyprland", "wayland portal backend")
	} else {
		RenderCheckItem(false, "xdg-desktop-portal-hyprland", "missing portal backend")
		*issues++
	}
}

func checkUserGroups(issues *int) {
	u, err := user.Current()
	if err != nil {
		RenderCheckItem(false, "user lookup", "failed querying user")
		return
	}

	gids, err := u.GroupIds()
	if err != nil {
		RenderCheckItem(false, "groups lookup", "failed listing groups")
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
			RenderCheckItem(true, fmt.Sprintf("User '%s' is in '%s' group", u.Username, reqGroup))
		} else {
			RenderCheckItem(false, fmt.Sprintf("User '%s' not in '%s' group", u.Username, reqGroup), "sudo usermod -aG "+reqGroup+" "+u.Username)
		}
	}
}

func checkHyprlandIPC() {
	sock1, err1 := getHyprlandSocketPath(".socket.sock")
	if err1 != nil {
		RenderCheckItem(true, "Compositor IPC", "ready (activates automatically when Hyprland starts)")
		return
	}

	sig := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if sig != "" {
		RenderCheckItem(true, "Instance signature", sig)
	}

	RenderCheckItem(true, "IPC Command Socket (Socket 1)", sock1)

	if sock2, err2 := getHyprlandSocketPath(".socket2.sock"); err2 == nil {
		RenderCheckItem(true, "IPC Event Stream (Socket 2)", sock2)
	} else {
		RenderCheckItem(false, "IPC Event Stream (Socket 2)", err2.Error())
	}

	if err := PingHyprlandIPC(); err == nil {
		RenderCheckItem(true, "IPC Ping", "connected and responsive")
	} else {
		RenderCheckItem(false, "IPC Ping", err.Error())
	}
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
