package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7aa2f7")).
			MarginBottom(1)
)

var rootCmd = &cobra.Command{
	Use:   "kanso",
	Short: "Kanso: Minimal Arch Linux & Hyprland manager",
	Long: `Kanso is a minimal, keyboard-driven Arch Linux & Hyprland setup manager.
It provides system diagnostics, automated configuration deployment, and health checks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(titleStyle.Render("Kanso — Minimal Arch Linux & Hyprland"))

		var selectedAction string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("What would you like to do?").
					Options(
						huh.NewOption("Run System Diagnostics (doctor)", "doctor"),
						huh.NewOption("Check System Status (status)", "status"),
						huh.NewOption("Show Version Info (version)", "version"),
						huh.NewOption("Exit", "exit"),
					).
					Value(&selectedAction),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}

		switch selectedAction {
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
