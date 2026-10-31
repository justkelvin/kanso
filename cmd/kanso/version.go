package main

import (
	"fmt"
	"runtime"

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
		PrintBanner()

		card := RenderCard("Build Info", []KeyVal{
			{Key: "Version", Value: "v" + version, Badge: BadgeGreen("v" + version)},
			{Key: "Commit", Value: commit},
			{Key: "Built", Value: date},
			{Key: "Runtime", Value: fmt.Sprintf("%s (%s/%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH)},
			{Key: "Platform", Value: "Arch Linux / Wayland"},
		})

		fmt.Println(card)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
