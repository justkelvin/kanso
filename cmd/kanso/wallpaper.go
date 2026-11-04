package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type WallpaperItem struct {
	Name          string
	FileName      string
	FullPath      string
	ThumbnailPath string
	ModTime       time.Time
	Size          int64
	IsActive      bool
}

type WallpaperState struct {
	ActiveFile string    `json:"active_file"`
	ActiveName string    `json:"active_name"`
	UpdatedAt  time.Time `json:"updated_at"`
}

var wallpaperCmd = &cobra.Command{
	Use:     "wallpaper",
	Aliases: []string{"wp"},
	Short:   "Manage desktop wallpapers, thumbnail cache, and visual grid picker",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWallpaperList()
	},
}

var wpIndexCmd = &cobra.Command{
	Use:     "index",
	Aliases: []string{"sync", "reindex"},
	Short:   "Scan wallpaper directory and generate square thumbnails in cache",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWallpaperIndex()
	},
}

var wpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available wallpapers and current selection",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWallpaperList()
	},
}

var wpSetCmd = &cobra.Command{
	Use:   "set <name-or-path>",
	Short: "Apply a specific wallpaper by filename or path",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWallpaperSet(args[0])
	},
}

var wpNextCmd = &cobra.Command{
	Use:   "next",
	Short: "Cycle to the next wallpaper in the collection",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWallpaperCycle(1)
	},
}

var wpPrevCmd = &cobra.Command{
	Use:     "prev",
	Aliases: []string{"previous"},
	Short:   "Cycle to the previous wallpaper in the collection",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWallpaperCycle(-1)
	},
}

var wpPickCmd = &cobra.Command{
	Use:     "pick",
	Aliases: []string{"picker", "menu"},
	Short:   "Open the visual Rofi thumbnail grid picker",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWallpaperPicker()
	},
}

func init() {
	wallpaperCmd.AddCommand(wpIndexCmd)
	wallpaperCmd.AddCommand(wpListCmd)
	wallpaperCmd.AddCommand(wpSetCmd)
	wallpaperCmd.AddCommand(wpNextCmd)
	wallpaperCmd.AddCommand(wpPrevCmd)
	wallpaperCmd.AddCommand(wpPickCmd)
	rootCmd.AddCommand(wallpaperCmd)
}

// Directory Helpers

func getWallpaperDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	dir := filepath.Join(home, ".config", "hypr", "wallpapers")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func getThumbnailCacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	dir := filepath.Join(home, ".cache", "kanso", "thumbnails")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func getWallpaperStateFile() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "kanso")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "wallpaper.json")
}

func getActiveWallpaperSymlink() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "hypr", "wallpaper.png")
}

// Read current active wallpaper path from symlink or state file
func getActiveWallpaperPath() string {
	symlink := getActiveWallpaperSymlink()
	if target, err := os.Readlink(symlink); err == nil && target != "" {
		return target
	}

	stateFile := getWallpaperStateFile()
	if data, err := os.ReadFile(stateFile); err == nil {
		var s WallpaperState
		if err := json.Unmarshal(data, &s); err == nil && s.ActiveFile != "" {
			return s.ActiveFile
		}
	}

	return ""
}

