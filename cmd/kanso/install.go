package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	dryRunFlag bool

	bannerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7aa2f7")).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7aa2f7")).
			Padding(0, 2).
			MarginBottom(1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#414868")).
			Padding(0, 1).
			MarginTop(1).
			MarginBottom(1)
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Run the interactive Kanso installation and configuration wizard",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInstaller(dryRunFlag)
	},
}

func init() {
	installCmd.Flags().BoolVarP(&dryRunFlag, "dry-run", "d", false, "Simulate actions without making changes")
	rootCmd.AddCommand(installCmd)
}

func runInstaller(dryRun bool) error {
	fmt.Println(bannerStyle.Render("Kanso Hyprland Installer"))
	if dryRun {
		fmt.Println(warnStyle.Render("ℹ DRY RUN MODE: No files or packages will be modified.\n"))
	}

	repoRoot, err := findRepoRoot()
	if err != nil {
		return fmt.Errorf("unable to locate kanso repository root: %w", err)
	}

	// 1. Choose installer steps
	var selectedSteps []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select installation steps to run:").
				Options(
					huh.NewOption("Install Arch packages (pacman)", "packages").Selected(true),
					huh.NewOption("Deploy configuration files (with backups)", "dotfiles").Selected(true),
					huh.NewOption("Enable user services (PipeWire & WirePlumber)", "services").Selected(true),
					huh.NewOption("Configure TTY1 autostart (~/.bash_profile)", "bash").Selected(true),
				).
				Value(&selectedSteps),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	if len(selectedSteps) == 0 {
		fmt.Println("No steps selected. Exiting.")
		return nil
	}

	stepMap := make(map[string]bool)
	for _, step := range selectedSteps {
		stepMap[step] = true
	}

	// Step 1: Packages
	if stepMap["packages"] {
		fmt.Println(sectionStyle.Render("Step 1: Installing Arch Packages"))
		if err := installArchPackages(repoRoot, dryRun); err != nil {
			fmt.Printf("  %s Failed installing packages: %v\n", errorStyle.Render("✖"), err)
			var proceed bool
			_ = huh.NewConfirm().
				Title("Package installation encountered an issue. Proceed with remaining steps?").
				Value(&proceed).
				Run()
			if !proceed {
				return err
			}
		}
	}

	home, _ := os.UserHomeDir()
	sessionBackupDir := filepath.Join(home, ".config", "kanso", "backups", time.Now().Format("2006-01-02_150405"))
	backedUpCount := 0

	// Step 2: Dotfiles
	if stepMap["dotfiles"] {
		fmt.Println(sectionStyle.Render("Step 2: Deploying Desktop Configurations"))
		if err := deployAllConfigs(repoRoot, sessionBackupDir, &backedUpCount, dryRun); err != nil {
			return fmt.Errorf("dotfile deployment failed: %w", err)
		}
	}

	// Step 3: Bash / TTY1 autostart
	if stepMap["bash"] {
		fmt.Println(sectionStyle.Render("Step 3: Configuring Shell & TTY Autostart"))
		if err := deployBashConfigs(repoRoot, sessionBackupDir, &backedUpCount, dryRun); err != nil {
			return fmt.Errorf("shell configuration failed: %w", err)
		}
	}

	// Step 4: Systemd User Services
	if stepMap["services"] {
		fmt.Println(sectionStyle.Render("Step 4: Activating User Services"))
		if err := enableUserServices(dryRun); err != nil {
			fmt.Printf("  %s Service activation note: %v\n", warnStyle.Render("!"), err)
		}
	}

	// Finish summary
	fmt.Println()
	fmt.Println(successStyle.Render("✔ Kanso setup completed successfully!"))
	if backedUpCount > 0 {
		relBackup, _ := filepath.Rel(home, sessionBackupDir)
		fmt.Printf("  %s Preserved %d existing configuration file(s) in:\n    %s\n",
			warnStyle.Render("↺"),
			backedUpCount,
			successStyle.Render("~/"+relBackup+"/"),
		)
	}

	// Offer to start Hyprland if in TTY
	if !dryRun && os.Getenv("WAYLAND_DISPLAY") == "" && os.Getenv("DISPLAY") == "" {
		var launch bool
		_ = huh.NewConfirm().
			Title("Would you like to launch Hyprland now?").
			Value(&launch).
			Run()

		if launch {
			fmt.Println("Starting Hyprland...")
			c := exec.Command("Hyprland")
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		}
	}

	return nil
}

