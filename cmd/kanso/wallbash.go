package main

import (
	"bufio"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
)

// WallbashPalette stores the 4 primary colors, contrast text, and 9-step shade curves
type WallbashPalette struct {
	Mode   string // "dark" or "light"
	Pry    [4]colorful.Color
	Txt    [4]colorful.Color
	Shades [4][9]colorful.Color
}

// Standard relative luminance (ITU-R BT.709)
func colorLuminance(c colorful.Color) float64 {
	return 0.2126*c.R + 0.7152*c.G + 0.0722*c.B
}

// Generate a WallbashPalette from an image file
func ExtractPalette(imgPath string, mode string) (*WallbashPalette, error) {
	f, err := os.Open(imgPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open wallpaper %s: %w", imgPath, err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		// Fallback to ImageMagick if native Go decoding fails
		return extractPaletteViaMagick(imgPath, mode)
	}

	b := img.Bounds()
	if b.Dx() == 0 || b.Dy() == 0 {
		return nil, fmt.Errorf("invalid image dimensions")
	}

	// Sample up to 3600 pixels across the image
	stepX := b.Dx() / 60
	stepY := b.Dy() / 60
	if stepX < 1 {
		stepX = 1
	}
	if stepY < 1 {
		stepY = 1
	}

	var sampled []colorful.Color
	totalLum := 0.0

	for y := b.Min.Y; y < b.Max.Y; y += stepY {
		for x := b.Min.X; x < b.Max.X; x += stepX {
			c, _ := colorful.MakeColor(img.At(x, y))
			sampled = append(sampled, c)
			totalLum += colorLuminance(c)
		}
	}

	if len(sampled) == 0 {
		return nil, fmt.Errorf("no pixels sampled from image")
	}

	avgLum := totalLum / float64(len(sampled))

	// Determine light/dark mode
	resolvedMode := "dark"
	switch mode {
	case "light":
		resolvedMode = "light"
	case "dark":
		resolvedMode = "dark"
	default: // "auto"
		if avgLum >= 0.45 {
			resolvedMode = "light"
		} else {
			resolvedMode = "dark"
		}
	}

	// K-Means clustering (k=4) in Lab space for perceptual accuracy
	centroids := kMeansClusters(sampled, 4, 15)

	// Sort centroids by luminance
	sort.Slice(centroids, func(i, j int) bool {
		if resolvedMode == "light" {
			return colorLuminance(centroids[i]) > colorLuminance(centroids[j])
		}
		return colorLuminance(centroids[i]) < colorLuminance(centroids[j])
	})

	var pal WallbashPalette
	pal.Mode = resolvedMode

	for i := 0; i < 4; i++ {
		c := centroids[i%len(centroids)]
		// Ensure background has enough contrast
		if i == 0 {
			h, s, v := c.Hsv()
			if resolvedMode == "dark" {
				if v > 0.25 {
					v = 0.14
				}
				if s > 0.4 {
					s = 0.3
				}
				c = colorful.Hsv(h, s, v)
			} else {
				if v < 0.85 {
					v = 0.94
				}
				if s > 0.25 {
					s = 0.15
				}
				c = colorful.Hsv(h, s, v)
			}
		}
		pal.Pry[i] = c

		// Contrast text calculation
		lum := colorLuminance(c)
		if lum < 0.45 {
			pal.Txt[i] = colorful.Color{R: 0.94, G: 0.96, B: 0.99} // Crisp light
		} else {
			pal.Txt[i] = colorful.Color{R: 0.09, G: 0.10, B: 0.14} // Deep dark
		}

		// 9-step saturation/brightness curve
		h, s, _ := c.Hsv()
		for j := 0; j < 9; j++ {
			var bri, sat float64
			factor := float64(j+1) / 10.0
			if resolvedMode == "dark" {
				bri = 0.20 + (factor * 0.75) // 0.275 to 0.95
				sat = s * (0.5 + factor*0.5)
			} else {
				bri = 0.95 - (factor * 0.75) // 0.875 to 0.20
				sat = s * (0.4 + factor*0.6)
			}
			pal.Shades[i][j] = colorful.Hsv(h, math.Min(1.0, math.Max(0.0, sat)), math.Min(1.0, math.Max(0.0, bri)))
		}
	}

	return &pal, nil
}

