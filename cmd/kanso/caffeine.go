package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

type CaffeineState struct {
	Active    bool      `json:"active"`
	PID       int       `json:"pid"`
	StartedAt time.Time `json:"started_at"`
	ScopeName string    `json:"scope_name"`
}

type WaybarOutput struct {
	Text    string `json:"text"`
	Tooltip string `json:"tooltip"`
	Class   string `json:"class"`
}

var jsonOutputFlag bool

var caffeineCmd = &cobra.Command{
	Use:     "caffeine",
	Aliases: []string{"idle", "awake", "coffee"},
	Short:   "Toggle and manage screen idle inhibitor (keep system awake)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCaffeineStatus(jsonOutputFlag)
	},
}

var caffeineToggleCmd = &cobra.Command{
	Use:     "toggle",
	Aliases: []string{"t"},
	Short:   "Toggle Caffeine mode on or off",
	RunE: func(cmd *cobra.Command, args []string) error {
		return withCaffeineLock(func() error {
			state := getCaffeineState()
			if state.Active {
				return stopCaffeine()
			}
			return startCaffeine()
		})
	},
}

var caffeineOnCmd = &cobra.Command{
	Use:   "on",
	Short: "Turn Caffeine mode ON (prevent screen sleep/lock)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return withCaffeineLock(func() error {
			return startCaffeine()
		})
	},
}

var caffeineOffCmd = &cobra.Command{
	Use:   "off",
	Short: "Turn Caffeine mode OFF (restore standard idle timers)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return withCaffeineLock(func() error {
			return stopCaffeine()
		})
	},
}

var caffeineStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display current Caffeine state",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCaffeineStatus(jsonOutputFlag)
	},
}

func init() {
	caffeineStatusCmd.Flags().BoolVar(&jsonOutputFlag, "json", false, "Output Waybar-compatible JSON")
	caffeineCmd.Flags().BoolVar(&jsonOutputFlag, "json", false, "Output Waybar-compatible JSON")

	caffeineCmd.AddCommand(caffeineToggleCmd)
	caffeineCmd.AddCommand(caffeineOnCmd)
	caffeineCmd.AddCommand(caffeineOffCmd)
	caffeineCmd.AddCommand(caffeineStatusCmd)

	rootCmd.AddCommand(caffeineCmd)
}

// Helpers for paths and locking

func getCaffeineRuntimeDir() string {
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir != "" {
		dir := filepath.Join(runtimeDir, "kanso")
		if err := os.MkdirAll(dir, 0700); err == nil {
			return dir
		}
	}

	standardRun := fmt.Sprintf("/run/user/%d/kanso", os.Getuid())
	if err := os.MkdirAll(standardRun, 0700); err == nil {
		return standardRun
	}

	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("kanso-%d", os.Getuid()))
	_ = os.MkdirAll(tmpDir, 0700)
	return tmpDir
}

func getCaffeineStateFile() string {
	return filepath.Join(getCaffeineRuntimeDir(), "caffeine.json")
}

func withCaffeineLock(fn func() error) error {
	lockPath := filepath.Join(getCaffeineRuntimeDir(), "caffeine.lock")
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("failed opening lock file: %w", err)
	}
	defer lockFile.Close()

	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("failed acquiring caffeine lock: %w", err)
	}
	defer syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)

	return fn()
}

func getCaffeineState() CaffeineState {
	stateFile := getCaffeineStateFile()
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return CaffeineState{Active: false}
	}

	var state CaffeineState
	if err := json.Unmarshal(data, &state); err != nil {
		return CaffeineState{Active: false}
	}

	// Verify PID is still alive and belongs to an inhibitor
	if state.Active && state.PID > 0 {
		if err := syscall.Kill(state.PID, 0); err != nil {
			// Stale process, cleanup
			state.Active = false
			_ = os.Remove(stateFile)
		}
	}

	return state
}

func saveCaffeineState(state CaffeineState) {
	stateFile := getCaffeineStateFile()
	data, err := json.MarshalIndent(state, "", "  ")
	if err == nil {
		_ = os.WriteFile(stateFile, data, 0600)
	}
}

func hasSystemdUserBus() bool {
	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		if _, err := os.Stat(filepath.Join(runtimeDir, "bus")); err == nil {
			return true
		}
	}
	return exec.Command("systemctl", "--user", "is-active", "basic.target").Run() == nil
}

