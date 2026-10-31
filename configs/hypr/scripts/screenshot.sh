#!/usr/bin/env bash
# Kanso Screenshot Helper Script
# Handles area selection, screen freezing, copying, saving, and annotation.

set -e

DIR="${XDG_PICTURES_DIR:-$HOME/Pictures}/Screenshots"
mkdir -p "$DIR"
FILENAME="screenshot_$(date +'%Y-%m-%d_%H%M%S').png"
FILEPATH="$DIR/$FILENAME"

annotate() {
    local target="$1"
    if command -v satty &>/dev/null; then
        satty --filename "$target" --output-filename "$target"
    elif command -v swappy &>/dev/null; then
        swappy -f "$target" -o "$target"
    fi
}

notify() {
    local msg="$1"
    if command -v notify-send &>/dev/null; then
        notify-send -i "$FILEPATH" "Screenshot" "$msg"
    fi
}

case "$1" in
    area)
        # Select area or window, copy + save, open annotation
        GEOM=$(slurp 2>/dev/null)
        if [ -n "$GEOM" ]; then
            grim -g "$GEOM" "$FILEPATH"
            wl-copy < "$FILEPATH"
            notify "Saved to $FILEPATH and copied to clipboard"
            annotate "$FILEPATH"
        fi
        ;;

    area-freeze)
        # Freeze screen first with hyprpicker, select, copy + save, open annotation
        if command -v hyprpicker &>/dev/null; then
            hyprpicker -r -z &>/dev/null &
            PICKER_PID=$!
            sleep 0.15
            GEOM=$(slurp 2>/dev/null)
            kill "$PICKER_PID" 2>/dev/null || true
        else
            GEOM=$(slurp 2>/dev/null)
        fi

        if [ -n "$GEOM" ]; then
            grim -g "$GEOM" "$FILEPATH"
            wl-copy < "$FILEPATH"
            notify "Saved to $FILEPATH and copied to clipboard (frozen)"
            annotate "$FILEPATH"
        fi
        ;;

    all)
        # Capture all monitor outputs, copy + save
        grim "$FILEPATH"
        wl-copy < "$FILEPATH"
        notify "All monitors captured and copied to clipboard"
        ;;

    *)
        echo "Usage: $0 {area|area-freeze|all}"
        exit 1
        ;;
esac
