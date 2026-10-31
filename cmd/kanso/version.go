package main

import (
	"fmt"
	"runtime"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
	commit  = "dev"
	date    = "2026-10-31"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Kanso version and build details",
	Run: func(cmd *cobra.Command, args []string) {
		title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7aa2f7")).Render("Kanso")
		v := lipgloss.NewStyle().Foreground(lipgloss.Color("#9ece6a")).Render("v" + version)
		dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#565f89"))

		fmt.Printf("%s %s\n", title, v)
		fmt.Printf("  %s %s\n", dim.Render("commit:"), commit)
		fmt.Printf("  %s %s\n", dim.Render("built: "), date)
		fmt.Printf("  %s %s (%s/%s)\n", dim.Render("runtime:"), runtime.Version(), runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