// Scans wallpaper directory for images (.png, .jpg, .jpeg, .webp)
func discoverWallpapers() ([]WallpaperItem, error) {
	wpDir := getWallpaperDir()
	if wpDir == "" {
		return nil, fmt.Errorf("could not determine wallpaper directory")
	}

	cacheDir := getThumbnailCacheDir()
	activePath := getActiveWallpaperPath()

	entries, err := os.ReadDir(wpDir)
	if err != nil {
		return nil, fmt.Errorf("unable to read wallpaper directory: %w", err)
	}

	var items []WallpaperItem
	validExts := map[string]bool{
		".png":  true,
		".jpg":  true,
		".jpeg": true,
		".webp": true,
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if !validExts[ext] {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		fullPath := filepath.Join(wpDir, name)
		baseName := strings.TrimSuffix(name, filepath.Ext(name))
		thumbPath := filepath.Join(cacheDir, baseName+".png")

		isActive := false
		if activePath != "" && (activePath == fullPath || filepath.Base(activePath) == name) {
			isActive = true
		}

		items = append(items, WallpaperItem{
			Name:          baseName,
			FileName:      name,
			FullPath:      fullPath,
			ThumbnailPath: thumbPath,
			ModTime:       info.ModTime(),
			Size:          info.Size(),
			IsActive:      isActive,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	return items, nil
}

// Generates a high-quality 240x240 center-cropped square thumbnail
func generateSquareThumbnail(srcPath, dstPath string, size int) error {
	srcStat, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	// Skip if thumbnail exists and is newer than source
	if dstStat, err := os.Stat(dstPath); err == nil {
		if dstStat.ModTime().After(srcStat.ModTime()) {
			return nil
		}
	}

	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	src, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("decode error: %w", err)
	}

	bounds := src.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	if w <= 0 || h <= 0 {
		return fmt.Errorf("invalid dimensions %dx%d", w, h)
	}

	minDim := w
	if h < minDim {
		minDim = h
	}

	xOffset := bounds.Min.X + (w-minDim)/2
	yOffset := bounds.Min.Y + (h-minDim)/2

	dst := image.NewRGBA(image.Rect(0, 0, size, size))

	// Area-averaging downsampler
	for dy := 0; dy < size; dy++ {
		srcY0 := yOffset + (dy*minDim)/size
		srcY1 := yOffset + ((dy+1)*minDim)/size
		if srcY1 <= srcY0 {
			srcY1 = srcY0 + 1
		}
		for dx := 0; dx < size; dx++ {
			srcX0 := xOffset + (dx*minDim)/size
			srcX1 := xOffset + ((dx+1)*minDim)/size
			if srcX1 <= srcX0 {
				srcX1 = srcX0 + 1
			}

			stepX := 1
			if span := srcX1 - srcX0; span > 16 {
				stepX = span / 16
			}
			stepY := 1
			if span := srcY1 - srcY0; span > 16 {
				stepY = span / 16
			}

			var rSum, gSum, bSum, aSum uint64
			var count uint64

			for sy := srcY0; sy < srcY1; sy += stepY {
				for sx := srcX0; sx < srcX1; sx += stepX {
					r, g, b, a := src.At(sx, sy).RGBA()
					rSum += uint64(r)
					gSum += uint64(g)
					bSum += uint64(b)
					aSum += uint64(a)
					count++
				}
			}

			if count > 0 {
				dst.SetRGBA(dx, dy, color.RGBA{
					R: uint8((rSum / count) >> 8),
					G: uint8((gSum / count) >> 8),
					B: uint8((bSum / count) >> 8),
					A: uint8((aSum / count) >> 8),
				})
			}
		}
	}

	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()

	return png.Encode(out, dst)
}

// Applies a wallpaper to the system (symlink + IPC + state + notification)
func applyWallpaper(item WallpaperItem) error {
	// 1. Update active symlink
	symlink := getActiveWallpaperSymlink()
	_ = os.Remove(symlink)
	if err := os.Symlink(item.FullPath, symlink); err != nil {
		// Fallback to copy if symlinking fails
		data, readErr := os.ReadFile(item.FullPath)
		if readErr == nil {
			_ = os.WriteFile(symlink, data, 0644)
		}
	}

	// 2. Send IPC command to Hyprpaper
	ipcSuccess := false
	sockPath, sockErr := getHyprlandSocketPath(".hyprpaper.sock")
	if sockErr == nil {
		conn, err := net.DialTimeout("unix", sockPath, 1500*time.Millisecond)
		if err == nil {
			fmt.Fprintf(conn, "wallpaper ,%s\n", item.FullPath)
			conn.Close()
			ipcSuccess = true
		}
	}

	if !ipcSuccess {
		// Try via hyprctl hyprpaper command
		if exec.Command("hyprctl", "hyprpaper", "wallpaper", fmt.Sprintf(",%s", item.FullPath)).Run() == nil {
			ipcSuccess = true
		}
	}

	// Fallback to swaybg if hyprpaper isn't active
	if !ipcSuccess {
		if exec.Command("pgrep", "swaybg").Run() == nil {
			_ = exec.Command("pkill", "swaybg").Run()
			_ = exec.Command("swaybg", "-i", item.FullPath, "-m", "cover").Start()
			ipcSuccess = true
		}
	}

	// 3. Persist State
	state := WallpaperState{
		ActiveFile: item.FullPath,
		ActiveName: item.Name,
		UpdatedAt:  time.Now(),
	}
	if data, err := json.MarshalIndent(state, "", "  "); err == nil {
		_ = os.WriteFile(getWallpaperStateFile(), data, 0644)
	}

	// 4. Update Theme State & Trigger Wallbash if dynamic mode is active
	themeState := loadThemeState()
	themeState.CurrentWallpaper = item.FullPath
	saveThemeState(themeState)

	if themeState.WallbashMode > 0 {
		modeStr := themeModes[themeState.WallbashMode%len(themeModes)]
		_ = applyWallbashFromWallpaper(item.FullPath, modeStr)
	}

	// 5. Send Desktop Notification
	thumbIcon := item.ThumbnailPath
	if _, err := os.Stat(thumbIcon); err != nil {
		thumbIcon = "preferences-desktop-wallpaper"
	}
	_ = exec.Command("notify-send", "-a", "Kanso", "-i", thumbIcon, "Wallpaper Applied", item.Name).Run()

	return nil
}

// CLI Actions

func runWallpaperIndex() error {
	PrintBanner()
	items, err := discoverWallpapers()
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Printf("  %s %s\n\n", cWarn.Render("!"), cMuted.Render("No wallpapers found in "+getWallpaperDir()))
		fmt.Printf("  %s %s\n\n", cCyan.Render("Tip:"), "Drop images (.png, .jpg, .webp) into "+getWallpaperDir()+" and re-run.")
		return nil
	}

	fmt.Printf("  %s %s\n\n", cCyan.Render("•"), fmt.Sprintf("Scanning %d wallpapers in %s...", len(items), getWallpaperDir()))

	start := time.Now()
	newCount := 0

	for _, item := range items {
		_, err := os.Stat(item.ThumbnailPath)
		isNew := os.IsNotExist(err)

		if err := generateSquareThumbnail(item.FullPath, item.ThumbnailPath, 240); err != nil {
			RenderCheckItem(false, item.Name, err.Error())
		} else {
			if isNew {
				newCount++
				RenderCheckItem(true, item.Name, "generated 240x240 thumbnail")
			}
		}
	}

	elapsed := time.Since(start).Round(time.Millisecond)

	var cardItems []KeyVal
	cardItems = append(cardItems, KeyVal{Key: "Directory", Value: getWallpaperDir()})
	cardItems = append(cardItems, KeyVal{Key: "Total Wallpapers", Value: fmt.Sprintf("%d", len(items))})
	cardItems = append(cardItems, KeyVal{Key: "New Thumbnails", Value: fmt.Sprintf("%d", newCount)})
	cardItems = append(cardItems, KeyVal{Key: "Cache Directory", Value: getThumbnailCacheDir()})
	cardItems = append(cardItems, KeyVal{Key: "Duration", Value: fmt.Sprintf("%v", elapsed)})
	cardItems = append(cardItems, KeyVal{Key: "Status", Badge: BadgeGreen("INDEXED")})

	fmt.Println(RenderCard("Wallpaper indexing summary", cardItems))
	return nil
}

func runWallpaperList() error {
	PrintBanner()
	items, err := discoverWallpapers()
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Printf("  %s %s\n\n", cWarn.Render("!"), cMuted.Render("No wallpapers found in "+getWallpaperDir()))
		return nil
	}

	var cardItems []KeyVal
	for _, item := range items {
		status := cMuted.Render(fmt.Sprintf("%.1f MB", float64(item.Size)/(1024*1024)))
		if item.IsActive {
			cardItems = append(cardItems, KeyVal{Key: item.Name, Value: status, Badge: BadgeGreen("ACTIVE")})
		} else {
			cardItems = append(cardItems, KeyVal{Key: item.Name, Value: status})
		}
	}

	cardItems = append(cardItems, KeyVal{Key: "Collection Path", Value: getWallpaperDir()})
	fmt.Println(RenderCard("Available wallpapers", cardItems))
	return nil
}

func runWallpaperSet(query string) error {
	items, err := discoverWallpapers()
	if err != nil {
		return err
	}

	// 1. Direct file path check
	if _, err := os.Stat(query); err == nil {
		absPath, _ := filepath.Abs(query)
		baseName := strings.TrimSuffix(filepath.Base(absPath), filepath.Ext(absPath))
		cacheDir := getThumbnailCacheDir()
		thumb := filepath.Join(cacheDir, baseName+".png")
		_ = generateSquareThumbnail(absPath, thumb, 240)

		targetItem := WallpaperItem{
			Name:          baseName,
			FileName:      filepath.Base(absPath),
			FullPath:      absPath,
			ThumbnailPath: thumb,
			IsActive:      true,
		}

		if err := applyWallpaper(targetItem); err != nil {
			return err
		}
		PrintBanner()
		var cardItems []KeyVal
		cardItems = append(cardItems, KeyVal{Key: "Wallpaper", Value: targetItem.Name})
		cardItems = append(cardItems, KeyVal{Key: "Source", Value: targetItem.FullPath})
		cardItems = append(cardItems, KeyVal{Key: "State", Badge: BadgeGreen("APPLIED")})
		fmt.Println(RenderCard("Wallpaper updated", cardItems))
		return nil
	}

	// 2. Search by name in discovered collection
	var matched *WallpaperItem
	queryLower := strings.ToLower(query)
	for i := range items {
		if strings.ToLower(items[i].Name) == queryLower || strings.ToLower(items[i].FileName) == queryLower {
			matched = &items[i]
			break
		}
	}

	// Fuzzy substring fallback
	if matched == nil {
		for i := range items {
			if strings.Contains(strings.ToLower(items[i].Name), queryLower) {
				matched = &items[i]
				break
			}
		}
	}

	if matched == nil {
		return fmt.Errorf("wallpaper matching '%s' not found in %s", query, getWallpaperDir())
	}

	// Ensure thumbnail exists
	_ = generateSquareThumbnail(matched.FullPath, matched.ThumbnailPath, 240)

	if err := applyWallpaper(*matched); err != nil {
		return err
	}

	PrintBanner()
	var cardItems []KeyVal
	cardItems = append(cardItems, KeyVal{Key: "Wallpaper", Value: matched.Name})
	cardItems = append(cardItems, KeyVal{Key: "File", Value: matched.FileName})
	cardItems = append(cardItems, KeyVal{Key: "Path", Value: matched.FullPath})
	cardItems = append(cardItems, KeyVal{Key: "State", Badge: BadgeGreen("APPLIED")})
	fmt.Println(RenderCard("Wallpaper updated", cardItems))
	return nil
}

func runWallpaperCycle(direction int) error {
	items, err := discoverWallpapers()
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return fmt.Errorf("no wallpapers found in %s", getWallpaperDir())
	}

	activeIndex := -1
	for i, item := range items {
		if item.IsActive {
			activeIndex = i
			break
		}
	}

	var nextIndex int
	if activeIndex == -1 {
		nextIndex = 0
	} else {
		nextIndex = (activeIndex + direction) % len(items)
		if nextIndex < 0 {
			nextIndex = len(items) - 1
		}
	}

	target := items[nextIndex]
	_ = generateSquareThumbnail(target.FullPath, target.ThumbnailPath, 240)

	if err := applyWallpaper(target); err != nil {
		return err
	}

	PrintBanner()
	var cardItems []KeyVal
	cardItems = append(cardItems, KeyVal{Key: "Wallpaper", Value: target.Name})
	cardItems = append(cardItems, KeyVal{Key: "Index", Value: fmt.Sprintf("%d of %d", nextIndex+1, len(items))})
	cardItems = append(cardItems, KeyVal{Key: "File", Value: target.FileName})
	cardItems = append(cardItems, KeyVal{Key: "State", Badge: BadgeGreen("APPLIED")})
	fmt.Println(RenderCard("Wallpaper cycled", cardItems))
	return nil
}

