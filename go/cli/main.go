// Package main provides the entry point for the Canvus CLI.
//
// The CLI is a thin wrapper over github.com/jaypaulb/MT-Canvus-Tools/go/sdk:
// command parsing/flag wiring/output formatting live here; everything
// API-shaped delegates to the SDK.
package main

import (
	"fmt"
	"os"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/audit"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/canvas"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/canvas/background"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/client"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/colorpreset"
	cmdconfig "github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/folder"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/group"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/mipmap"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/system"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/token"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/upload"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/user"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/videoinput"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/videooutput"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/widget"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/widget/anchor"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/widget/browser"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/widget/connector"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/widget/image"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/widget/note"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/widget/pdf"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/widget/video"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/commands/workspace"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/session"
	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version information (set by build flags at link time, see README for the
// recommended -ldflags incantation).
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "canvus",
	Short: "Canvus CLI - command-line interface for the Canvus collaborative workspace",
	Long: `Canvus CLI provides a command-line interface for managing canvases,
widgets, users, and groups in your Canvus collaborative workspace.

The CLI is a thin wrapper around the MT-Canvus-Tools Go SDK and supports:
- Creating and managing canvases and folders
- Adding and editing widgets (notes, images, PDFs, videos, anchors, browsers)
- Managing users, groups, and access tokens
- Administering system configuration, video inputs/outputs, clients

Configuration can be provided via command-line flags, environment variables,
or a configuration file at ~/.canvus/config.yaml.

Environment variables:
  CANVUS_URL        Server URL (e.g. https://canvus.example.com)
  CANVUS_API_KEY    Private-Token value (preferred auth method)
  CANVUS_USERNAME   Username for login-based auth (alternative to API key)
  CANVUS_PASSWORD   Password for login-based auth
  CANVUS_INSECURE   "true" to disable TLS verification (opt-in)
  CANVUS_OUTPUT     Default output format (json|yaml|table|text)
  CANVUS_VERBOSE    "true" to enable Debug-level CLI logging
  LOG_FORMAT        "json" to emit slog JSON to stderr (default: text)`,
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip configuration loading for commands that don't need it.
		if cmd.Name() == "version" || cmd.Name() == "help" || cmd.Name() == "login" ||
			(cmd.Parent() != nil && cmd.Parent().Name() == "config") {
			return nil
		}

		verbose := viper.GetBool("verbose")
		util.InitLogger(verbose)
		util.Debug("starting Canvus CLI v%s (commit: %s)", version, commit)

		cfg, err := config.Load()
		if err != nil {
			util.Error("failed to load configuration: %v", err)
			return fmt.Errorf("failed to load configuration: %w", err)
		}
		util.Debug("configuration loaded: url=%s output=%s", cfg.URL, cfg.Output)

		ctx := config.WithConfig(cmd.Context(), cfg)

		sess, err := session.NewSession(cfg)
		if err != nil {
			util.Error("failed to create session: %v", err)
			return fmt.Errorf("failed to create session: %w", err)
		}
		ctx = session.WithSession(ctx, sess)
		cmd.SetContext(ctx)

		util.Debug("SDK session created successfully")
		if cfg.Verbose {
			util.Info("verbose mode enabled")
			util.Debug("config precedence: flags > env > file > defaults")
		}
		return nil
	},
}

func init() {
	commands.Version = version
	commands.Commit = commit
	commands.Date = date

	// Global flags
	rootCmd.PersistentFlags().String("url", "", "Canvus server URL (env: CANVUS_URL)")
	rootCmd.PersistentFlags().String("api-key", "", "API key for authentication (env: CANVUS_API_KEY)")
	rootCmd.PersistentFlags().String("username", "", "Username for login authentication (env: CANVUS_USERNAME)")
	rootCmd.PersistentFlags().String("password", "", "Password for login authentication (env: CANVUS_PASSWORD)")
	rootCmd.PersistentFlags().Bool("insecure", false, "Skip TLS certificate verification (env: CANVUS_INSECURE)")
	rootCmd.PersistentFlags().StringP("output", "o", "table", "Output format: json, yaml, table, text (env: CANVUS_OUTPUT)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging (env: CANVUS_VERBOSE)")

	bindFlags := []struct{ vKey, fKey string }{
		{"url", "url"},
		{"api_key", "api-key"},
		{"username", "username"},
		{"password", "password"},
		{"insecure", "insecure"},
		{"output", "output"},
		{"verbose", "verbose"},
	}
	for _, b := range bindFlags {
		if err := viper.BindPFlag(b.vKey, rootCmd.PersistentFlags().Lookup(b.fKey)); err != nil {
			panic(fmt.Sprintf("viper.BindPFlag(%q): %v", b.vKey, err))
		}
	}

	viper.SetEnvPrefix("CANVUS")
	viper.AutomaticEnv()

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Top-level command groups
	rootCmd.AddCommand(canvas.CanvasCmd)
	rootCmd.AddCommand(widget.WidgetCmd)
	rootCmd.AddCommand(user.UserCmd)
	rootCmd.AddCommand(group.GroupCmd)
	rootCmd.AddCommand(system.SystemCmd)
	rootCmd.AddCommand(client.ClientCmd)
	rootCmd.AddCommand(folder.FolderCmd)
	rootCmd.AddCommand(workspace.WorkspaceCmd)
	rootCmd.AddCommand(token.TokenCmd)
	rootCmd.AddCommand(mipmap.MipmapCmd)
	rootCmd.AddCommand(upload.UploadCmd)
	rootCmd.AddCommand(colorpreset.ColorPresetCmd)
	rootCmd.AddCommand(videoinput.VideoInputCmd)
	rootCmd.AddCommand(videooutput.VideoOutputCmd)
	rootCmd.AddCommand(audit.AuditCmd)

	// Sub-command attachments
	canvas.CanvasCmd.AddCommand(background.BackgroundCmd)
	widget.WidgetCmd.AddCommand(note.NoteCmd)
	widget.WidgetCmd.AddCommand(image.ImageCmd)
	widget.WidgetCmd.AddCommand(pdf.PDFCmd)
	widget.WidgetCmd.AddCommand(video.VideoCmd)
	widget.WidgetCmd.AddCommand(anchor.AnchorCmd)
	widget.WidgetCmd.AddCommand(browser.BrowserCmd)
	widget.WidgetCmd.AddCommand(connector.ConnectorCmd)

	// Auth/config/info commands
	rootCmd.AddCommand(commands.GetLoginCmd())
	rootCmd.AddCommand(cmdconfig.ConfigCmd)
	rootCmd.AddCommand(commands.GetVersionCmd())
	rootCmd.AddCommand(commands.TUICmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		verbose := viper.GetBool("verbose")
		exitCode := util.HandleError(err, verbose)
		os.Exit(exitCode)
	}
}
