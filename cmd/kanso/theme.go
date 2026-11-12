package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

type ThemeState struct {
	CurrentTheme     string `json:"current_theme"`
	WallbashMode     int    `json:"wallbash_mode"` // 0: theme, 1: auto, 2: dark, 3: light
	CurrentWallpaper string `json:"current_wallpaper"`
}

var themeModes = []string{"theme", "auto", "dark", "light"}

var themeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Manage desktop theme presets and dynamic Wallbash coloring",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var themeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available theme presets",
	RunE: func(cmd *cobra.Command, args []string) error {
		themes, err := listAvailableThemes()
		if err != nil {
			return err
		}

		state := loadThemeState()
		fmt.Printf("Available Themes (%d):\n\n", len(themes))
		for _, t := range themes {
			badge := ""
			if t == state.CurrentTheme {
				badge = " " + BadgeGreen("[ACTIVE]")
			}
			fmt.Printf("  • %-20s%s\n", cCyan.Render(t), badge)
		}
		fmt.Println()
		return nil
	},
}

var themeGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get current theme and Wallbash mode",
	Run: func(cmd *cobra.Command, args []string) {
		state := loadThemeState()
		modeStr := themeModes[state.WallbashMode%len(themeModes)]

		card := RenderCard("Active theme state", []KeyVal{
			{Key: "Theme Preset", Value: state.CurrentTheme, Badge: BadgeGreen(state.CurrentTheme)},
			{Key: "Wallbash Mode", Value: fmt.Sprintf("%s (%d)", modeStr, state.WallbashMode)},
			{Key: "Active Wallpaper", Value: state.CurrentWallpaper},
		})
		fmt.Println(card)
	},
}

var themeSetCmd = &cobra.Command{
	Use:   "set [theme_name]",
	Short: "Switch to a specific theme preset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		themeName := args[0]
		return applyTheme(themeName)
	},
}

var themePickCmd = &cobra.Command{
	Use:   "pick",
	Short: "Open Rofi visual theme picker grid",
	RunE: func(cmd *cobra.Command, args []string) error {
		return pickThemeRofi()
	},
}

var themeModeCmd = &cobra.Command{
	Use:   "mode [0|1|2|3|theme|auto|dark|light]",
	Short: "Get or set Wallbash mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		state := loadThemeState()
		if len(args) == 0 {
			modeStr := themeModes[state.WallbashMode%len(themeModes)]
			fmt.Printf("Current Wallbash mode: %s (%d)\n", cCyan.Render(modeStr), state.WallbashMode)
			return nil
		}

		newMode := parseModeString(args[0])
		if newMode < 0 || newMode > 3 {
			return fmt.Errorf("invalid mode '%s': must be 0 (theme), 1 (auto), 2 (dark), or 3 (light)", args[0])
		}

		state.WallbashMode = newMode
		saveThemeState(state)

		modeStr := themeModes[newMode]
		fmt.Printf("Wallbash mode set to: %s (%d)\n", cGreen.Render(modeStr), newMode)

		// Re-apply active theme or wallbash palette
		if newMode == 0 {
			return applyTheme(state.CurrentTheme)
		}

		return applyWallbashFromWallpaper(state.CurrentWallpaper, modeStr)
	},
}

var themeModeMenuCmd = &cobra.Command{
	Use:   "mode-menu",
	Short: "Open Rofi Wallbash mode selector",
	RunE: func(cmd *cobra.Command, args []string) error {
		return pickWallbashModeRofi()
	},
}

var themeWallbashCmd = &cobra.Command{
	Use:   "wallbash [image_path]",
	Short: "Regenerate Wallbash color palette from wallpaper",
	RunE: func(cmd *cobra.Command, args []string) error {
		state := loadThemeState()
		targetImg := state.CurrentWallpaper
		if len(args) > 0 {
			targetImg = args[0]
		}
		if targetImg == "" {
			return fmt.Errorf("no image specified and no active wallpaper found")
		}
		modeStr := themeModes[state.WallbashMode%len(themeModes)]
		return applyWallbashFromWallpaper(targetImg, modeStr)
	},
}

