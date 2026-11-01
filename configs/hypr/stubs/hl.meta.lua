---@meta
--- LuaCATS definitions and stubs for Hyprland's native Lua configuration API.
--- This file is used by Lua Language Server (LuaLS) for autocompletion, type checking, and documentation.

---@class hl
---@field dsp hl.dsp Dispatcher functions for binding actions
hl = {}

---@class hl.dsp
---@field window hl.dsp.window Window dispatchers
---@field workspace hl.dsp.workspace Workspace dispatchers
hl.dsp = {}

---@class hl.dsp.window
hl.dsp.window = {}

---Close the active window
---@return any
function hl.dsp.window.close() end

---Toggle fullscreen for the active window
---@param fullscreen_mode? integer Fullscreen mode (0: real, 1: maximize)
---@return any
function hl.dsp.window.fullscreen(fullscreen_mode) end

---Toggle floating state for the active window
---@return any
function hl.dsp.window.toggle_floating() end

---Move window focus in a direction
---@param direction "l"|"r"|"u"|"d"|"left"|"right"|"up"|"down" Direction to move focus
---@return any
function hl.dsp.window.move_focus(direction) end

---Move active window to a workspace
---@param workspace string Workspace name or index
---@return any
function hl.dsp.window.move_to_workspace(workspace) end

---Move active window silently to a workspace
---@param workspace string Workspace name or index
---@return any
function hl.dsp.window.move_to_workspace_silent(workspace) end

---@class hl.dsp.workspace
hl.dsp.workspace = {}

---Switch to a workspace
---@param workspace string Workspace name or number
---@return any
function hl.dsp.workspace.switch(workspace) end

---Toggle a special workspace
---@param name? string Special workspace name (default: "special")
---@return any
function hl.dsp.workspace.toggle_special(name) end

---Execute a shell command via dispatcher
---@param command string Shell command to run
---@return any
function hl.dsp.exec_cmd(command) end

---Execute a generic dispatcher with arguments
---@param dispatcher string Dispatcher name
---@param args? string Arguments
---@return any
function hl.dsp.dispatch(dispatcher, args) end

---Configure Hyprland settings via a structured table
---@param config hl.Config
function hl.config(config) end

---Bind a key combination to an action or function
---@param keys string Key combination (e.g. "SUPER + RETURN", "XF86AudioMute")
---@param action any Dispatcher object, function, or command
---@param flags? table Binding flags
function hl.bind(keys, action, flags) end

---Listen to Hyprland compositor events
---@param event "hyprland.start"|"hyprland.reload"|"workspace.change"|"window.open"|"window.close"|string
---@param callback fun(...) Function called when the event triggers
function hl.on(event, callback) end

---Execute an arbitrary command immediately
---@param command string Shell command to execute
function hl.exec_cmd(command) end

---Dispatch a compositor action immediately
---@param dispatcher string Dispatcher name
---@param args? string Arguments
function hl.dispatch(dispatcher, args) end

---Retrieve active window details
---@return table Window table
function hl.get_active_window() end

---Declare a touchpad gesture
---@param options hl.GestureOptions
function hl.gesture(options) end

---Register a timing curve for animations
---@param name string Name of the curve
---@param options hl.CurveOptions
function hl.curve(name, options) end

---Configure animation behavior for a layout scope
---@param options hl.AnimationOptions
function hl.animation(options) end

---Set an individual configuration option
---@param key string Option path (e.g. "general:gaps_in")
---@param value any Value to set
function hl.set_option(key, value) end

--------------------------------------------------------------------------------
-- API Helper Types
--------------------------------------------------------------------------------

---@class hl.GestureOptions
---@field fingers integer Number of fingers (e.g. 3)
---@field direction "horizontal"|"vertical"|"up"|"down"|"left"|"right" Direction
---@field action "workspace"|"close"|string Action to perform
---@field mods? string Modifier keys

---@class hl.CurveOptions
---@field type "bezier"|"spring" Curve type
---@field points? number[][] Control points for bezier (e.g. {{0.05, 0.9}, {0.1, 1.05}})
---@field mass? number Spring mass
---@field stiffness? number Spring stiffness
---@field dampening? number Spring dampening