// Locate repository root containing packages/ and configs/
func findRepoRoot() (string, error) {
	// 1. Current working directory
	cwd, err := os.Getwd()
	if err == nil && isRepoDir(cwd) {
		return cwd, nil
	}

	// 2. Executable directory
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		if isRepoDir(exeDir) {
			return exeDir, nil
		}
		// Parent of executable dir (e.g. build/ -> repo root)
		parentDir := filepath.Dir(exeDir)
		if isRepoDir(parentDir) {
			return parentDir, nil
		}
	}

	return "", fmt.Errorf("could not find kanso repository containing packages/arch.txt")
}

func isRepoDir(dir string) bool {
	_, err1 := os.Stat(filepath.Join(dir, "packages", "arch.txt"))
	_, err2 := os.Stat(filepath.Join(dir, "configs"))
	return err1 == nil && err2 == nil
}

// Read packages/arch.txt and run pacman -S --needed
func installArchPackages(repoRoot string, dryRun bool) error {
	pkgFile := filepath.Join(repoRoot, "packages", "arch.txt")
	file, err := os.Open(pkgFile)
	if err != nil {
		return err
	}
	defer file.Close()

	var pkgs []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		pkgs = append(pkgs, line)
	}

	if len(pkgs) == 0 {
		fmt.Printf("  %s No packages found in packages/arch.txt\n", warnStyle.Render("!"))
		return nil
	}

	fmt.Printf("  %s Found %d packages in %s\n", dimStyle.Render("•"), len(pkgs), dimStyle.Render(pkgFile))

	if dryRun {
		fmt.Printf("  %s Would run: sudo pacman -S --needed %s\n", successStyle.Render("✔"), strings.Join(pkgs[:min(5, len(pkgs))], " ")+"...")
		return nil
	}

	var confirmPacman bool
	err = huh.NewConfirm().
		Title(fmt.Sprintf("Install %d packages via 'sudo pacman -S --needed'?", len(pkgs))).
		Value(&confirmPacman).
		Run()

	if err != nil || !confirmPacman {
		fmt.Printf("  %s Package installation skipped by user\n", warnStyle.Render("○"))
		return nil
	}

	args := append([]string{"pacman", "-S", "--needed"}, pkgs...)
	cmd := exec.Command("sudo", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Deploy desktop configs with safe native Go copying and automatic centralized backups
func deployAllConfigs(repoRoot, backupDir string, backedUpCount *int, dryRun bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(home, ".config")

	// Target directories to copy
	targets := []struct {
		src string
		dst string
	}{
		{filepath.Join(repoRoot, "configs", "hypr"), filepath.Join(configDir, "hypr")},
		{filepath.Join(repoRoot, "configs", "waybar"), filepath.Join(configDir, "waybar")},
		{filepath.Join(repoRoot, "configs", "kitty"), filepath.Join(configDir, "kitty")},
		{filepath.Join(repoRoot, "configs", "rofi"), filepath.Join(configDir, "rofi")},
		{filepath.Join(repoRoot, "configs", "mako"), filepath.Join(configDir, "mako")},
	}

	for _, target := range targets {
		if err := copyDirectoryWithBackup(target.src, target.dst, backupDir, backedUpCount, dryRun); err != nil {
			return err
		}
	}

	// Copy wallpaper asset to ~/.config/hypr/wallpaper.png
	srcWallpaper := filepath.Join(repoRoot, "assets", "wallpaper.png")
	if _, err := os.Stat(srcWallpaper); err == nil {
		dstWallpaper := filepath.Join(configDir, "hypr", "wallpaper.png")
		if err := safeCopyFile(srcWallpaper, dstWallpaper, backupDir, backedUpCount, dryRun); err != nil {
			return err
		}
	}

	return nil
}

// Deploy .bashrc and .bash_profile to user's home directory
func deployBashConfigs(repoRoot, backupDir string, backedUpCount *int, dryRun bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	bashConfigs := []struct {
		src string
		dst string
	}{
		{filepath.Join(repoRoot, "configs", "bash", ".bash_profile"), filepath.Join(home, ".bash_profile")},
		{filepath.Join(repoRoot, "configs", "bash", ".bashrc"), filepath.Join(home, ".bashrc")},
	}

	for _, b := range bashConfigs {
		if err := safeCopyFile(b.src, b.dst, backupDir, backedUpCount, dryRun); err != nil {
			return err
		}
	}

	return nil
}

// Recursively copies a directory, creating backups in the backup hierarchy for existing files that differ
func copyDirectoryWithBackup(srcDir, dstDir, backupDir string, backedUpCount *int, dryRun bool) error {
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil
	}

	return filepath.Walk(srcDir, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, srcPath)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dstDir, relPath)

		if info.IsDir() {
			if !dryRun {
				return os.MkdirAll(dstPath, 0755)
			}
			return nil
		}

		return safeCopyFile(srcPath, dstPath, backupDir, backedUpCount, dryRun)
	})
}