func init() {
	themeCmd.AddCommand(themeListCmd)
	themeCmd.AddCommand(themeGetCmd)
	themeCmd.AddCommand(themeSetCmd)
	themeCmd.AddCommand(themePickCmd)
	themeCmd.AddCommand(themeModeCmd)
	themeCmd.AddCommand(themeModeMenuCmd)
	themeCmd.AddCommand(themeWallbashCmd)
	rootCmd.AddCommand(themeCmd)
}

func loadThemeState() ThemeState {
	home, _ := os.UserHomeDir()
	stateFile := filepath.Join(home, ".config", "kanso", "theme.json")

	defaultState := ThemeState{
		CurrentTheme:     "Tokyo Night",
		WallbashMode:     0,
		CurrentWallpaper: filepath.Join(home, ".config", "hypr", "wallpaper.png"),
	}

	data, err := os.ReadFile(stateFile)
	if err != nil {
		return defaultState
	}

	var state ThemeState
	if err := json.Unmarshal(data, &state); err != nil {
		return defaultState
	}

	if state.CurrentTheme == "" {
		state.CurrentTheme = defaultState.CurrentTheme
	}
	return state
}

func saveThemeState(state ThemeState) {
	home, _ := os.UserHomeDir()
	kansoDir := filepath.Join(home, ".config", "kanso")
	_ = os.MkdirAll(kansoDir, 0755)
	stateFile := filepath.Join(kansoDir, "theme.json")

	data, err := json.MarshalIndent(state, "", "  ")
	if err == nil {
		_ = os.WriteFile(stateFile, data, 0644)
	}
}

func parseModeString(s string) int {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "0", "theme":
		return 0
	case "1", "auto":
		return 1
	case "2", "dark":
		return 2
	case "3", "light":
		return 3
	default:
		return -1
	}
}

func findThemeDir(themeName string) (string, error) {
	home, _ := os.UserHomeDir()
	// 1. User themes in ~/.config/kanso/themes/
	userDir := filepath.Join(home, ".config", "kanso", "themes", themeName)
	if info, err := os.Stat(userDir); err == nil && info.IsDir() {
		return userDir, nil
	}

	// 2. Repo root themes/
	repoRoot, err := findRepoRoot()
	if err == nil {
		repoThemeDir := filepath.Join(repoRoot, "themes", themeName)
		if info, err := os.Stat(repoThemeDir); err == nil && info.IsDir() {
			return repoThemeDir, nil
		}
	}

	return "", fmt.Errorf("theme '%s' not found", themeName)
}

func listAvailableThemes() ([]string, error) {
	themeMap := make(map[string]bool)

	home, _ := os.UserHomeDir()
	userDir := filepath.Join(home, ".config", "kanso", "themes")
	if entries, err := os.ReadDir(userDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				themeMap[e.Name()] = true
			}
		}
	}

	repoRoot, err := findRepoRoot()
	if err == nil {
		repoThemes := filepath.Join(repoRoot, "themes")
		if entries, err := os.ReadDir(repoThemes); err == nil {
			for _, e := range entries {
				if e.IsDir() {
					themeMap[e.Name()] = true
				}
			}
		}
	}

	var themes []string
	for t := range themeMap {
		themes = append(themes, t)
	}
	sort.Strings(themes)

	if len(themes) == 0 {
		return nil, fmt.Errorf("no themes found in themes/ or ~/.config/kanso/themes/")
	}
	return themes, nil
}