---@class hl.AnimationOptions
---@field leaf "global"|"windows"|"windowsIn"|"windowsOut"|"border"|"fade"|"workspaces" Animation target scope
---@field enabled? boolean Whether animation is enabled
---@field speed? number Animation speed
---@field bezier? string Registered curve name
---@field spring? string Registered spring curve name
---@field style? string Animation style (e.g. "slide", "popin", "popin 80%")

--------------------------------------------------------------------------------
-- Configuration Table Types
--------------------------------------------------------------------------------

---@class hl.Config
---@field general? hl.Config.General General layout and appearance settings
---@field decoration? hl.Config.Decoration Window decorations, rounding, blur, shadows
---@field animations? hl.Config.Animations Animation enable switch
---@field input? hl.Config.Input Keyboard, mouse, and touchpad settings
---@field dwindle? hl.Config.Dwindle Dwindle layout settings
---@field master? hl.Config.Master Master layout settings
---@field misc? hl.Config.Misc Miscellaneous behavior settings
---@field monitor? string[] Display output rules
---@field windowrule? string[] Legacy window rules
---@field windowrulev2? string[] Modern window rules v2
---@field env? table<string, string> Environment variables

---@class hl.Config.General
---@field gaps_in? integer Inner gaps between windows
---@field gaps_out? integer Outer gaps between windows and screen edges
---@field border_size? integer Border width in pixels
---@field col? hl.Config.General.Col Border colors
---@field layout? "dwindle"|"master" Layout engine
---@field resize_on_border? boolean Allow resizing windows by dragging borders
---@field allow_tearing? boolean Allow tearing for gaming

---@class hl.Config.General.Col
---@field active_border? string|table Active window border gradient/color
---@field inactive_border? string|table Inactive window border color

---@class hl.Config.Decoration
---@field rounding? integer Corner rounding radius in pixels
---@field rounding_power? number Rounding curve power
---@field active_opacity? number Active window opacity (0.0 - 1.0)
---@field inactive_opacity? number Inactive window opacity (0.0 - 1.0)
---@field shadow? hl.Config.Decoration.Shadow Window shadows
---@field blur? hl.Config.Decoration.Blur Window background blur

---@class hl.Config.Decoration.Shadow
---@field enabled? boolean Enable window drop shadows
---@field range? integer Shadow range/blur radius
---@field render_power? integer Shadow render power
---@field color? integer|string Shadow color

---@class hl.Config.Decoration.Blur
---@field enabled? boolean Enable background blur
---@field size? integer Blur size
---@field passes? integer Blur quality passes
---@field vibrancy? number Blur color vibrancy

---@class hl.Config.Animations
---@field enabled? boolean Enable animations
---@field bezier? table<string, string> Custom bezier curve definitions
---@field animation? string[] Animation rules ("target, enable, speed, curve[, style]")

---@class hl.Config.Input
---@field kb_layout? string Keyboard layout (e.g. "us")
---@field kb_variant? string Keyboard variant
---@field kb_model? string Keyboard model
---@field kb_options? string Keyboard options
---@field follow_mouse? integer Mouse focus mode (0: click, 1: cursor, 2: cursor detached, 3: separate)
---@field sensitivity? number Mouse sensitivity (-1.0 to 1.0)
---@field touchpad? hl.Config.Input.Touchpad Touchpad settings

---@class hl.Config.Input.Touchpad
---@field natural_scroll? boolean Enable reverse/natural scrolling
---@field tap_to_click? boolean Enable tap to click
---@field disable_while_typing? boolean Disable touchpad while typing
---@field middle_button_emulation? boolean Enable 3-finger middle click emulation
---@field clickfinger_behavior? boolean 1/2/3 fingers for left/right/middle click

---@class hl.Config.Dwindle
---@field preserve_split? boolean Preserve split direction when closing windows

---@class hl.Config.Master
---@field new_status? string New window placement ("master", "slave", "inherit")

---@class hl.Config.Misc
---@field disable_hyprland_logo? boolean Disable anime mascot / default wallpaper
---@field disable_splash_rendering? boolean Disable splash text
---@field force_default_wallpaper? integer Default wallpaper mode
---@field vfr? boolean Variable frame rate
---@field vrr? integer Variable refresh rate (0: off, 1: on, 2: fullscreen only)
