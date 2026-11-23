-- Kanso: Minimal Hyprland master configuration in Lua
-- Modular configuration files located in ~/.config/hypr/

local home = os.getenv("HOME")
local source = debug.getinfo(1, "S").source:gsub("^@", "")
local hypr_dir = source:match("(.*/)") or (home .. "/.config/hypr/")
package.path = hypr_dir .. "?.lua;" .. hypr_dir .. "?/init.lua;" .. package.path

-- Environment variables for Qt/GTK consistency
if hl and hl.env then
    hl.env("QT_QPA_PLATFORMTHEME", "qt6ct")
    hl.env("QT_STYLE_OVERRIDE", "kvantum")
    hl.env("QT_WAYLAND_DISABLE_WINDOWDECORATION", "1")
    hl.env("QT_QPA_PLATFORM", "wayland;xcb")
end

require("monitors")
require("autostart")
require("input")
require("look")
-- Load theme configuration directly in Lua
package.loaded["theme"] = nil
local ok, theme = pcall(require, "theme")
if ok and type(theme) == "table" then
    hl.config(theme)
end
require("animations")
require("keybinds")
require("rules")