// Apply a theme preset
func applyTheme(themeName string) error {
	themeDir, err := findThemeDir(themeName)
	if err != nil {
		return err
	}

	repoRoot, _ := findRepoRoot()
	home, _ := os.UserHomeDir()
	state := loadThemeState()
	state.CurrentTheme = themeName

	// 0. Ensure GTK themes, icons, and cursors are deployed to ~/.local/share/
	_ = ensureThemeAssets(repoRoot)

	// 1. Process theme design files (.theme)
	themeFiles, _ := filepath.Glob(filepath.Join(themeDir, "*.theme"))
	for _, tf := range themeFiles {
		_ = deployThemeDesignFile(tf)
	}

	// 2. Copy Kvantum theme assets if present
	kvantumSrc := filepath.Join(themeDir, "kvantum")
	if info, err := os.Stat(kvantumSrc); err == nil && info.IsDir() {
		kvantumDst := filepath.Join(home, ".config", "Kvantum", "wallbash")
		_ = os.MkdirAll(kvantumDst, 0755)

		if data, err := os.ReadFile(filepath.Join(kvantumSrc, "kvantum.theme")); err == nil {
			// Lines after header
			lines := strings.Split(string(data), "\n")
			body := strings.Join(lines[1:], "\n")
			_ = os.WriteFile(filepath.Join(kvantumDst, "wallbash.svg"), []byte(body), 0644)
		}
		if data, err := os.ReadFile(filepath.Join(kvantumSrc, "kvconfig.theme")); err == nil {
			lines := strings.Split(string(data), "\n")
			body := strings.Join(lines[1:], "\n")
			_ = os.WriteFile(filepath.Join(kvantumDst, "wallbash.kvconfig"), []byte(body), 0644)
		}

		// Activate wallbash theme in Kvantum
		kvMain := filepath.Join(home, ".config", "Kvantum", "kvantum.kvconfig")
		_ = os.WriteFile(kvMain, []byte("[General]\ntheme=wallbash\n"), 0644)
	}

	// Deploy native Lua Hyprland theme if present
	themeLuaSrc := filepath.Join(themeDir, "theme.lua")
	if _, err := os.Stat(themeLuaSrc); err == nil {
		hyprDst := filepath.Join(home, ".config", "hypr", "theme.lua")
		_ = copyFile(themeLuaSrc, hyprDst)
		_ = exec.Command("hyprctl", "reload").Run()
	}

	// 4. Resolve and set default theme wallpaper
	themeWallpaper := resolveThemeWallpaper(themeDir)
	if themeWallpaper != "" {
		state.CurrentWallpaper = themeWallpaper
		setDesktopWallpaper(themeWallpaper)
	}

	// 5. Handle Wallbash modes (0 = static theme, 1-3 = dynamic)
	if state.WallbashMode > 0 && themeWallpaper != "" {
		modeStr := themeModes[state.WallbashMode%len(themeModes)]
		_ = applyWallbashFromWallpaper(themeWallpaper, modeStr)
	} else {
		// In Mode 0: Check if theme has theme.dcol override
		dcolFile := filepath.Join(themeDir, "theme.dcol")
		if _, err := os.Stat(dcolFile); err == nil {
			if pal, err := LoadDcolFile(dcolFile); err == nil && repoRoot != "" {
				_ = ApplyWallbashTemplates(repoRoot, pal)
			}
		}
	}

	// 6. Extract GTK and interface metadata from hypr.theme (ensures rofi & kdeglobals retain icon & font contrast)
	hyprThemeFile := filepath.Join(themeDir, "hypr.theme")
	if data, err := os.ReadFile(hyprThemeFile); err == nil {
		applyInterfaceMetadata(string(data))
	}

	saveThemeState(state)

	// Desktop notification
	_ = exec.Command("notify-send", "-a", "Kanso", "Theme Applied", fmt.Sprintf("%s (Mode: %s)", themeName, themeModes[state.WallbashMode%len(themeModes)])).Run()

	fmt.Printf("  %s Switched to theme: %s\n", cSuccess.Render("✓"), cCyan.Render(themeName))
	return nil
}

