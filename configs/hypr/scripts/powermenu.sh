#!/usr/bin/env bash
# Kanso Minimal Power & Session Menu
# Uses Rofi dmenu to prompt for Lock, Logout, Shutdown, or Reboot.

set -e

lock="󰌾  Lock"
logout="󰍃  Logout"
reboot="󰜉  Reboot"
shutdown="󰐥  Shutdown"

options="$lock\n$logout\n$reboot\n$shutdown"

# Open Rofi in dmenu mode with 4 rows and case-insensitive matching
chosen=$(echo -e "$options" | rofi -dmenu -i -p "Session" -lines 4 2>/dev/null)

case "$chosen" in
    "$lock")
        if command -v hyprlock &>/dev/null; then
            hyprlock
        fi
        ;;
    "$logout")
        hyprctl dispatch exit
        ;;
    "$reboot")
        systemctl reboot
        ;;
    "$shutdown")
        systemctl poweroff
        ;;
    *)
        exit 0
        ;;
esac
