package configeditor

// GetEmbeddedSchema returns a manually created comprehensive schema
// based on the mt-canvus.ini documentation and example file.
// This schema is manually maintained and should be updated when new settings are added.
func GetEmbeddedSchema() *ConfigSchema {
	schema := NewConfigSchema()

	// ============================================================================
	// ROOT SECTION (no section header)
	// ============================================================================
	// Note: Some settings that appear in root in older documentation are actually
	// in [system] section in the example file. They are placed in [system] below.


	// ============================================================================
	// [system] SECTION
	// ============================================================================

	// canvus-folder
	schema.AddOption(&ConfigOption{
		Section:     "system",
		Key:         "canvus-folder",
		Default:     "",
		Description: "The directory that is opened from the finger menu. DEFAULT: Home directory of the active user in single user mode or folder stored inside content/root in multi-user mode.",
		Type:        ValueTypeFilePath,
	})

	// inactive-timeout
	schema.AddOption(&ConfigOption{
		Section:     "system",
		Key:         "inactive-timeout",
		Default:     "14400",
		Description: "Timeout (in seconds) after which workspace is closed and returned to the welcome screen if the workspace hasn't been interacted with. Setting inactive-timeout=0 disables this feature.",
		Type:        ValueTypeNumber,
	})

	// menu-timeout
	schema.AddOption(&ConfigOption{
		Section:     "system",
		Key:         "menu-timeout",
		Default:     "15",
		Description: "Timeout (in seconds) for expanded main menus. Set to zero for no timeout.",
		Type:        ValueTypeNumber,
	})

	// virtual-keyboard-enabled
	schema.AddOption(&ConfigOption{
		Section:     "system",
		Key:         "virtual-keyboard-enabled",
		Default:     "auto",
		Description: "Should Canvus use a built-in virtual keyboard. Available options: true - Virtual keyboard is opened automatically, false - Virtual keyboard is not opened automatically, the system is expected to have a physical keyboard, auto - Virtual keyboard is used in multi-user mode or when running in Windows 10 Tablet mode.",
		Type:        ValueTypeEnum,
		EnumValues:  []string{"true", "false", "auto"},
	})

	// multi-user-mode-enabled
	schema.AddOption(&ConfigOption{
		Section:     "system",
		Key:         "multi-user-mode-enabled",
		Default:     "auto",
		Description: "Enabling multi-user mode optimizes the application user experience for large, multi-user installations. Disable it for a single user experience on personal devices like laptops. Available options: true - Multi-user mode is enabled. QR code login is preferred, the user interface is optimized for multiple users. false - Multi-user mode is disabled. All local and network disk drives are shown in the volume list instead of just removable USB drives. auto - The system is considered to be in multi-user mode if the application display configuration (screen.xml) contains multiple windows, or if the touch configuration (config.txt) uses MultiTaction displays.",
		Type:        ValueTypeEnum,
		EnumValues:  []string{"true", "false", "auto"},
	})

	// virtual-keyboard-layouts
	schema.AddOption(&ConfigOption{
		Section:     "system",
		Key:         "virtual-keyboard-layouts",
		Default:     "en",
		Description: "A comma separated list of enabled virtual keyboard layouts. The first item in the list is the default layout. Available choices are: en (English), fr (French), ru (Russian)",
		Type:        ValueTypeCommaList,
		EnumValues:  []string{"en", "fr", "ru"},
	})

	// default-canvas
	schema.AddOption(&ConfigOption{
		Section:     "system",
		Key:         "default-canvas",
		Default:     "",
		Description: "URL of the default canvas that can be quickly opened from the welcome screen.",
		Type:        ValueTypeString,
	})

	// lock-config
	schema.AddOption(&ConfigOption{
		Section:     "system",
		Key:         "lock-config",
		Default:     "false",
		Description: "Prevent Canvus client from making any changes to this configuration file. This option is primarily meant for public installations.",
		Type:        ValueTypeBoolean,
		EnumValues:  []string{"true", "false"},
	})

	// show-volumes
	schema.AddOption(&ConfigOption{
		Section:     "system",
		Key:         "show-volumes",
		Default:     "*",
		Description: "A comma separated list of drive letters, which should always be displayed on the side menu if mounted. A value of '*' means all drive letters. This setting is only supported on Windows. DEFAULT (multi-user-mode-enabled=false): * DEFAULT (multi-user-mode-enabled=true): empty",
		Type:        ValueTypeCommaList,
	})

	// codice-server
	schema.AddOption(&ConfigOption{
		Section:     "system",
		Key:         "codice-server",
		Default:     "",
		Description: "Name of the Codice server where all Codice data like the personal folders and Canvas Codice maps are stored. To use these features, configure MT Canvus multisite server by adding a new section for the server, for instance [server:My server], and then setting codice-server=My server. DEFAULT: empty, which disables all Codice server features",
		Type:        ValueTypeString,
	})

	// enable-personal-folders
	schema.AddOption(&ConfigOption{
		Section:     "system",
		Key:         "enable-personal-folders",
		Default:     "true",
		Description: "Enables personal folder -feature that is triggered with Codice markers. See also server/codice-server setting in this file.",
		Type:        ValueTypeBoolean,
		EnumValues:  []string{"true", "false"},
	})

	// enable-codice-folders
	schema.AddOption(&ConfigOption{
		Section:     "system",
		Key:         "enable-codice-folders",
		Default:     "true",
		Description: "Enables Codice folder -feature that is triggered with Codice markers. Also configure system/codice-server to use this feature.",
		Type:        ValueTypeBoolean,
		EnumValues:  []string{"true", "false"},
	})

	// enforce-canvas-password
	schema.AddOption(&ConfigOption{
		Section:     "system",
		Key:         "enforce-canvas-password",
		Default:     "false",
		Description: "Specify if all newly-created local canvases must be protected by a password. This won't affect canvases on a remote server.",
		Type:        ValueTypeBoolean,
		EnumValues:  []string{"true", "false"},
	})

	// admin-info
	schema.AddOption(&ConfigOption{
		Section:     "system",
		Key:         "admin-info",
		Default:     "",
		Description:  "Contact information for someone who can provide assistance when passwords have been lost. New lines can be created using \\n syntax.",
		Type:        ValueTypeString,
	})

	// presentation-control-timeout-seconds
	schema.AddOption(&ConfigOption{
		Section:     "system",
		Key:         "presentation-control-timeout-seconds",
		Default:     "30",
		Description: "Timeout (in seconds) for menu controls in full screen presentation mode.",
		Type:        ValueTypeNumber,
	})

	// max-snapshot-resolution
	schema.AddOption(&ConfigOption{
		Section:     "system",
		Key:         "max-snapshot-resolution",
		Default:     "8192",
		Description: "Max resolution (in pixels) for snapshot images created in MT Canvus.",
		Type:        ValueTypeNumber,
	})

	// snapshot-scale
	schema.AddOption(&ConfigOption{
		Section:     "system",
		Key:         "snapshot-scale",
		Default:     "1.0",
		Description: "When taking snapshots, scale images using this factor. Value 1.0 means that the snapshot is taken in native resolution. Values smaller than 1 make the snapshot smaller and more blurry, while bigger values increase the resolution and makes the picture sharper. Almost always 1.0 is good choice, but 2.0 could also be used for \"hi-dpi\" content. Notice that the snapshot resolution will still be limited by max-snapshot-resolution -setting.",
		Type:        ValueTypeNumber,
	})

	// password-char-display-time
	schema.AddOption(&ConfigOption{
		Section:     "system",
		Key:         "password-char-display-time",
		Default:     "0.5",
		Description: "The time in seconds to display a typed character in a password field before it is obscured.",
		Type:        ValueTypeNumber,
	})

	// min-workspace-width-for-info-panel
	schema.AddOption(&ConfigOption{
		Section:     "system",
		Key:         "min-workspace-width-for-info-panel",
		Default:     "860",
		Description: "The minimum width of a split workspace, below which the information panel is hidden. If 0, the information panel is never hidden.",
		Type:        ValueTypeNumber,
	})

	// minimum-window-size
	schema.AddOption(&ConfigOption{
		Section:     "system",
		Key:         "minimum-window-size",
		Default:     "1024 768",
		Description: "The minimum size the user can set the window to when single user mode is enabled. Set minimum-window-size=0 0 to remove the limit.",
		Type:        ValueTypeString,
	})

	// max-open-files
	schema.AddOption(&ConfigOption{
		Section:     "system",
		Key:         "max-open-files",
		Default:     "4096",
		Description: "The maximum number of files that can be opened by MT Canvus (Linux only).",
		Type:        ValueTypeNumber,
	})

	// max-concurrent-tasks
	schema.AddOption(&ConfigOption{
		Section:     "system",
		Key:         "max-concurrent-tasks",
		Default:     "0",
		Description: "The maximum number of background tasks that can be performed at the same time. DEFAULT: 0 (automatically chosen based on available CPU cores and other factors)",
		Type:        ValueTypeNumber,
	})

	// ============================================================================
	// [smtp] SECTION
	// ============================================================================

	// host
	schema.AddOption(&ConfigOption{
		Section:     "smtp",
		Key:         "host",
		Default:     "",
		Description: "SMTP server hostname.",
		Type:        ValueTypeString,
	})

	// port
	schema.AddOption(&ConfigOption{
		Section:     "smtp",
		Key:         "port",
		Default:     "",
		Description: "SMTP server port.",
		Type:        ValueTypeString,
	})

	// username
	schema.AddOption(&ConfigOption{
		Section:     "smtp",
		Key:         "username",
		Default:     "",
		Description: "SMTP username.",
		Type:        ValueTypeString,
	})

	// password
	schema.AddOption(&ConfigOption{
		Section:     "smtp",
		Key:         "password",
		Default:     "",
		Description: "SMTP password.",
		Type:        ValueTypeString,
	})

	// connection-encryption
	schema.AddOption(&ConfigOption{
		Section:     "smtp",
		Key:         "connection-encryption",
		Default:     "auto",
		Description: "Encryption option for email server connection.",
		Type:        ValueTypeEnum,
		EnumValues:  []string{"auto", "none", "SSL", "TLS", "SSL-ignore-errors"},
	})

	// ignore-proxy
	schema.AddOption(&ConfigOption{
		Section:     "smtp",
		Key:         "ignore-proxy",
		Default:     "false",
		Description: "Should system proxy settings be ignored while connecting to SMTP.",
		Type:        ValueTypeBoolean,
		EnumValues:  []string{"true", "false"},
	})

	// ============================================================================
	// [mail] SECTION
	// ============================================================================

	// sender-name
	schema.AddOption(&ConfigOption{
		Section:     "mail",
		Key:         "sender-name",
		Default:     "Canvus",
		Description: "Sender name of the email",
		Type:        ValueTypeString,
	})

	// sender-address
	schema.AddOption(&ConfigOption{
		Section:     "mail",
		Key:         "sender-address",
		Default:     "noreply@multitaction.com",
		Description: "Sender address of the email",
		Type:        ValueTypeString,
	})

	// subject
	schema.AddOption(&ConfigOption{
		Section:     "mail",
		Key:         "subject",
		Default:     "Content from Canvus",
		Description: "Subject for the emails.",
		Type:        ValueTypeString,
	})

	// max-attachment-size
	schema.AddOption(&ConfigOption{
		Section:     "mail",
		Key:         "max-attachment-size",
		Default:     "32",
		Description: "Maximum email attachment size in megabytes",
		Type:        ValueTypeNumber,
	})

	// default-email-address
	schema.AddOption(&ConfigOption{
		Section:     "mail",
		Key:         "default-email-address",
		Default:     "",
		Description: "Default email address when sending email.",
		Type:        ValueTypeString,
	})

	// allow-user-email
	schema.AddOption(&ConfigOption{
		Section:     "mail",
		Key:         "allow-user-email",
		Default:     "true",
		Description: "Allow users to specify another email address when sending email. Note: Disabling this setting and setting default-email-address to an empty value is not a valid configuration. Canvus will refuse to start if you specify it.",
		Type:        ValueTypeBoolean,
		EnumValues:  []string{"true", "false"},
	})

	// ============================================================================
	// [fixed-workspace:<index>] SECTION (Compound)
	// ============================================================================
	// Mark as compound section
	if section := schema.GetSection("fixed-workspace"); section == nil {
		section = &ConfigSection{
			Name:            "fixed-workspace",
			Description:     "Fixed workspaces that cannot be resized or removed. Multiple workspaces can be defined that cannot be changed while the application is running. Each workspace is defined with the format [fixed-workspace:<index>], where the index (starting at 1) defines the left to right positional order of the workspaces.",
			Options:         []*ConfigOption{},
			IsCompound:      true,
			Pattern:         "fixed-workspace",
			CompoundEntries: make(map[string][]*ConfigOption),
		}
		schema.Sections["fixed-workspace"] = section
	}
	fixedWorkspaceSection := schema.GetSection("fixed-workspace")
	fixedWorkspaceSection.IsCompound = true
	fixedWorkspaceSection.Pattern = "fixed-workspace"

	// size (required for each fixed-workspace)
	schema.AddOption(&ConfigOption{
		Section:     "fixed-workspace",
		Key:         "size",
		Default:     "",
		Description: "The size setting must be specified for each workspace. Format: width height (e.g., 3240 1920)",
		Type:        ValueTypeString,
		IsCompound:  true,
		Pattern:     "fixed-workspace",
	})

	// name
	schema.AddOption(&ConfigOption{
		Section:     "fixed-workspace",
		Key:         "name",
		Default:     "",
		Description: "If specified, this name appears in the workspace label, otherwise the standard \"Workspace n\" naming is used.",
		Type:        ValueTypeString,
		IsCompound:  true,
		Pattern:     "fixed-workspace",
	})

	// view-location
	schema.AddOption(&ConfigOption{
		Section:     "fixed-workspace",
		Key:         "view-location",
		Default:     "0 0",
		Description: "The initial location of the view on the workspace. If 0 0, the default location will be used. NOTE: Values specifying an area outside the view size will be adjusted to ensure the viewport is valid. NOTE: setting either view-location or view-scale will override any other settings that influence the initial position and scale of the view.",
		Type:        ValueTypeString,
		IsCompound:  true,
		Pattern:     "fixed-workspace",
	})

	// view-scale
	schema.AddOption(&ConfigOption{
		Section:     "fixed-workspace",
		Key:         "view-scale",
		Default:     "0",
		Description: "The initial scale of the view on the workspace. If zero, the default scale will be used. NOTE: The maximum scale factor value allowed is 10.",
		Type:        ValueTypeNumber,
		IsCompound:  true,
		Pattern:     "fixed-workspace",
	})

	// ============================================================================
	// [annotation] SECTION
	// ============================================================================

	// eraser-marker-codes
	schema.AddOption(&ConfigOption{
		Section:     "annotation",
		Key:         "eraser-marker-codes",
		Default:     "612, 614",
		Description: "Codice eraser codes. You can assign a new value or add a comma-separated list of values. By default, Codice cards 612 and 614 are defined as Codice eraser cards.",
		Type:        ValueTypeCommaList,
	})

	// enable-canvas-codices
	schema.AddOption(&ConfigOption{
		Section:     "annotation",
		Key:         "enable-canvas-codices",
		Default:     "false",
		Description: "Allows assignment of Codice marker to be used as a shortcut for canvas selection.",
		Type:        ValueTypeBoolean,
		EnumValues:  []string{"true", "false"},
	})

	// eraser-size
	schema.AddOption(&ConfigOption{
		Section:     "annotation",
		Key:         "eraser-size",
		Default:     "60 100",
		Description: "Eraser size used by the rectangular shaped eraser.",
		Type:        ValueTypeString,
	})

	// ============================================================================
	// [welcome-screen] SECTION
	// ============================================================================

	// allow-exit
	schema.AddOption(&ConfigOption{
		Section:     "welcome-screen",
		Key:         "allow-exit",
		Default:     "true",
		Description: "Should the exit button be visible in the welcome screen",
		Type:        ValueTypeBoolean,
		EnumValues:  []string{"true", "false"},
	})

	// ============================================================================
	// [remote-touch] SECTION
	// ============================================================================

	// interface
	schema.AddOption(&ConfigOption{
		Section:     "remote-touch",
		Key:         "interface",
		Default:     "",
		Description: "Network interface for remote touch. This is the interface remote devices will connect to. DEFAULT: em1 (Linux (if server-ip is not specified, otherwise empty)), empty (other OS)",
		Type:        ValueTypeString,
	})

	// port
	schema.AddOption(&ConfigOption{
		Section:     "remote-touch",
		Key:         "port",
		Default:     "5010",
		Description: "First port to use for remote touch. Each source uses a different port starting from the specified port.",
		Type:        ValueTypeNumber,
	})

	// server-ip
	schema.AddOption(&ConfigOption{
		Section:     "remote-touch",
		Key:         "server-ip",
		Default:     "",
		Description: "IP address of the remote touch server (if no interface is specified). This is useful for Windows platform where the interface cannot be relied on to remain constant.",
		Type:        ValueTypeString,
	})

	// ============================================================================
	// [password-security] SECTION
	// ============================================================================

	// min-length
	schema.AddOption(&ConfigOption{
		Section:     "password-security",
		Key:         "min-length",
		Default:     "4",
		Description: "Canvas password must be at least n characters. Set to 0 or empty for no restriction.",
		Type:        ValueTypeNumber,
	})

	// min-lowercase
	schema.AddOption(&ConfigOption{
		Section:     "password-security",
		Key:         "min-lowercase",
		Default:     "0",
		Description: "Canvas password must contain at least n lowercase characters. Set to 0 or empty for no restriction.",
		Type:        ValueTypeNumber,
	})

	// min-uppercase
	schema.AddOption(&ConfigOption{
		Section:     "password-security",
		Key:         "min-uppercase",
		Default:     "0",
		Description: "Canvas password must contain at least n uppercase characters. Set to 0 or empty for no restriction.",
		Type:        ValueTypeNumber,
	})

	// min-numeric
	schema.AddOption(&ConfigOption{
		Section:     "password-security",
		Key:         "min-numeric",
		Default:     "0",
		Description: "Canvas password must contain at least n numeric characters. Set to 0 or empty for no restriction.",
		Type:        ValueTypeNumber,
	})

	// min-symbol
	schema.AddOption(&ConfigOption{
		Section:     "password-security",
		Key:         "min-symbol",
		Default:     "0",
		Description: "Canvas password must contain at least n symbol characters. Set to 0 or empty for no restriction.",
		Type:        ValueTypeNumber,
	})

	// max-repeats
	schema.AddOption(&ConfigOption{
		Section:     "password-security",
		Key:         "max-repeats",
		Default:     "0",
		Description: "Canvas password cannot contain any string of more than n repeating characters. Set to 0 or empty for no restriction.",
		Type:        ValueTypeNumber,
	})

	// max-sequence
	schema.AddOption(&ConfigOption{
		Section:     "password-security",
		Key:         "max-sequence",
		Default:     "0",
		Description: "Canvas password cannot contain any ascending or descending sequence of more than n alphanumeric characters. Set to 0 or empty for no restriction.",
		Type:        ValueTypeNumber,
	})

	// ============================================================================
	// [content] SECTION
	// ============================================================================

	// root
	schema.AddOption(&ConfigOption{
		Section:     "content",
		Key:         "root",
		Default:     "",
		Description: "Root folder for the application. DEFAULT: ~/MultiTaction/canvus-data (Linux) or C:\\Users\\username\\AppData\\Roaming\\MultiTaction/canvus-data (Windows). NOTE: Backslashes in paths need special handling on Windows computers. Replace each \\ with \\\\ or /.",
		Type:        ValueTypeFilePath,
	})

	// ============================================================================
	// [canvas] SECTION
	// ============================================================================

	// size
	schema.AddOption(&ConfigOption{
		Section:     "canvas",
		Key:         "size",
		Default:     "9600",
		Description: "Size of the canvas area. This can be specified with either one or two values. Two values (separated by space) can be used to define the width and height of the canvas. If only one value is specified, it represents the length of the longer edge of the canvas area. The shorter edge is calculated automatically when a new canvas is created so that the canvas will have the same aspect ratio as the the installation it is created on.",
		Type:        ValueTypeString,
	})

	// pin-canvas
	schema.AddOption(&ConfigOption{
		Section:     "canvas",
		Key:         "pin-canvas",
		Default:     "false",
		Description: "If true, the canvas is automatically pinned when it is opened.",
		Type:        ValueTypeBoolean,
		EnumValues:  []string{"true", "false"},
	})

	// anchor-viewport
	schema.AddOption(&ConfigOption{
		Section:     "canvas",
		Key:         "anchor-viewport",
		Default:     "false",
		Description: "If true, the canvas is automatically moved to the first anchor position when it is opened. If there are no anchors or it is false, the zoom-viewport setting is used to determine the position.",
		Type:        ValueTypeBoolean,
		EnumValues:  []string{"true", "false"},
	})

	// zoom-viewport
	schema.AddOption(&ConfigOption{
		Section:     "canvas",
		Key:         "zoom-viewport",
		Default:     "false",
		Description: "If true, the canvas is automatically zoomed to ensure all widgets are visible when it is opened. If false, the initial position will be determined by the last saved position within the session. NOTE: If there is no saved position then the viewport will be centralized.",
		Type:        ValueTypeBoolean,
		EnumValues:  []string{"true", "false"},
	})

	// auto-pin-after
	schema.AddOption(&ConfigOption{
		Section:     "canvas",
		Key:         "auto-pin-after",
		Default:     "0",
		Description: "If set, an unpinned canvas is pinned after it has not been scaled or moved for the specified number of seconds. If zero, the canvas is never auto-pinned.",
		Type:        ValueTypeNumber,
	})

	// custom-menu
	schema.AddOption(&ConfigOption{
		Section:     "canvas",
		Key:         "custom-menu",
		Default:     "",
		Description: "Custom menu file in YAML format used to display a user-defined menu structure on the finger menu. NOTE: Backslashes in paths need special handling on Windows computers. Replace each \\ with \\\\ or use forward slashes.",
		Type:        ValueTypeFilePath,
	})

	// ============================================================================
	// [widget] SECTION
	// ============================================================================

	// auto-pin-after
	schema.AddOption(&ConfigOption{
		Section:     "widget",
		Key:         "auto-pin-after",
		Default:     "0",
		Description: "This setting only applies to local canvases. If set, an unpinned widget is pinned after it has not been scaled or moved for the specified number of seconds. If zero, the widget is never auto-pinned.",
		Type:        ValueTypeNumber,
	})

	// activation-gestures
	schema.AddOption(&ConfigOption{
		Section:     "widget",
		Key:         "activation-gestures",
		Default:     "all",
		Description: "Comma separated list of gestures that activate the unpinned item. When item is activated, its border menus are displayed. Possible gestures are: all - Any interaction with the item will activate it, tap - Item is tapped or clicked, tap-and-delay - Item is tapped and no interaction occurs for short duration, hold - Item has finger(s) held on top of it for a short duration.",
		Type:        ValueTypeCommaList,
		EnumValues:  []string{"all", "tap", "tap-and-delay", "hold"},
	})

	// ============================================================================
	// [pdf] SECTION
	// ============================================================================

	// raster-resolution
	schema.AddOption(&ConfigOption{
		Section:     "pdf",
		Key:         "raster-resolution",
		Default:     "2048",
		Description: "Resolution in pixels for PDFs to be rasterized. The longer edge of a PDF page will match the specified resolution.",
		Type:        ValueTypeNumber,
	})

	// ============================================================================
	// [video] SECTION
	// ============================================================================

	// tap-to-play
	schema.AddOption(&ConfigOption{
		Section:     "video",
		Key:         "tap-to-play",
		Default:     "false",
		Description: "If true, tapping a video will play or pause the video. If false, the play button must be tapped explicitly.",
		Type:        ValueTypeBoolean,
		EnumValues:  []string{"true", "false"},
	})

	// screenshare-audio
	schema.AddOption(&ConfigOption{
		Section:     "video",
		Key:         "screenshare-audio",
		Default:     "false",
		Description: "If false, audio will be initially muted when a screenshare widget is opened. DEFAULT: false (videos are muted)",
		Type:        ValueTypeBoolean,
		EnumValues:  []string{"true", "false"},
	})

	// ============================================================================
	// [video-stream] SECTION
	// ============================================================================

	// min-fps
	schema.AddOption(&ConfigOption{
		Section:     "video-stream",
		Key:         "min-fps",
		Default:     "14",
		Description: "Set the minimum preferred frame rate",
		Type:        ValueTypeNumber,
	})

	// max-fps
	schema.AddOption(&ConfigOption{
		Section:     "video-stream",
		Key:         "max-fps",
		Default:     "31",
		Description: "Set the maximum preferred frame rate",
		Type:        ValueTypeNumber,
	})

	// min-resolution
	schema.AddOption(&ConfigOption{
		Section:     "video-stream",
		Key:         "min-resolution",
		Default:     "800 600",
		Description: "Set the minimum preferred pixel resolution",
		Type:        ValueTypeString,
	})

	// max-resolution
	schema.AddOption(&ConfigOption{
		Section:     "video-stream",
		Key:         "max-resolution",
		Default:     "3840 2160",
		Description: "Set the maximum preferred pixel resolution",
		Type:        ValueTypeString,
	})

	// prefer-uncompressed
	schema.AddOption(&ConfigOption{
		Section:     "video-stream",
		Key:         "prefer-uncompressed",
		Default:     "true",
		Description: "Indicate preference for using full-frame (uncompressed) video streams",
		Type:        ValueTypeBoolean,
		EnumValues:  []string{"true", "false"},
	})

	// ============================================================================
	// [image-search] SECTION
	// ============================================================================

	// api-key
	schema.AddOption(&ConfigOption{
		Section:     "image-search",
		Key:         "api-key",
		Default:     "",
		Description: "Google Custom Search API key. See Image search documentation for details. DEFAULT: <Multitaction API key - private data>",
		Type:        ValueTypeString,
	})

	// engine-id
	schema.AddOption(&ConfigOption{
		Section:     "image-search",
		Key:         "engine-id",
		Default:     "",
		Description: "Google Custom Search Engine id. DEFAULT: <Multitaction Engine Id - private data>",
		Type:        ValueTypeString,
	})

	// ============================================================================
	// [hardware] SECTION
	// ============================================================================

	// third-party-touch
	schema.AddOption(&ConfigOption{
		Section:     "hardware",
		Key:         "third-party-touch",
		Default:     "false",
		Description: "Enable or disable third party touch mode. This mode is meant to be used with devices that do not have a separate pen and/or eraser functionality and only work with touch input. When enabled, users will see extra UI that allows them to use their fingers to annotate and erase.",
		Type:        ValueTypeBoolean,
		EnumValues:  []string{"true", "false"},
	})

	// ============================================================================
	// [touch-mode-selector] SECTION
	// ============================================================================

	// timeout
	schema.AddOption(&ConfigOption{
		Section:     "touch-mode-selector",
		Key:         "timeout",
		Default:     "10",
		Description: "Idle time in seconds before the selector is returned to its home location in finger mode.",
		Type:        ValueTypeNumber,
	})

	// home-x-percent
	schema.AddOption(&ConfigOption{
		Section:     "touch-mode-selector",
		Key:         "home-x-percent",
		Default:     "90",
		Description: "Home position of the selector as a percentage of the window width.",
		Type:        ValueTypeNumber,
	})

	// home-y-percent
	schema.AddOption(&ConfigOption{
		Section:     "touch-mode-selector",
		Key:         "home-y-percent",
		Default:     "90",
		Description: "Home position of the selector as a percentage of the window height.",
		Type:        ValueTypeNumber,
	})

	// home-radius
	schema.AddOption(&ConfigOption{
		Section:     "touch-mode-selector",
		Key:         "home-radius",
		Default:     "500",
		Description: "Distance in pixels from the home position outside which the selector will return to the home position.",
		Type:        ValueTypeNumber,
	})

	// ============================================================================
	// [server:<name>] SECTION (Compound)
	// ============================================================================
	// Mark as compound section
	if section := schema.GetSection("server"); section == nil {
		section = &ConfigSection{
			Name:            "server",
			Description:     "Canvus Connect server list. Configure each server with [server:<name>] section. You can have multiple servers, but each server name must be unique.",
			Options:         []*ConfigOption{},
			IsCompound:      true,
			Pattern:         "server",
			CompoundEntries: make(map[string][]*ConfigOption),
		}
		schema.Sections["server"] = section
	}
	serverSection := schema.GetSection("server")
	serverSection.IsCompound = true
	serverSection.Pattern = "server"

	// server (hostname)
	schema.AddOption(&ConfigOption{
		Section:     "server",
		Key:         "server",
		Default:     "",
		Description: "The Canvus Connect server host to connect to. NOTE: This value must be specified.",
		Type:        ValueTypeString,
		IsCompound:  true,
		Pattern:     "server",
	})

	// protocol
	schema.AddOption(&ConfigOption{
		Section:     "server",
		Key:         "protocol",
		Default:     "ssl",
		Description: "Protocol used for communication with the Canvus Connect server. Options are ssl or tcp. The default protocol is determined by the chosen port below. DEFAULT (port 443): ssl DEFAULT (port 80): tcp",
		Type:        ValueTypeEnum,
		EnumValues:  []string{"ssl", "tcp"},
		IsCompound:  true,
		Pattern:     "server",
	})

	// port
	schema.AddOption(&ConfigOption{
		Section:     "server",
		Key:         "port",
		Default:     "443",
		Description: "Port used for communication with the Canvus Connect server. The default port is determined by the chosen protocol above. DEFAULT (ssl protocol): 443 DEFAULT (tcp protocol): 80",
		Type:        ValueTypeNumber,
		IsCompound:  true,
		Pattern:     "server",
	})

	// rest-api-access
	schema.AddOption(&ConfigOption{
		Section:     "server",
		Key:         "rest-api-access",
		Default:     "none",
		Description: "Control REST API access to this client from this server.",
		Type:        ValueTypeEnum,
		EnumValues:  []string{"none", "ro", "rw"},
		IsCompound:  true,
		Pattern:     "server",
	})

	// connection-password
	schema.AddOption(&ConfigOption{
		Section:     "server",
		Key:         "connection-password",
		Default:     "",
		Description: "Set the connection password (if required) for connecting to this server. The password cannot contain semicolon, or quotation marks.",
		Type:        ValueTypeString,
		IsCompound:  true,
		Pattern:     "server",
	})

	// ============================================================================
	// [identity] SECTION
	// ============================================================================

	// installation-name
	schema.AddOption(&ConfigOption{
		Section:     "identity",
		Key:         "installation-name",
		Default:     "",
		Description: "Name of the installation of this Canvus client. Is shown in the participant list for this client. If empty, the hostname of the machine is used.",
		Type:        ValueTypeString,
	})

	// ============================================================================
	// [remote-desktop:<name>] SECTION (Compound)
	// ============================================================================
	// Mark as compound section
	if section := schema.GetSection("remote-desktop"); section == nil {
		section = &ConfigSection{
			Name:            "remote-desktop",
			Description:     "Remote desktop connection configuration block. There is no limit on the number of remote-desktop configuration blocks as long as they all have unique names.",
			Options:         []*ConfigOption{},
			IsCompound:      true,
			Pattern:         "remote-desktop",
			CompoundEntries: make(map[string][]*ConfigOption),
		}
		schema.Sections["remote-desktop"] = section
	}
	remoteDesktopSection := schema.GetSection("remote-desktop")
	remoteDesktopSection.IsCompound = true
	remoteDesktopSection.Pattern = "remote-desktop"

	// host (note: example file uses "host" not "hostname")
	schema.AddOption(&ConfigOption{
		Section:     "remote-desktop",
		Key:         "host",
		Default:     "",
		Description: "Hostname or IP address of the RDP server with an optional port number. If this parameter is empty, the whole configuration block is disabled. Examples: localhost, rdp.example.com, 10.0.0.1:1234",
		Type:        ValueTypeString,
	})

	// username
	schema.AddOption(&ConfigOption{
		Section:     "remote-desktop",
		Key:         "username",
		Default:     "",
		Description: "Username to use to log in to the server.",
		Type:        ValueTypeString,
	})

	// password
	schema.AddOption(&ConfigOption{
		Section:     "remote-desktop",
		Key:         "password",
		Default:     "",
		Description: "Password for the given username. Leave empty if password is not used or if you wish to enter the password manually when connecting.",
		Type:        ValueTypeString,
	})

	// domain
	schema.AddOption(&ConfigOption{
		Section:     "remote-desktop",
		Key:         "domain",
		Default:     "",
		Description: "Domain to use for authentication.",
		Type:        ValueTypeString,
	})

	// max-connections
	schema.AddOption(&ConfigOption{
		Section:     "remote-desktop",
		Key:         "max-connections",
		Default:     "0",
		Description: "Maximum number of concurrent connections to this server. DEFAULT: 0 (unlimited)",
		Type:        ValueTypeNumber,
	})

	// show-in-server-list
	schema.AddOption(&ConfigOption{
		Section:     "remote-desktop",
		Key:         "show-in-server-list",
		Default:     "true",
		Description: "Show this server in the Remote Desktop list. If this is set to false, you can't directly connect to the server, instead this connection is just used to open files. Either show-in-server-list must be true or file-extensions must not be empty.",
		Type:        ValueTypeBoolean,
		EnumValues:  []string{"true", "false"},
	})

	// file-extensions
	schema.AddOption(&ConfigOption{
		Section:     "remote-desktop",
		Key:         "file-extensions",
		Default:     "",
		Description: "Comma-separated list of file extensions of files that are opened through this server. Either show-in-server-list must be true or file-extensions must not be empty. Example: doc,docx,ppt,pptx,xls,xlsx",
		Type:        ValueTypeCommaList,
	})

	// gateway
	schema.AddOption(&ConfigOption{
		Section:     "remote-desktop",
		Key:         "gateway",
		Default:     "",
		Description: "RDP gateway hostname with optional port number. Leave empty if gateway is not used.",
		Type:        ValueTypeString,
	})

	// gateway-username
	schema.AddOption(&ConfigOption{
		Section:     "remote-desktop",
		Key:         "gateway-username",
		Default:     "",
		Description: "Gateway username.",
		Type:        ValueTypeString,
	})

	// gateway-password
	schema.AddOption(&ConfigOption{
		Section:     "remote-desktop",
		Key:         "gateway-password",
		Default:     "",
		Description: "Gateway password.",
		Type:        ValueTypeString,
	})

	// gateway-domain
	schema.AddOption(&ConfigOption{
		Section:     "remote-desktop",
		Key:         "gateway-domain",
		Default:     "",
		Description: "Gateway Domain.",
		Type:        ValueTypeString,
	})

	// desktop-size
	schema.AddOption(&ConfigOption{
		Section:     "remote-desktop",
		Key:         "desktop-size",
		Default:     "1920 1080",
		Description: "Initial RDP desktop resolution for new RDP widgets.",
		Type:        ValueTypeString,
	})

	// app
	schema.AddOption(&ConfigOption{
		Section:     "remote-desktop",
		Key:         "app",
		Default:     "",
		Description: "RemoteApp to launch on connect. This setting is used only when connecting using the Remote Desktop list. NOTE: Backslashes in paths need special handling on Windows computers. Replace each \\ with \\\\. Example: %WINDIR%\\\\System32\\\\notepad.exe Example: ||notepad (using configured RemoteApp alias) DEFAULT: empty (show the whole desktop)",
		Type:        ValueTypeFilePath,
	})

	// app-args
	schema.AddOption(&ConfigOption{
		Section:     "remote-desktop",
		Key:         "app-args",
		Default:     "",
		Description: "Command line arguments to RemoteApp, used when \"app\" setting is used.",
		Type:        ValueTypeString,
	})

	// extra-args
	schema.AddOption(&ConfigOption{
		Section:     "remote-desktop",
		Key:         "extra-args",
		Default:     "",
		Description: "Extra FreeRDP command line arguments. Example: /drive:home,/home/multi",
		Type:        ValueTypeString,
	})

	// ============================================================================
	// [cache] SECTION
	// ============================================================================

	// max-age
	schema.AddOption(&ConfigOption{
		Section:     "cache",
		Key:         "max-age",
		Default:     "2592000",
		Description: "Maximum time in seconds that assets are allowed to be cached. DEFAULT: 2592000 (30 days). To disable caching set max-age=0.",
		Type:        ValueTypeNumber,
	})

	// ============================================================================
	// [browser] SECTION
	// ============================================================================

	// start-page
	schema.AddOption(&ConfigOption{
		Section:     "browser",
		Key:         "start-page",
		Default:     "https://www.google.com",
		Description: "Browser start page that is opened when a new web browser widget is created.",
		Type:        ValueTypeString,
	})

	// search-url
	schema.AddOption(&ConfigOption{
		Section:     "browser",
		Key:         "search-url",
		Default:     "https://www.google.com/search?q=QUERY",
		Description: "Search URL that is opened when a search term is entered in the web browser address bar. A word QUERY written in uppercase letters is replaced with the search terms.",
		Type:        ValueTypeString,
	})

	// upload-sites
	schema.AddOption(&ConfigOption{
		Section:     "browser",
		Key:         "upload-sites",
		Default:     "*",
		Description: "Specify which internet sites allow content to be uploaded by dragging onto a browser widget. if upload-sites=* or missing, allow all uploads. if upload-sites is empty, prevent all uploads. if upload-sites are specified a widget can only be dropped on the browser if its current URL contains one of the specified names, e.g. upload-sites=drive.google.com,docs.google.com,mail.google.com",
		Type:        ValueTypeCommaList,
	})

	// persistent-sessions
	schema.AddOption(&ConfigOption{
		Section:     "browser",
		Key:         "persistent-sessions",
		Default:     "true",
		Description: "Specify if browser sessions should be saved on disk so that cookies, browsing history and other session data persists between application restarts. Set to false to use in-memory sessions that are cleared when application is restarted or user logs out.",
		Type:        ValueTypeBoolean,
		EnumValues:  []string{"true", "false"},
	})

	// mouse-emulation
	schema.AddOption(&ConfigOption{
		Section:     "browser",
		Key:         "mouse-emulation",
		Default:     "false",
		Description: "Browsers can handle touches by either passing all touches to the browser or converting the first touch to mouse clicks and drags. Different web apps work better with the two alternative behaviors. The user can pick the best behavior on the browser. This setting determines the default state for new browsers. If true, then, by default, touches are converted to mouse clicks.",
		Type:        ValueTypeBoolean,
		EnumValues:  []string{"true", "false"},
	})

	// ============================================================================
	// [backup] SECTION
	// ============================================================================

	// root
	schema.AddOption(&ConfigOption{
		Section:     "backup",
		Key:         "root",
		Default:     "",
		Description: "Default path to store automatic client backups. Set to empty to disable automatic backup on software update. DEFAULT (Linux): ~/MultiTaction/canvus/backups DEFAULT (Windows): <LOCAL APP DATA>\\MultiTaction\\canvus\\backups",
		Type:        ValueTypeFilePath,
	})

	// ============================================================================
	// [auxiliary-pc] SECTION
	// ============================================================================

	// host
	schema.AddOption(&ConfigOption{
		Section:     "auxiliary-pc",
		Key:         "host",
		Default:     "",
		Description: "IP address of the auxiliary PC - leave blank to disable auxiliary PC functionality.",
		Type:        ValueTypeString,
	})

	// port
	schema.AddOption(&ConfigOption{
		Section:     "auxiliary-pc",
		Key:         "port",
		Default:     "",
		Description: "Port to use for communicating with the MTCanvusAgent on the auxiliary PC.",
		Type:        ValueTypeString,
	})

	// primary
	schema.AddOption(&ConfigOption{
		Section:     "auxiliary-pc",
		Key:         "primary",
		Default:     "USB",
		Description: "Mount folder name under the location specified by mount-folder (see below) for remote unencrypted USB drive.",
		Type:        ValueTypeString,
	})

	// secondary
	schema.AddOption(&ConfigOption{
		Section:     "auxiliary-pc",
		Key:         "secondary",
		Default:     "Encrypted",
		Description: "Mount folder name under the location specified by mount-folder (see below) for remote encrypted USB drive.",
		Type:        ValueTypeString,
	})

	// aux
	schema.AddOption(&ConfigOption{
		Section:     "auxiliary-pc",
		Key:         "aux",
		Default:     "Auxpc",
		Description: "Mount folder name under the location specified by mount-folder (see below) for sharing files to be opened on the auxiliary PC.",
		Type:        ValueTypeString,
	})

	// primary-share
	schema.AddOption(&ConfigOption{
		Section:     "auxiliary-pc",
		Key:         "primary-share",
		Default:     "",
		Description: "Auxiliary PC share name for unencrypted USB drive, e.g. CanvusUSB. This value is configured on the Auxiliary PC but can be overridden here.",
		Type:        ValueTypeString,
	})

	// secondary-share
	schema.AddOption(&ConfigOption{
		Section:     "auxiliary-pc",
		Key:         "secondary-share",
		Default:     "",
		Description: "Auxiliary PC share name for encrypted USB drive, e.g. CanvusEncrypted. This value is configured on the Auxiliary PC but can be overridden here.",
		Type:        ValueTypeString,
	})

	// aux-share
	schema.AddOption(&ConfigOption{
		Section:     "auxiliary-pc",
		Key:         "aux-share",
		Default:     "Desktop",
		Description: "Auxiliary PC share name for shared files to be opened on the auxiliary PC.",
		Type:        ValueTypeString,
	})

	// username
	schema.AddOption(&ConfigOption{
		Section:     "auxiliary-pc",
		Key:         "username",
		Default:     "multi",
		Description: "Windows user for auxiliary PC.",
		Type:        ValueTypeString,
	})

	// password
	schema.AddOption(&ConfigOption{
		Section:     "auxiliary-pc",
		Key:         "password",
		Default:     "",
		Description: "Windows user password for auxiliary PC.",
		Type:        ValueTypeString,
	})

	// remote-source
	schema.AddOption(&ConfigOption{
		Section:     "auxiliary-pc",
		Key:         "remote-source",
		Default:     "/dev/video0",
		Description: "Video source being used for the auxiliary PC desktop. Example strings for Linux and Windows are suggested in the example file.",
		Type:        ValueTypeString,
	})

	// desktop-mode-exts
	schema.AddOption(&ConfigOption{
		Section:     "auxiliary-pc",
		Key:         "desktop-mode-exts",
		Default:     "doc,docx,ppt,pptx,xls,xlsx",
		Description: "Comma-separated list of file extensions supported.",
		Type:        ValueTypeCommaList,
	})

	// log-messages
	schema.AddOption(&ConfigOption{
		Section:     "auxiliary-pc",
		Key:         "log-messages",
		Default:     "true",
		Description: "Set to true to log TCP messages to/from the auxiliary PC.",
		Type:        ValueTypeBoolean,
		EnumValues:  []string{"true", "false"},
	})

	// canvus-mapped-drive
	schema.AddOption(&ConfigOption{
		Section:     "auxiliary-pc",
		Key:         "canvus-mapped-drive",
		Default:     "",
		Description: "The drive letter assigned on the AuxPC to the network share on the CanvusPC (as configured in the 'local-share' section below).",
		Type:        ValueTypeString,
	})

	// remote-touch-interface
	schema.AddOption(&ConfigOption{
		Section:     "auxiliary-pc",
		Key:         "remote-touch-interface",
		Default:     "",
		Description: "Network interface for remote touch on the auxiliary PC. This is the interface remote devices will connect to. If empty, the following values are used: * Linux: em2 (if remote-touch-server-ip is not specified, otherwise empty) * Windows: empty",
		Type:        ValueTypeString,
	})

	// remote-touch-port
	schema.AddOption(&ConfigOption{
		Section:     "auxiliary-pc",
		Key:         "remote-touch-port",
		Default:     "5020",
		Description: "The port number to use for remote touch on any auxiliary PC widgets (this should differ from the port value used in the [remote touch] section above).",
		Type:        ValueTypeNumber,
	})

	// remote-touch-server-ip
	schema.AddOption(&ConfigOption{
		Section:     "auxiliary-pc",
		Key:         "remote-touch-server-ip",
		Default:     "",
		Description: "IP address of the auxiliary PC remote touch server (if no interface is specified). This is useful for Windows platform where the interface cannot be relied on to remain constant.",
		Type:        ValueTypeString,
	})

	// ============================================================================
	// [local-share] SECTION
	// ============================================================================

	// shared-folder
	schema.AddOption(&ConfigOption{
		Section:     "local-share",
		Key:         "shared-folder",
		Default:     "",
		Description: "The folder on the local Canvus PC which has been exposed as an SMB share for the AuxPC to connect to. NOTE: Backslashes in paths need special handling on Windows computers. Replace each \\ with \\\\ or /.",
		Type:        ValueTypeFilePath,
	})

	// share-name
	schema.AddOption(&ConfigOption{
		Section:     "local-share",
		Key:         "share-name",
		Default:     "",
		Description: "The name of the share given to the folder above (i.e. as it appears on the network).",
		Type:        ValueTypeString,
	})

	// ============================================================================
	// [output:<name>] SECTION (Compound)
	// ============================================================================
	// Mark as compound section
	if section := schema.GetSection("output"); section == nil {
		section = &ConfigSection{
			Name:            "output",
			Description:     "Video output settings. Configure an output block with [output:name] for each output. You can have multiple outputs, each output name must be unique. The output name will be visible in Canvus.",
			Options:         []*ConfigOption{},
			IsCompound:      true,
			Pattern:         "output",
			CompoundEntries: make(map[string][]*ConfigOption),
		}
		schema.Sections["output"] = section
	}
	outputSection := schema.GetSection("output")
	outputSection.IsCompound = true
	outputSection.Pattern = "output"

	// location
	schema.AddOption(&ConfigOption{
		Section:     "output",
		Key:         "location",
		Default:     "",
		Description: "Location of the output (x y coordinates)",
		Type:        ValueTypeString,
		IsCompound:  true,
		Pattern:     "output",
	})

	// size
	schema.AddOption(&ConfigOption{
		Section:     "output",
		Key:         "size",
		Default:     "",
		Description: "Size of the output (width height)",
		Type:        ValueTypeString,
		IsCompound:  true,
		Pattern:     "output",
	})

	// ============================================================================
	// [remote-mount] SECTION
	// ============================================================================

	// daemon-port
	schema.AddOption(&ConfigOption{
		Section:     "remote-mount",
		Key:         "daemon-port",
		Default:     "8081",
		Description: "Port to use for communicating with the local mt-canvus-daemon.",
		Type:        ValueTypeNumber,
	})

	// mount-options
	schema.AddOption(&ConfigOption{
		Section:     "remote-mount",
		Key:         "mount-options",
		Default:     "",
		Description: "Comma separated list of additional options to pass to the mount command, e.g. vers=2.0 to set the SMB version. Note that this setting is ignored on Windows installations.",
		Type:        ValueTypeCommaList,
	})

	// mount-folder
	schema.AddOption(&ConfigOption{
		Section:     "remote-mount",
		Key:         "mount-folder",
		Default:     "/mnt",
		Description: "Linux: The root folder on this PC for mounting shares into. Windows: The local folder to hold links to remote shares. NOTE: Backslashes in paths need special handling on Windows computers. Replace each \\ with \\\\ or /.",
		Type:        ValueTypeFilePath,
	})

	// log-messages
	schema.AddOption(&ConfigOption{
		Section:     "remote-mount",
		Key:         "log-messages",
		Default:     "true",
		Description: "Set to true to log TCP messages to/from the mt-canvus-daemon.",
		Type:        ValueTypeBoolean,
		EnumValues:  []string{"true", "false"},
	})

	return schema
}