// Deploy an individual .theme design file
func deployThemeDesignFile(srcPath string) error {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return nil
	}

	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".config")

	header := strings.TrimSpace(lines[0])
	parts := strings.SplitN(header, "|", 2)
	targetPath := strings.TrimSpace(parts[0])
	var postCmd string
	if len(parts) > 1 {
		postCmd = strings.TrimSpace(parts[1])
	}

	targetPath = strings.ReplaceAll(targetPath, "$HOME", home)
	targetPath = strings.ReplaceAll(targetPath, "${HOME}", home)
	targetPath = strings.ReplaceAll(targetPath, "$confDir", configDir)
	targetPath = strings.ReplaceAll(targetPath, "${confDir}", configDir)

	if targetPath == "" {
		return nil
	}

	body := strings.Join(lines[1:], "\n")
	_ = os.MkdirAll(filepath.Dir(targetPath), 0755)
	if err := os.WriteFile(targetPath, []byte(body), 0644); err != nil {
		return err
	}

	if postCmd != "" {
		postCmd = strings.ReplaceAll(postCmd, "$HOME", home)
		postCmd = strings.ReplaceAll(postCmd, "${HOME}", home)
		go func(c string) {
			_ = exec.Command("bash", "-c", c).Run()
		}(postCmd)
	}

	return nil
}

