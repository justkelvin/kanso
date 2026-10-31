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
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(titleStyle.Render("Kanso Hyprland Setup"))

		var confirm bool
		err := huh.NewConfirm().
			Title("Welcome to Kanso. Ready to configure your system?").
			Value(&confirm).
			Run()

		if err != nil {
			return err
		}

		if !confirm {
			fmt.Println("Setup aborted.")
			return nil
		}

		fmt.Println("Proceeding with setup...")
		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
