package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// UI Theme Colors (Tokyo Night / Daytona Minimal)
var (
	colorCyan   = lipgloss.Color("#7aa2f7")
	colorGreen  = lipgloss.Color("#9ece6a")
	colorYellow = lipgloss.Color("#e0af68")
	colorRed    = lipgloss.Color("#f7768e")
	colorPurple = lipgloss.Color("#bb9af7")
	colorFg     = lipgloss.Color("#c0caf5")
	colorMuted  = lipgloss.Color("#565f89")
	colorBorder = lipgloss.Color("#2e3440")

	cSuccess = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	cGreen   = lipgloss.NewStyle().Foreground(colorGreen)
	cWarn    = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
	cError   = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	cMuted   = lipgloss.NewStyle().Foreground(colorMuted)
	cCyan    = lipgloss.NewStyle().Foreground(colorCyan)
	cFg      = lipgloss.NewStyle().Foreground(colorFg)

	// Uniform width for all Daytona-style cards
	cardWidth = 70

	// ASCII Banner in modern slant style
	asciiArt = "    __ __                        \n" +
		"   / //_/___ _____  _________    \n" +
		"  / ,<  / __ `/ __ \\/ ___/ __ \\   \n" +
		" / /| |/ /_/ / / / (__  ) /_/ /   \n" +
		"/_/ |_|\\__,_/_/ /_/____/\\____/    "
)

// PrintBanner outputs the Kanso ASCII art and subtitle
func PrintBanner() {
	artStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorCyan)

	kanjiStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorPurple)

	subStyle := lipgloss.NewStyle().
		Foreground(colorMuted)

	verStyle := lipgloss.NewStyle().
		Foreground(colorGreen)

	fmt.Println(artStyle.Render(asciiArt))
	fmt.Printf("   %s  %s  %s\n\n",
		kanjiStyle.Render("簡 素"),
		subStyle.Render("·  ARCH LINUX & HYPRLAND  ·"),
		verStyle.Render("v"+version),
	)
}

// KeyVal represents a single row in an info card
type KeyVal struct {
	Key   string
	Value string
	Badge string
}

// RenderCard renders a Daytona-style bordered information card with uniform width
func RenderCard(title string, rows []KeyVal) string {
	keyColWidth := 26

	var sb strings.Builder

	// Header badge inside the card
	badge := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color("#c0caf5")).
		Foreground(lipgloss.Color("#15161e")).
		Padding(0, 1).
		Render(title)

	sb.WriteString(badge + "\n\n")

	for _, r := range rows {
		keyStr := lipgloss.NewStyle().
			Foreground(colorMuted).
			Width(keyColWidth).
			Render(r.Key)

		valStr := lipgloss.NewStyle().
			Foreground(colorFg).
			Render(r.Value)

		if r.Badge != "" {
			valStr = r.Badge
		}

		sb.WriteString(fmt.Sprintf("%s%s\n", keyStr, valStr))
	}

	cardStyle := lipgloss.NewStyle().
		Width(cardWidth).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(1, 2).
		MarginBottom(1)

	return cardStyle.Render(strings.TrimRight(sb.String(), "\n"))
}

// RenderCheckItem outputs a Daytona-style step item (✓ or ✖)
func RenderCheckItem(success bool, text string, detail ...string) {
	icon := cSuccess.Render("✓")
	if !success {
		icon = cError.Render("✖")
	}

	msg := lipgloss.NewStyle().Foreground(colorFg).Render(text)
	if len(detail) > 0 && detail[0] != "" {
		dim := cMuted.Render("(" + detail[0] + ")")
		fmt.Printf("  %s %s %s\n", icon, msg, dim)
	} else {
		fmt.Printf("  %s %s\n", icon, msg)
	}
}

// Badge helpers for running / active / missing states
func BadgeGreen(text string) string {
	return lipgloss.NewStyle().
		Foreground(colorGreen).
		Bold(true).
		Render(text)
}

func BadgeYellow(text string) string {
	return lipgloss.NewStyle().
		Foreground(colorYellow).
		Bold(true).
		Render(text)
}

func BadgeRed(text string) string {
	return lipgloss.NewStyle().
		Foreground(colorRed).
		Bold(true).
		Render(text)
}

func BadgeMuted(text string) string {
	return lipgloss.NewStyle().
		Foreground(colorMuted).
		Render(text)
}