func applyInterfaceMetadata(hyprThemeContent string) {
	re := regexp.MustCompile(`\$([A-Z_]+)\s*=\s*([^\n\r]+)`)
	matches := re.FindAllStringSubmatch(hyprThemeContent, -1)

	var gtkTheme, iconTheme, cursorTheme, cursorSize, colorScheme string
	cursorSize = "24"
	for _, m := range matches {
		key := m[1]
		val := strings.TrimSpace(m[2])
		switch key {
		case "GTK_THEME":
			gtkTheme = val
		case "ICON_THEME":
			iconTheme = val
		case "CURSOR_THEME":
			cursorTheme = val
		case "CURSOR_SIZE":
			cursorSize = val
		case "COLOR_SCHEME":
			colorScheme = val
		}
	}

	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".config")

	// 1. GNOME Desktop Interface via gsettings
	if gtkTheme != "" {
		_ = exec.Command("gsettings", "set", "org.gnome.desktop.interface", "gtk-theme", gtkTheme).Run()
	}
	if iconTheme != "" {
		_ = exec.Command("gsettings", "set", "org.gnome.desktop.interface", "icon-theme", iconTheme).Run()
	}
	if cursorTheme != "" {
		_ = exec.Command("gsettings", "set", "org.gnome.desktop.interface", "cursor-theme", cursorTheme).Run()
		_ = exec.Command("gsettings", "set", "org.gnome.desktop.interface", "cursor-size", cursorSize).Run()
	}
	if colorScheme != "" {
		_ = exec.Command("gsettings", "set", "org.gnome.desktop.interface", "color-scheme", colorScheme).Run()
	}

	// 2. GTK-3.0 Configuration (~/.config/gtk-3.0/settings.ini)
	gtk3Dir := filepath.Join(configDir, "gtk-3.0")
	_ = os.MkdirAll(gtk3Dir, 0755)
	gtk3Settings := fmt.Sprintf("[Settings]\ngtk-theme-name=%s\ngtk-icon-theme-name=%s\ngtk-cursor-theme-name=%s\ngtk-cursor-theme-size=%s\ngtk-font-name=Cantarell 10\n",
		gtkTheme, iconTheme, cursorTheme, cursorSize)
	_ = os.WriteFile(filepath.Join(gtk3Dir, "settings.ini"), []byte(gtk3Settings), 0644)

	// 3. GTK-2.0 Configuration (~/.gtkrc-2.0)
	gtk2Settings := fmt.Sprintf("gtk-theme-name=\"%s\"\ngtk-icon-theme-name=\"%s\"\ngtk-cursor-theme-name=\"%s\"\ngtk-cursor-theme-size=%s\ngtk-font-name=\"Cantarell 10\"\n",
		gtkTheme, iconTheme, cursorTheme, cursorSize)
	_ = os.WriteFile(filepath.Join(home, ".gtkrc-2.0"), []byte(gtk2Settings), 0644)

	// 4. GTK-4.0 Symlink and ~/.themes legacy compatibility link
	if gtkTheme != "" {
		themeBaseDir := filepath.Join(home, ".local", "share", "themes", gtkTheme)
		gtk4Src := filepath.Join(themeBaseDir, "gtk-4.0")
		gtk4Dst := filepath.Join(configDir, "gtk-4.0")

		if info, err := os.Stat(gtk4Src); err == nil && info.IsDir() {
			_ = os.Remove(gtk4Dst)
			_ = os.Symlink(gtk4Src, gtk4Dst)
		}

		legacyThemesDir := filepath.Join(home, ".themes")
		_ = os.MkdirAll(legacyThemesDir, 0755)
		legacyLink := filepath.Join(legacyThemesDir, gtkTheme)
		_ = os.Remove(legacyLink)
		_ = os.Symlink(themeBaseDir, legacyLink)
	}

	// 5. xsettingsd Configuration (~/.config/xsettingsd/xsettingsd.conf)
	xsettingsDir := filepath.Join(configDir, "xsettingsd")
	_ = os.MkdirAll(xsettingsDir, 0755)
	xsettingsContent := fmt.Sprintf("Net/ThemeName \"%s\"\nNet/IconThemeName \"%s\"\nGtk/CursorThemeName \"%s\"\nGtk/CursorThemeSize %s\n",
		gtkTheme, iconTheme, cursorTheme, cursorSize)
	_ = os.WriteFile(filepath.Join(xsettingsDir, "xsettingsd.conf"), []byte(xsettingsContent), 0644)
	_ = exec.Command("pkill", "-HUP", "xsettingsd").Run()

	// 6. Default Cursor Theme for X11 & Wayland fallback
	if cursorTheme != "" {
		cursorDefault := fmt.Sprintf("[Icon Theme]\nName=Default\nComment=Default Cursor Theme\nInherits=%s\n", cursorTheme)
		dotIconsDefault := filepath.Join(home, ".icons", "default")
		_ = os.MkdirAll(dotIconsDefault, 0755)
		_ = os.WriteFile(filepath.Join(dotIconsDefault, "index.theme"), []byte(cursorDefault), 0644)

		localIconsDefault := filepath.Join(home, ".local", "share", "icons", "default")
		_ = os.MkdirAll(localIconsDefault, 0755)
		_ = os.WriteFile(filepath.Join(localIconsDefault, "index.theme"), []byte(cursorDefault), 0644)

		_ = exec.Command("hyprctl", "setcursor", cursorTheme, cursorSize).Run()
	}

	// 7. KDE Globals (~/.config/kdeglobals)
	kdeGlobalsPath := filepath.Join(configDir, "kdeglobals")
	updateKdeGlobals(kdeGlobalsPath, iconTheme, colorScheme)

	// 8. Update Rofi Theme with active Icon Theme
	if iconTheme != "" {
		rofiThemePath := filepath.Join(configDir, "rofi", "theme.rasi")
		if data, err := os.ReadFile(rofiThemePath); err == nil {
			rContent := string(data)
			reRofiIcon := regexp.MustCompile(`(?s)configuration\s*\{[^}]*icon-theme:[^}]*\}`)
			newCfg := fmt.Sprintf("configuration {\n    icon-theme: \"%s\";\n}", iconTheme)
			if reRofiIcon.MatchString(rContent) {
				rContent = reRofiIcon.ReplaceAllString(rContent, newCfg)
			} else {
				rContent = newCfg + "\n\n" + rContent
			}
			_ = os.WriteFile(rofiThemePath, []byte(rContent), 0644)
		}
	}

	// 9. Reload Waybar to pick up new GTK icon and theme settings
	_ = exec.Command("pkill", "-SIGUSR2", "waybar").Run()
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	_ = os.MkdirAll(filepath.Dir(dst), 0755)
	return os.WriteFile(dst, data, 0644)
}

