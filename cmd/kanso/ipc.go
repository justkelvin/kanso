package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Hyprland IPC Models

type HyprActiveWorkspace struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type HyprMonitor struct {
	ID              int                 `json:"id"`
	Name            string              `json:"name"`
	Description     string              `json:"description"`
	Width           int                 `json:"width"`
	Height          int                 `json:"height"`
	RefreshRate     float64             `json:"refreshRate"`
	Scale           float64             `json:"scale"`
	Focused         bool                `json:"focused"`
	ActiveWorkspace HyprActiveWorkspace `json:"activeWorkspace"`
}

type HyprWorkspace struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Monitor string `json:"monitor"`
	Windows int    `json:"windows"`
}

type HyprActiveWindow struct {
	Address   string              `json:"address"`
	Title     string              `json:"title"`
	Class     string              `json:"class"`
	Floating  bool                `json:"floating"`
	PID       int                 `json:"pid"`
	Workspace HyprActiveWorkspace `json:"workspace"`
}

// Locate the Hyprland UNIX socket (.socket.sock or .socket2.sock)
func getHyprlandSocketPath(socketName string) (string, error) {
	sig := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		runtimeDir = fmt.Sprintf("/run/user/%d", os.Getuid())
	}

	// 1. Direct path from environment signature
	if sig != "" {
		sock := filepath.Join(runtimeDir, "hypr", sig, socketName)
		if _, err := os.Stat(sock); err == nil {
			return sock, nil
		}
	}

	// 2. Discover running instance in $XDG_RUNTIME_DIR/hypr/*/
	hyprDir := filepath.Join(runtimeDir, "hypr")
	entries, err := os.ReadDir(hyprDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				candidate := filepath.Join(hyprDir, entry.Name(), socketName)
				if _, err := os.Stat(candidate); err == nil {
					return candidate, nil
				}
			}
		}
	}

	return "", fmt.Errorf("hyprland socket %s not found (is Hyprland running?)", socketName)
}

// Send a synchronous command to Hyprland Socket 1 and read response
func HyprlandRequest(command string) (string, error) {
	sockPath, err := getHyprlandSocketPath(".socket.sock")
	if err != nil {
		return "", err
	}

	conn, err := net.DialTimeout("unix", sockPath, 1500*time.Millisecond)
	if err != nil {
		return "", fmt.Errorf("failed connecting to Hyprland IPC: %w", err)
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))

	if _, err := conn.Write([]byte(command)); err != nil {
		return "", fmt.Errorf("failed writing to Hyprland IPC: %w", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, conn); err != nil {
		return "", fmt.Errorf("failed reading from Hyprland IPC: %w", err)
	}

	return buf.String(), nil
}

// PingHyprlandIPC tests whether Hyprland IPC is reachable
func PingHyprlandIPC() error {
	resp, err := HyprlandRequest("version")
	if err != nil {
		return err
	}
	if !strings.Contains(resp, "Hyprland") && !strings.Contains(resp, "tag") {
		return fmt.Errorf("unexpected Hyprland IPC response: %s", strings.TrimSpace(resp))
	}
	return nil
}

// GetHyprlandMonitors fetches connected monitors over IPC
func GetHyprlandMonitors() ([]HyprMonitor, error) {
	raw, err := HyprlandRequest("j/monitors")
	if err != nil {
		return nil, err
	}

	var monitors []HyprMonitor
	if err := json.Unmarshal([]byte(raw), &monitors); err != nil {
		return nil, fmt.Errorf("failed parsing monitors JSON: %w", err)
	}
	return monitors, nil
}

// GetHyprlandWorkspaces fetches active workspaces over IPC
func GetHyprlandWorkspaces() ([]HyprWorkspace, error) {
	raw, err := HyprlandRequest("j/workspaces")
	if err != nil {
		return nil, err
	}

	var workspaces []HyprWorkspace
	if err := json.Unmarshal([]byte(raw), &workspaces); err != nil {
		return nil, fmt.Errorf("failed parsing workspaces JSON: %w", err)
	}
	return workspaces, nil
}

// GetHyprlandActiveWindow returns details on currently focused window
func GetHyprlandActiveWindow() (*HyprActiveWindow, error) {
	raw, err := HyprlandRequest("j/activewindow")
	if err != nil {
		return nil, err
	}

	var win HyprActiveWindow
	if err := json.Unmarshal([]byte(raw), &win); err != nil {
		return nil, fmt.Errorf("failed parsing active window JSON: %w", err)
	}
	return &win, nil
}
