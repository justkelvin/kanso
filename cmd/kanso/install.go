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
	"github.com/spf13/cobra"
)

var dryRunFlag bool

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
	PrintBanner()

	if dryRun {
		fmt.Printf("  %s %s\n\n", cWarn.Render("!"), cCyan.Render("DRY RUN MODE: Simulating setup without writing changes."))
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
				Title("Select installation steps:").
				Options(
					huh.NewOption("Install Arch packages (pacman)", "packages").Selected(true),
					huh.NewOption("Deploy configuration files (with backups)", "dotfiles").Selected(true),
					huh.NewOption("Configure TTY1 autostart (~/.bash_profile)", "bash").Selected(true),
					huh.NewOption("Enable user services (PipeWire & WirePlumber)", "services").Selected(true),
					huh.NewOption("Install Kanso CLI binary (~/.local/bin/kanso)", "binary").Selected(true),
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
		fmt.Println(cMuted.Render("  1. PACKAGES"))
		if err := installArchPackages(repoRoot, dryRun); err != nil {
			RenderCheckItem(false, "Failed installing packages", err.Error())
			var proceed bool
			_ = huh.NewConfirm().
				Title("Package installation encountered an issue. Proceed with remaining steps?").
				Value(&proceed).
				Run()
			if !proceed {
				return err
			}
		}
		fmt.Println()
	}

	home, _ := os.UserHomeDir()
	sessionBackupDir := filepath.Join(home, ".config", "kanso", "backups", time.Now().Format("2006-01-02_150405"))
	backedUpCount := 0

	// Step 2: Dotfiles
	if stepMap["dotfiles"] {
		fmt.Println(cMuted.Render("  2. CONFIGURATION DEPLOYMENT"))
		if err := deployAllConfigs(repoRoot, sessionBackupDir, &backedUpCount, dryRun); err != nil {
			return fmt.Errorf("dotfile deployment failed: %w", err)
		}
		fmt.Println()
	}

	// Step 3: Bash / TTY1 autostart
	if stepMap["bash"] {
		fmt.Println(cMuted.Render("  3. SHELL & TTY AUTOSTART"))
		if err := deployBashConfigs(repoRoot, sessionBackupDir, &backedUpCount, dryRun); err != nil {
			return fmt.Errorf("shell configuration failed: %w", err)
		}
		fmt.Println()
	}

	// Step 4: Systemd User Services
	if stepMap["services"] {
		fmt.Println(cMuted.Render("  4. SYSTEMD USER SERVICES"))
		if err := enableUserServices(dryRun); err != nil {
			RenderCheckItem(false, "Service activation note", err.Error())
		}
		fmt.Println()
	}

	// Step 5: Install CLI binary
	if stepMap["binary"] {
		fmt.Println(cMuted.Render("  5. CLI BINARY INSTALLATION"))
		if err := installKansoBinary(repoRoot, dryRun); err != nil {
			RenderCheckItem(false, "Binary installation failed", err.Error())
		}
		fmt.Println()
	}

	// Backup Summary Card (if backups were made)
	if backedUpCount > 0 {
		relBackup, _ := filepath.Rel(home, sessionBackupDir)
		backupCard := RenderCard("Backup preserved", []KeyVal{
			{Key: "Backup Dir", Value: "~/" + relBackup + "/"},
			{Key: "Preserved", Value: fmt.Sprintf("%d configuration file(s)", backedUpCount)},
			{Key: "Hierarchy", Badge: BadgeGreen("IDENTICAL TO $HOME")},
		})
		fmt.Println(backupCard)
	}

	// Finish Summary Card
	summaryItems := []KeyVal{
		{Key: "Compositor", Value: "Hyprland (Native Lua)"},
		{Key: "Surfaces", Value: "Waybar, Kitty, Rofi, Mako"},
		{Key: "Audio & Power", Value: "PipeWire, WirePlumber, Hypridle"},
		{Key: "Security & Secrets", Value: "Polkit, GNOME Keyring (Secret Service)"},
	}
	if stepMap["binary"] {
		summaryItems = append(summaryItems, KeyVal{Key: "CLI Binary", Value: "~/.local/bin/kanso"})
	}
	summaryItems = append(summaryItems, KeyVal{Key: "State", Badge: BadgeGreen("READY")})
	summaryCard := RenderCard("Installation summary", summaryItems)
	fmt.Println(summaryCard)

	// Check PAM configuration for GNOME Keyring auto-unlock
	if !checkPamKeyringConfigured() {
		printPamNotice()
	}

	// Offer to start Hyprland if in TTY
	if !dryRun && os.Getenv("WAYLAND_DISPLAY") == "" && os.Getenv("DISPLAY") == "" {
		var launch bool
		_ = huh.NewConfirm().
			Title("Launch Hyprland desktop now?").
			Value(&launch).
			Run()

		if launch {
			fmt.Println("Launching Hyprland session via start-hyprland...")
			c := exec.Command("start-hyprland")
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

	return "", fmt.Errorf("could not find kanso repository (please run 'kanso install' from within the cloned repository directory)")
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
		RenderCheckItem(false, "No packages found in packages/arch.txt")
		return nil
	}

	RenderCheckItem(true, fmt.Sprintf("Parsed %d packages from packages/arch.txt", len(pkgs)))

	if dryRun {
		RenderCheckItem(true, "Would run: sudo pacman -S --needed [42 packages...]")
		return nil
	}

	var confirmPacman bool
	err = huh.NewConfirm().
		Title(fmt.Sprintf("Install %d packages via 'sudo pacman -S --needed'?", len(pkgs))).
		Value(&confirmPacman).
		Run()

	if err != nil || !confirmPacman {
		RenderCheckItem(false, "Package installation skipped by user")
		return nil
	}

	// Validate or prompt for sudo password directly before pacman runs
	if err := ensureSudo(); err != nil {
		return fmt.Errorf("sudo authentication failed: %w", err)
	}

	args := append([]string{"pacman", "-S", "--needed"}, pkgs...)
	cmd := exec.Command("sudo", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func ensureSudo() error {
	cmd := exec.Command("sudo", "-v")
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
		{filepath.Join(repoRoot, "configs", "xdg-desktop-portal"), filepath.Join(configDir, "xdg-desktop-portal")},
	}

	for _, target := range targets {
		if err := copyDirectoryWithBackup(target.src, target.dst, backupDir, backedUpCount, dryRun); err != nil {
			return err
		}
	}

	// Deploy default wallpaper to ~/.config/hypr/wallpapers/blue_distortion.png
	srcWallpaper := filepath.Join(repoRoot, "assets", "wallpaper.png")
	if _, err := os.Stat(srcWallpaper); err == nil {
		wpDir := filepath.Join(configDir, "hypr", "wallpapers")
		_ = os.MkdirAll(wpDir, 0755)
		dstWallpaper := filepath.Join(wpDir, "blue_distortion.png")
		if err := safeCopyFile(srcWallpaper, dstWallpaper, backupDir, backedUpCount, dryRun); err != nil {
			return err
		}

		// Ensure active symlink ~/.config/hypr/wallpaper.png points to default wallpaper
		activeSymlink := filepath.Join(configDir, "hypr", "wallpaper.png")
		if !dryRun {
			_ = os.Remove(activeSymlink)
			_ = os.Symlink(dstWallpaper, activeSymlink)

			// Prime thumbnail cache
			cacheDir := filepath.Join(home, ".cache", "kanso", "thumbnails")
			_ = os.MkdirAll(cacheDir, 0755)
			_ = generateSquareThumbnail(dstWallpaper, filepath.Join(cacheDir, "blue_distortion.png"), 240)
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
			RenderCheckItem(true, filepath.Base(dst), "up to date")
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
			RenderCheckItem(true, "Would backup "+filepath.Base(dst), "~/.config/kanso/backups/.../"+displayRel)
			RenderCheckItem(true, "Would update "+filepath.Base(dst))
			return nil
		}

		if err := os.MkdirAll(filepath.Dir(backupDst), 0755); err != nil {
			return fmt.Errorf("failed creating backup directory structure: %w", err)
		}

		if err := copyRawFile(dst, backupDst, dstInfo.Mode()); err != nil {
			return fmt.Errorf("failed creating backup of %s: %w", dst, err)
		}
		RenderCheckItem(true, "Backed up "+filepath.Base(dst), "~/.config/kanso/backups/.../"+displayRel)
	}

	if dryRun {
		RenderCheckItem(true, "Would deploy "+filepath.Base(dst))
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

	RenderCheckItem(true, "Deployed "+filepath.Base(dst))
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

// Enable PipeWire, WirePlumber, and GNOME Keyring user services
func enableUserServices(dryRun bool) error {
	services := []string{
		"pipewire.service",
		"pipewire-pulse.service",
		"wireplumber.service",
		"gnome-keyring-daemon.socket",
	}

	if dryRun {
		for _, svc := range services {
			RenderCheckItem(true, "Would enable and start "+svc)
		}
		return nil
	}

	for _, svc := range services {
		cmd := exec.Command("systemctl", "--user", "enable", "--now", svc)
		if out, err := cmd.CombinedOutput(); err != nil {
			RenderCheckItem(false, "Failed activating "+svc, strings.TrimSpace(string(out)))
		} else {
			RenderCheckItem(true, "Activated and started "+svc)
		}
	}

	return nil
}

// Build and install the standalone kanso CLI binary into ~/.local/bin/kanso
func installKansoBinary(repoRoot string, dryRun bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("unable to resolve user home directory: %w", err)
	}

	binDir := filepath.Join(home, ".local", "bin")
	targetPath := filepath.Join(binDir, "kanso")

	if dryRun {
		RenderCheckItem(true, "Simulated building kanso binary", targetPath)
		return nil
	}

	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", binDir, err)
	}

	// Compile using Go toolchain if available
	if goPath, err := exec.LookPath("go"); err == nil && goPath != "" {
		cmd := exec.Command(goPath, "build", "-ldflags", "-s -w", "-o", targetPath, "./cmd/kanso")
		cmd.Dir = repoRoot
		cmd.Env = os.Environ()
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("go build failed: %s (%w)", strings.TrimSpace(string(out)), err)
		}
	} else {
		// Fallback: Copy current running executable into ~/.local/bin/kanso
		currentExe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("unable to determine running executable: %w", err)
		}
		if err := copyRawFile(currentExe, targetPath, 0755); err != nil {
			return fmt.Errorf("failed to copy executable to %s: %w", targetPath, err)
		}
	}

	if err := os.Chmod(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions on %s: %w", targetPath, err)
	}

	RenderCheckItem(true, "Installed kanso CLI binary", targetPath)

	if !strings.Contains(os.Getenv("PATH"), binDir) {
		fmt.Printf("    %s %s\n", cCyan.Render("Tip:"), cMuted.Render("Run 'source ~/.bashrc' or restart terminal to use 'kanso' anywhere."))
	}

	return nil
}

func checkPamKeyringConfigured() bool {
	data, err := os.ReadFile("/etc/pam.d/login")
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "pam_gnome_keyring.so")
}

func printPamNotice() {
	noticeRows := []KeyVal{
		{Key: "Feature", Value: "GNOME Keyring Auto-Unlock"},
		{Key: "Target File", Value: "/etc/pam.d/login"},
		{Key: "Action Needed", Badge: BadgeYellow("RECOMMENDED")},
	}
	card := RenderCard("Post-Install: PAM Keyring", noticeRows)
	fmt.Println(card)
	fmt.Printf("  %s %s\n", cCyan.Render("Note:"), cFg.Render("To auto-unlock your password keyring on TTY1 login without prompts,"))
	fmt.Printf("        %s\n\n", cFg.Render("add these lines to /etc/pam.d/login (see README.md):"))
	fmt.Printf("        %s\n", cMuted.Render("auth     optional  pam_gnome_keyring.so"))
	fmt.Printf("        %s\n\n", cMuted.Render("session  optional  pam_gnome_keyring.so auto_start"))
}