func updateKdeGlobals(filePath, iconTheme, colorScheme string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return
	}
	content := string(data)

	// Remove any hardcoded ColorScheme under [General] so Qt/KDE doesn't fail looking for missing .colors files
	reGeneralScheme := regexp.MustCompile(`(?m)^ColorScheme=.*$`)
	content = reGeneralScheme.ReplaceAllString(content, "")

	if iconTheme != "" {
		reIcon := regexp.MustCompile(`(?m)^Theme=.*$`)
		if strings.Contains(content, "[Icons]") {
			if reIcon.MatchString(content) {
				content = reIcon.ReplaceAllString(content, "Theme="+iconTheme)
			} else {
				content = strings.Replace(content, "[Icons]", "[Icons]\nTheme="+iconTheme, 1)
			}
		} else {
			content += fmt.Sprintf("\n[Icons]\nTheme=%s\n", iconTheme)
		}
	}

	if !strings.Contains(content, "widgetStyle=kvantum") {
		if strings.Contains(content, "[KDE]") {
			content = strings.Replace(content, "[KDE]", "[KDE]\nwidgetStyle=kvantum", 1)
		} else {
			content += "\n[KDE]\nwidgetStyle=kvantum\n"
		}
	}

	if !strings.Contains(content, "[UiSettings]") {
		content += "\n[UiSettings]\nColorScheme=colors\n"
	} else if !strings.Contains(content, "ColorScheme=colors") {
		content = strings.Replace(content, "[UiSettings]", "[UiSettings]\nColorScheme=colors", 1)
	}

	// Update text and background contrast for Qt apps (Dolphin, etc.)
	var fgNormal, bgNormal string
	if colorScheme == "prefer-light" {
		fgNormal = "25,25,25"
		bgNormal = "255,255,255"
	} else {
		fgNormal = "255,255,255"
		bgNormal = "25,25,25"
	}

	reFg := regexp.MustCompile(`(?m)^ForegroundNormal=.*$`)
	reBg := regexp.MustCompile(`(?m)^BackgroundNormal=.*$`)
	content = reFg.ReplaceAllString(content, "ForegroundNormal="+fgNormal)
	content = reBg.ReplaceAllString(content, "BackgroundNormal="+bgNormal)

	_ = os.WriteFile(filePath, []byte(content), 0644)
}

func ensureThemeAssets(repoRoot string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	themesDst := filepath.Join(home, ".local", "share", "themes")
	iconsDst := filepath.Join(home, ".local", "share", "icons")
	_ = os.MkdirAll(themesDst, 0755)
	_ = os.MkdirAll(iconsDst, 0755)

	var arcsDir string
	if repoRoot != "" {
		arcsDir = filepath.Join(repoRoot, "assets", "arcs")
	}
	if arcsDir == "" || !dirExists(arcsDir) {
		kansoArcs := filepath.Join(home, ".config", "kanso", "arcs")
		if dirExists(kansoArcs) {
			arcsDir = kansoArcs
		}
	}
	if arcsDir == "" || !dirExists(arcsDir) {
		return nil
	}

	archives := []struct {
		file       string
		dst        string
		checkEntry string
	}{
		{"gtk-themes.tar.xz", themesDst, "Graphite-Mono"},
		{"icon-themes.tar.xz", iconsDst, "MacTahoe-dark"},
		{"bibata-cursors.tar.xz", iconsDst, "Bibata-Modern-Classic"},
		{"mactahoe-cursors.tar.xz", iconsDst, "MacTahoe-dark-cursors"},
	}

	for _, a := range archives {
		srcPath := filepath.Join(arcsDir, a.file)
		checkPath := filepath.Join(a.dst, a.checkEntry)
		if _, err := os.Stat(checkPath); os.IsNotExist(err) {
			if _, err := os.Stat(srcPath); err == nil {
				cmd := exec.Command("tar", "-xf", srcPath, "-C", a.dst)
				_ = cmd.Run()
			}
		}
	}
	return nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func isImageFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp"
}

func setDesktopWallpaper(wpPath string) error {
	baseName := strings.TrimSuffix(filepath.Base(wpPath), filepath.Ext(wpPath))
	thumb := filepath.Join(getThumbnailCacheDir(), baseName+".png")
	_ = generateSquareThumbnail(wpPath, thumb, 240)
	item := WallpaperItem{
		Name:          baseName,
		FileName:      filepath.Base(wpPath),
		FullPath:      wpPath,
		ThumbnailPath: thumb,
		IsActive:      true,
	}
	return applyWallpaper(item)
}