// Fallback to ImageMagick if native decode fails
func extractPaletteViaMagick(imgPath, mode string) (*WallbashPalette, error) {
	if _, err := exec.LookPath("magick"); err != nil {
		return nil, fmt.Errorf("native decode failed and ImageMagick (magick) is not installed")
	}

	cmd := exec.Command("magick", imgPath, "-depth", "8", "-fuzz", "70%", "+dither", "-kmeans", "4", "-format", "%c", "histogram:info:")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("imagemagick extraction failed: %w", err)
	}

	var cols []colorful.Color
	reHex := regexp.MustCompile(`#([0-9a-fA-F]{6})`)
	matches := reHex.FindAllStringSubmatch(string(out), -1)
	for _, m := range matches {
		if c, err := colorful.Hex("#" + m[1]); err == nil {
			cols = append(cols, c)
		}
	}

	if len(cols) == 0 {
		return nil, fmt.Errorf("no colors extracted from imagemagick")
	}

	var pal WallbashPalette
	pal.Mode = mode
	if pal.Mode == "auto" || pal.Mode == "" {
		pal.Mode = "dark"
	}

	for i := 0; i < 4; i++ {
		c := cols[i%len(cols)]
		pal.Pry[i] = c
		if colorLuminance(c) < 0.45 {
			pal.Txt[i] = colorful.Color{R: 0.94, G: 0.96, B: 0.99}
		} else {
			pal.Txt[i] = colorful.Color{R: 0.09, G: 0.10, B: 0.14}
		}
		h, s, _ := c.Hsv()
		for j := 0; j < 9; j++ {
			factor := float64(j+1) / 10.0
			pal.Shades[i][j] = colorful.Hsv(h, s, 0.20+factor*0.75)
		}
	}

	return &pal, nil
}

// Simple K-Means clustering algorithm
func kMeansClusters(points []colorful.Color, k int, iterations int) []colorful.Color {
	if len(points) <= k {
		return points
	}

	centroids := make([]colorful.Color, k)
	r := rand.New(rand.NewSource(42))
	for i := 0; i < k; i++ {
		centroids[i] = points[r.Intn(len(points))]
	}

	for iter := 0; iter < iterations; iter++ {
		clusters := make([][]colorful.Color, k)
		for _, p := range points {
			bestDist := math.MaxFloat64
			bestIdx := 0
			for i, c := range centroids {
				dist := colorDistance(p, c)
				if dist < bestDist {
					bestDist = dist
					bestIdx = i
				}
			}
			clusters[bestIdx] = append(clusters[bestIdx], p)
		}

		for i := 0; i < k; i++ {
			if len(clusters[i]) == 0 {
				centroids[i] = points[r.Intn(len(points))]
				continue
			}
			var sumR, sumG, sumB float64
			for _, p := range clusters[i] {
				sumR += p.R
				sumG += p.G
				sumB += p.B
			}
			n := float64(len(clusters[i]))
			centroids[i] = colorful.Color{R: sumR / n, G: sumG / n, B: sumB / n}
		}
	}

	return centroids
}

func colorDistance(c1, c2 colorful.Color) float64 {
	dr := c1.R - c2.R
	dg := c1.G - c2.G
	db := c1.B - c2.B
	return dr*dr + dg*dg + db*db
}