// Copy a single file safely. Backs up to centralized backup directory if destination exists with different content.
func safeCopyFile(src, dst, backupDir string, backedUpCount *int, dryRun bool) error {
	home, _ := os.UserHomeDir()

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	dstInfo, err := os.Stat(dst)
	if err == nil && !dstInfo.IsDir() {
		// Destination exists. Check if content is identical.
		identical, _ := filesHaveIdenticalContent(src, dst)
		if identical {
			fmt.Printf("  %s %s %s\n", dimStyle.Render("•"), filepath.Base(dst), dimStyle.Render("(up to date)"))
			return nil
		}

		// Destination has different content, preserve in centralized backup directory preserving hierarchy
		relToHome, err := filepath.Rel(home, dst)
		var backupDst string
		if err == nil && !strings.HasPrefix(relToHome, "..") {
			backupDst = filepath.Join(backupDir, relToHome)
		} else {
			backupDst = filepath.Join(backupDir, filepath.Base(dst))
		}

		displayRel := relToHome
		if displayRel == "" {
			displayRel = filepath.Base(dst)
		}

		*backedUpCount++
		if dryRun {
			fmt.Printf("  %s Would backup %s -> %s\n", warnStyle.Render("!"), filepath.Base(dst), dimStyle.Render("~/.config/kanso/backups/.../"+displayRel))
			fmt.Printf("  %s Would update %s\n", successStyle.Render("✔"), dst)
			return nil
		}

		if err := os.MkdirAll(filepath.Dir(backupDst), 0755); err != nil {
			return fmt.Errorf("failed creating backup directory structure: %w", err)
		}

		if err := copyRawFile(dst, backupDst, dstInfo.Mode()); err != nil {
			return fmt.Errorf("failed creating backup of %s: %w", dst, err)
		}
		fmt.Printf("  %s Backed up %s -> %s\n", warnStyle.Render("↺"), filepath.Base(dst), dimStyle.Render("~/.config/kanso/backups/.../"+displayRel))
	}

	if dryRun {
		fmt.Printf("  %s Would deploy %s\n", successStyle.Render("✔"), dst)
		return nil
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// Copy content and preserve file mode (including executable permissions)
	if err := copyRawFile(src, dst, srcInfo.Mode()); err != nil {
		return err
	}

	fmt.Printf("  %s Deployed %s\n", successStyle.Render("✔"), filepath.Base(dst))
	return nil
}

func copyRawFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func filesHaveIdenticalContent(f1, f2 string) (bool, error) {
	b1, err := os.ReadFile(f1)
	if err != nil {
		return false, err
	}
	b2, err := os.ReadFile(f2)
	if err != nil {
		return false, err
	}
	return bytes.Equal(b1, b2), nil
}

// Enable PipeWire and WirePlumber user services
func enableUserServices(dryRun bool) error {
	services := []string{"pipewire.service", "pipewire-pulse.service", "wireplumber.service"}

	if dryRun {
		fmt.Printf("  %s Would enable services: %s\n", successStyle.Render("✔"), strings.Join(services, ", "))
		return nil
	}

	for _, svc := range services {
		cmd := exec.Command("systemctl", "--user", "enable", "--now", svc)
		if out, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("  %s Warning enabling %s: %s\n", warnStyle.Render("!"), svc, strings.TrimSpace(string(out)))
		} else {
			fmt.Printf("  %s Enabled and started %s\n", successStyle.Render("✔"), svc)
		}
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
