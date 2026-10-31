package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kanso",
	Short: "Kanso: Minimal Arch Linux & Hyprland manager",
	Long: `Kanso is a minimal, keyboard-driven Arch Linux & Hyprland setup manager.
It provides an interactive installer, system diagnostics, and configuration management.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Guard against running with sudo or as root
		if os.Geteuid() == 0 {
			PrintBanner()
			card := RenderCard("Security Notice", []KeyVal{
				{Key: "Status", Badge: BadgeRed("DO NOT RUN WITH SUDO")},
				{Key: "User", Value: "Kanso must be executed as your regular user"},
				{Key: "Reason", Value: "Configures user dotfiles (~/.config) and user systemd units"},
				{Key: "Elevation", Value: "Sudo will be requested automatically when installing packages"},
			})
			fmt.Println(card)
			os.Exit(1)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		PrintBanner()

		var selectedAction string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Choose an operation:").
					Options(
						huh.NewOption("Install & Configure System (install)", "install"),
						huh.NewOption("Run Diagnostics (doctor)", "doctor"),
						huh.NewOption("System Overview (status)", "status"),
						huh.NewOption("Version Information (version)", "version"),
						huh.NewOption("Exit", "exit"),
					).
					Value(&selectedAction),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}

		switch selectedAction {
		case "install":
			return installCmd.RunE(cmd, []string{})
		case "doctor":
			doctorCmd.Run(cmd, []string{})
		case "status":
			statusCmd.Run(cmd, []string{})
		case "version":
			versionCmd.Run(cmd, []string{})
		case "exit":
			fmt.Println("Goodbye.")
		}

		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