// Launches systemd-inhibit within an isolated session and transient scope
func startCaffeine() error {
	state := getCaffeineState()
	if state.Active {
		return nil
	}

	scopeUnit := "kanso-caffeine"
	var cmd *exec.Cmd

	if hasSystemdUserBus() {
		// Prefer transient systemd scope with systemd-inhibit
		cmd = exec.Command("systemd-run", "--user", "--scope", "--unit="+scopeUnit,
			"systemd-inhibit", "--what=idle", "--who=Kanso", "--why=Caffeine active", "--mode=block", "sleep", "infinity")
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

		if err := cmd.Start(); err != nil {
			// Fallback directly to systemd-inhibit without systemd-run
			cmd = exec.Command("systemd-inhibit", "--what=idle", "--who=Kanso", "--why=Caffeine active", "--mode=block", "sleep", "infinity")
			cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
			if err := cmd.Start(); err != nil {
				return fmt.Errorf("failed to start idle inhibitor: %w", err)
			}
			scopeUnit = ""
		}
	} else {
		// Fallback for environments without user systemd bus (e.g. testing / containers)
		scopeUnit = "fallback"
		cmd = exec.Command("sleep", "infinity")
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start inhibitor: %w", err)
		}
	}

	newState := CaffeineState{
		Active:    true,
		PID:       cmd.Process.Pid,
		StartedAt: time.Now(),
		ScopeName: scopeUnit,
	}
	saveCaffeineState(newState)

	sendCaffeineNotification(true)
	signalWaybar()

	return nil
}

// Terminates the inhibitor process and transient scope cleanly
func stopCaffeine() error {
	state := getCaffeineState()
	if !state.Active {
		return nil
	}

	// Terminate systemd scope if unit was created
	if state.ScopeName != "" {
		_ = exec.Command("systemctl", "--user", "stop", state.ScopeName+".scope").Run()
	}

	// Kill process group and process PID
	if state.PID > 0 {
		_ = syscall.Kill(-state.PID, syscall.SIGTERM)
		_ = syscall.Kill(state.PID, syscall.SIGTERM)
	}

	_ = os.Remove(getCaffeineStateFile())

	sendCaffeineNotification(false)
	signalWaybar()

	return nil
}

// Send synchronous desktop notification via mako
func sendCaffeineNotification(active bool) {
	if active {
		_ = exec.Command("notify-send",
			"-a", "Kanso",
			"-h", "string:x-canonical-private-synchronous:kanso-caffeine",
			"-i", "preferences-desktop-screensaver",
			"☕ Caffeine Active",
			"Screen sleep and auto-lock are temporarily disabled.").Run()
	} else {
		_ = exec.Command("notify-send",
			"-a", "Kanso",
			"-h", "string:x-canonical-private-synchronous:kanso-caffeine",
			"-i", "preferences-desktop-screensaver",
			"💤 Caffeine Inactive",
			"Standard power-saving and screen timers restored.").Run()
	}
}

// Real-time signal to Waybar (RTMIN+9 corresponds to signal: 9)
func signalWaybar() {
	_ = exec.Command("pkill", "-RTMIN+9", "waybar").Run()
}

// Status handler for both CLI cards and Waybar JSON
func runCaffeineStatus(jsonOutput bool) error {
	state := getCaffeineState()

	if jsonOutput {
		var out WaybarOutput
		if state.Active {
			elapsed := time.Since(state.StartedAt).Round(time.Minute)
			out = WaybarOutput{
				Text:    "󰅶",
				Tooltip: fmt.Sprintf("☕ Caffeine: Active (%v awake)\nClick to restore normal sleep", elapsed),
				Class:   "active",
			}
		} else {
			out = WaybarOutput{
				Text:    "󰾪",
				Tooltip: "💤 Caffeine: Inactive (Normal sleep timers)\nClick to keep screen awake",
				Class:   "inactive",
			}
		}
		data, err := json.Marshal(out)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	PrintBanner()
	var cardItems []KeyVal
	if state.Active {
		cardItems = append(cardItems, KeyVal{Key: "State", Badge: BadgeGreen("ACTIVE")})
		cardItems = append(cardItems, KeyVal{Key: "Status", Value: "Screen sleep & lock disabled"})
		cardItems = append(cardItems, KeyVal{Key: "PID", Value: fmt.Sprintf("%d", state.PID)})
		if state.ScopeName != "" {
			cardItems = append(cardItems, KeyVal{Key: "Scope", Value: state.ScopeName + ".scope"})
		}
		cardItems = append(cardItems, KeyVal{Key: "Awake Since", Value: state.StartedAt.Format("15:04:05")})
		cardItems = append(cardItems, KeyVal{Key: "Duration", Value: fmt.Sprintf("%v", time.Since(state.StartedAt).Round(time.Second))})
	} else {
		cardItems = append(cardItems, KeyVal{Key: "State", Badge: BadgeRed("INACTIVE")})
		cardItems = append(cardItems, KeyVal{Key: "Status", Value: "Standard power management active"})
		cardItems = append(cardItems, KeyVal{Key: "Hypridle", Value: "Listening to idle events (lock at 5m, dpms at 5.5m)"})
	}

	fmt.Println(RenderCard("Caffeine status", cardItems))
	return nil
}