// Load static .dcol theme overrides (e.g. themes/Graphite Mono/theme.dcol)
func LoadDcolFile(dcolPath string) (*WallbashPalette, error) {
	f, err := os.Open(dcolPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var pal WallbashPalette
	pal.Mode = "dark"

	valMap := make(map[string]string)
	scanner := bufio.NewScanner(f)
	reAssign := regexp.MustCompile(`^([a-zA-Z0-9_]+)="?([^"\n]*)"?`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if m := reAssign.FindStringSubmatch(line); len(m) > 2 {
			valMap[m[1]] = m[2]
		}
	}

	if mode, ok := valMap["dcol_mode"]; ok {
		pal.Mode = mode
	}

	for i := 1; i <= 4; i++ {
		if hex, ok := valMap[fmt.Sprintf("dcol_pry%d", i)]; ok {
			c, _ := colorful.Hex("#" + hex)
			pal.Pry[i-1] = c
		}
		if hex, ok := valMap[fmt.Sprintf("dcol_txt%d", i)]; ok {
			c, _ := colorful.Hex("#" + hex)
			pal.Txt[i-1] = c
		}
		for j := 1; j <= 9; j++ {
			if hex, ok := valMap[fmt.Sprintf("dcol_%dxa%d", i, j)]; ok {
				c, _ := colorful.Hex("#" + hex)
				pal.Shades[i-1][j-1] = c
			}
		}
	}

	return &pal, nil
}

// Process and deploy all Wallbash templates
func ApplyWallbashTemplates(repoRoot string, pal *WallbashPalette) error {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".config")

	// Template search directories: repo templates, then user overrides
	tplDirs := []string{
		filepath.Join(repoRoot, "configs", "wallbash"),
		filepath.Join(home, ".config", "kanso", "wallbash"),
	}

	// Build token lookup tables
	lookup := make(map[string]string)
	rgbLookup := make(map[string]string)

	lookup["wallbash_mode"] = pal.Mode

	for i := 0; i < 4; i++ {
		pIdx := i + 1
		pHex := strings.TrimPrefix(pal.Pry[i].Hex(), "#")
		tHex := strings.TrimPrefix(pal.Txt[i].Hex(), "#")

		lookup[fmt.Sprintf("wallbash_pry%d", pIdx)] = pHex
		lookup[fmt.Sprintf("wallbash_txt%d", pIdx)] = tHex

		rgbPry := fmt.Sprintf("%d,%d,%d", int(pal.Pry[i].R*255), int(pal.Pry[i].G*255), int(pal.Pry[i].B*255))
		rgbTxt := fmt.Sprintf("%d,%d,%d", int(pal.Txt[i].R*255), int(pal.Txt[i].G*255), int(pal.Txt[i].B*255))

		rgbLookup[fmt.Sprintf("wallbash_pry%d", pIdx)] = rgbPry
		rgbLookup[fmt.Sprintf("wallbash_txt%d", pIdx)] = rgbTxt
		lookup[fmt.Sprintf("wallbash_pry%d_rgb", pIdx)] = rgbPry
		lookup[fmt.Sprintf("wallbash_txt%d_rgb", pIdx)] = rgbTxt

		for j := 0; j < 9; j++ {
			sIdx := j + 1
			sHex := strings.TrimPrefix(pal.Shades[i][j].Hex(), "#")
			sKey := fmt.Sprintf("wallbash_%dxa%d", pIdx, sIdx)
			lookup[sKey] = sHex

			rgbShade := fmt.Sprintf("%d,%d,%d", int(pal.Shades[i][j].R*255), int(pal.Shades[i][j].G*255), int(pal.Shades[i][j].B*255))
			rgbLookup[sKey] = rgbShade
			lookup[sKey+"_rgb"] = rgbShade
		}
	}

	reRGBA := regexp.MustCompile(`<wallbash_([a-zA-Z0-9_]+)_rgba\(([^)]*)\)>`)

	processedFiles := make(map[string]bool)

	for _, tplDir := range tplDirs {
		entries, err := os.ReadDir(tplDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".dcol") {
				continue
			}

			tplPath := filepath.Join(tplDir, entry.Name())
			raw, err := os.ReadFile(tplPath)
			if err != nil {
				continue
			}

			lines := strings.Split(string(raw), "\n")
			if len(lines) == 0 {
				continue
			}

			// First line: TargetPath|PostCommand
			header := strings.TrimSpace(lines[0])
			parts := strings.SplitN(header, "|", 2)
			targetPath := strings.TrimSpace(parts[0])
			var postCmd string
			if len(parts) > 1 {
				postCmd = strings.TrimSpace(parts[1])
			}

			// Expand path variables
			targetPath = strings.ReplaceAll(targetPath, "$HOME", home)
			targetPath = strings.ReplaceAll(targetPath, "${HOME}", home)
			targetPath = strings.ReplaceAll(targetPath, "$confDir", configDir)
			targetPath = strings.ReplaceAll(targetPath, "${confDir}", configDir)

			if targetPath == "" || processedFiles[targetPath] {
				continue
			}
			processedFiles[targetPath] = true

			// Template body (everything after line 1)
			body := strings.Join(lines[1:], "\n")

			// 1. Replace RGBA tokens: <wallbash_..._rgba(alpha)>
			rendered := reRGBA.ReplaceAllStringFunc(body, func(m string) string {
				sub := reRGBA.FindStringSubmatch(m)
				key := "wallbash_" + sub[1]
				alpha := sub[2]
				if rgb, ok := rgbLookup[key]; ok {
					return fmt.Sprintf("rgba(%s,%s)", rgb, alpha)
				}
				return m
			})

			// 2. Replace standard tokens: <wallbash_...>
			for k, v := range lookup {
				rendered = strings.ReplaceAll(rendered, "<"+k+">", v)
			}

			// Write destination file safely
			_ = os.MkdirAll(filepath.Dir(targetPath), 0755)
			if err := os.WriteFile(targetPath, []byte(rendered), 0644); err != nil {
				continue
			}

			// Run post reload command if specified
			if postCmd != "" {
				postCmd = strings.ReplaceAll(postCmd, "$HOME", home)
				postCmd = strings.ReplaceAll(postCmd, "${HOME}", home)
				go func(c string) {
					_ = exec.Command("bash", "-c", c).Run()
				}(postCmd)
			}
		}
	}

	return nil
}