func runWallpaperPicker() error {
	items, err := discoverWallpapers()
	if err != nil {
		return err
	}

	if len(items) == 0 {
		_ = exec.Command("notify-send", "-a", "Kanso", "Wallpaper Picker", "No wallpapers found in "+getWallpaperDir()).Run()
		return fmt.Errorf("no wallpapers found in %s", getWallpaperDir())
	}

	// Ensure thumbnails are generated
	for _, item := range items {
		_ = generateSquareThumbnail(item.FullPath, item.ThumbnailPath, 240)
	}

	// Build Rofi dmenu input format: "DisplayName\0icon\x1fThumbPath\x1finfo\x1fFullPath\n"
	var buf bytes.Buffer
	for _, item := range items {
		prefix := ""
		if item.IsActive {
			prefix = "✓ "
		}
		buf.WriteString(fmt.Sprintf("%s%s\x00icon\x1f%s\x1finfo\x1f%s\n", prefix, item.Name, item.ThumbnailPath, item.FullPath))
	}

	// Find wallpaper-picker.rasi theme
	home, _ := os.UserHomeDir()
	themePath := filepath.Join(home, ".config", "rofi", "wallpaper-picker.rasi")
	if _, err := os.Stat(themePath); err != nil {
		// Fallback to repo path if run from dev
		repoRoot, repoErr := findRepoRoot()
		if repoErr == nil {
			themePath = filepath.Join(repoRoot, "configs", "rofi", "wallpaper-picker.rasi")
		}
	}

	cmdArgs := []string{"-dmenu", "-i", "-p", "Wallpapers", "-show-icons"}
	if _, err := os.Stat(themePath); err == nil {
		cmdArgs = append(cmdArgs, "-theme", themePath)
	}

	rofiCmd := exec.Command("rofi", cmdArgs...)
	rofiCmd.Stdin = &buf
	var out bytes.Buffer
	rofiCmd.Stdout = &out

	if err := rofiCmd.Run(); err != nil {
		// User pressed ESC or cancelled Rofi
		return nil
	}

	selected := strings.TrimSpace(out.String())
	if selected == "" {
		return nil
	}
	selected = strings.TrimPrefix(selected, "✓ ")

	// Find matching item
	for _, item := range items {
		if item.Name == selected {
			return applyWallpaper(item)
		}
	}

	return nil
}
