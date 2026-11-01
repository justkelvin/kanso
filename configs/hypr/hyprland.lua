-- Kanso: Minimal Hyprland master configuration in Lua
-- Modular configuration files located in ~/.config/hypr/

local home = os.getenv("HOME")
local source = debug.getinfo(1, "S").source:gsub("^@", "")
local hypr_dir = source:match("(.*/)") or (home .. "/.config/hypr/")
package.path = hypr_dir .. "?.lua;" .. hypr_dir .. "?/init.lua;" .. package.path

require("monitors")
require("autostart")
require("input")
require("look")
require("animations")
require("keybinds")
require("rules")
