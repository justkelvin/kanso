-- Kanso: Minimal Hyprland master configuration in Lua
-- Modular configuration files located in ~/.config/hypr/

local home = os.getenv("HOME")
local hypr_dir = home .. "/.config/hypr"
package.path = hypr_dir .. "/?.lua;" .. hypr_dir .. "/?/init.lua;" .. package.path

require("monitors")
require("autostart")
require("input")
require("look")
require("animations")
require("keybinds")
require("rules")