func resolveThemeWallpaper(themeDir string) string {
	wallSet := filepath.Join(themeDir, "wall.set")
	if target, err := os.Readlink(wallSet); err == nil {
		if !filepath.IsAbs(target) {
			target = filepath.Join(themeDir, target)
		}
		if _, err := os.Stat(target); err == nil {
			return target
		}
	}

	// Fallback to first image in wallpapers/
	wpDir := filepath.Join(themeDir, "wallpapers")
	if entries, err := os.ReadDir(wpDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && isImageFile(e.Name()) {
				return filepath.Join(wpDir, e.Name())
			}
		}
	}

	return ""
}

func applyWallbashFromWallpaper(imgPath, mode string) error {
	repoRoot, _ := findRepoRoot()

	pal, err := ExtractPalette(imgPath, mode)
	if err != nil {
		return fmt.Errorf("wallbash extraction failed: %w", err)
	}

	if err := ApplyWallbashTemplates(repoRoot, pal); err != nil {
		return fmt.Errorf("applying wallbash templates failed: %w", err)
	}

	return nil
}

func pickThemeRofi() error {
	themes, err := listAvailableThemes()
	if err != nil {
		return err
	}

	state := loadThemeState()
	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".cache", "kanso", "theme_thumbs")
	_ = os.MkdirAll(cacheDir, 0755)

	var sb strings.Builder
	for _, t := range themes {
		tDir, err := findThemeDir(t)
		if err != nil {
			continue
		}
		wp := resolveThemeWallpaper(tDir)
		thumbPath := ""
		if wp != "" {
			thumbPath = filepath.Join(cacheDir, t+".png")
			if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
				_ = generateSquareThumbnail(wp, thumbPath, 200)
			}
		}
		if thumbPath != "" {
			sb.WriteString(fmt.Sprintf("%s\x00icon\x1f%s\n", t, thumbPath))
		} else {
			sb.WriteString(t + "\n")
		}
	}

	rasiPath := filepath.Join(home, ".config", "rofi", "theme-picker.rasi")
	if _, err := os.Stat(rasiPath); err != nil {
		if repoRoot, repoErr := findRepoRoot(); repoErr == nil {
			rasiPath = filepath.Join(repoRoot, "configs", "rofi", "theme-picker.rasi")
		}
	}

	cmd := exec.Command("rofi", "-dmenu", "-theme", rasiPath, "-select", state.CurrentTheme)
	cmd.Stdin = strings.NewReader(sb.String())

	out, err := cmd.Output()
	if err != nil {
		return nil // User pressed Escape or closed Rofi
	}

	chosen := strings.TrimSpace(string(out))
	if chosen != "" {
		return applyTheme(chosen)
	}

	return nil
}

func pickWallbashModeRofi() error {
	home, _ := os.UserHomeDir()
	state := loadThemeState()
	items := []string{
		"theme  •  Preset design (locked)",
		"auto   •  Adaptive wallpaper colors",
		"dark   •  Forced dark wallpaper colors",
		"light  •  Forced light wallpaper colors",
	}

	rasiPath := filepath.Join(home, ".config", "rofi", "wallbash-mode.rasi")
	if _, err := os.Stat(rasiPath); err != nil {
		if repoRoot, repoErr := findRepoRoot(); repoErr == nil {
			rasiPath = filepath.Join(repoRoot, "configs", "rofi", "wallbash-mode.rasi")
		}
	}

	cmd := exec.Command("rofi", "-dmenu", "-theme", rasiPath, "-selected-row", fmt.Sprintf("%d", state.WallbashMode))
	cmd.Stdin = strings.NewReader(strings.Join(items, "\n"))

	out, err := cmd.Output()
	if err != nil {
		return nil // Cancelled
	}

	chosen := strings.TrimSpace(string(out))
	for i, item := range items {
		if chosen == item {
			state.WallbashMode = i
			saveThemeState(state)

			modeStr := themeModes[i]
			_ = exec.Command("notify-send", "-a", "Kanso", "Wallbash Mode", fmt.Sprintf("%s (%d)", modeStr, i)).Run()

			if i == 0 {
				return applyTheme(state.CurrentTheme)
			}
			return applyWallbashFromWallpaper(state.CurrentWallpaper, modeStr)
		}
	}

	return nil
}
