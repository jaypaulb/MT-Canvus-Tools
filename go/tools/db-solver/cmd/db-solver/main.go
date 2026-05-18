// Command db-solver is a Cobra CLI tool that recovers from corrupted Canvus-Server
// installations by querying the Canvus API for all assets, comparing against the
// on-disk asset store, and restoring missing files from configured backup folders.
//
// Quick start:
//
//	db-solver discover  # scan and report missing assets
//	db-solver lookup-hash --dry-run  # resolve no-hash assets via Postgres
//	db-solver run  # full discover → search → report workflow
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/commands"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/tools/db-solver/internal/logging"
)

var (
	version   = "dev"
	buildTime = "unknown"
	goVersion = "unknown"
)

var (
	configFileFlag   string
	dryRunFlag       bool
	iniPathFlag      string
	skipArchivedFlag bool
	lowMemoryFlag    bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "db-solver",
	Short: "Canvus Server DB Solver — Asset Recovery Tool",
	Long: `Canvus Server DB Solver identifies and restores missing Canvus assets.

It connects to a Canvus Server, identifies missing asset files by comparing API
data with the on-disk asset store, and restores files from backup locations.
Designed for large deployments with thousands of canvases and tens of thousands
of assets.

Key features:
  - Automated asset discovery via Canvus API
  - Missing asset detection through filesystem comparison
  - Backup search and recovery from multiple locations
  - Comprehensive text and CSV reports
  - Parallel processing for optimal performance`,
	Version: fmt.Sprintf("%s (built %s with %s)", version, buildTime, goVersion),
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFileFlag, "config", "", "Path to config file (default: ./config.yaml)")

	rootCmd.AddCommand(discoverCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(lookupHashCmd)

	lookupHashCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Dry-run mode: report what would be restored without copying files")
	lookupHashCmd.Flags().StringVar(&iniPathFlag, "ini-path", "", "Path to mt-canvus-server.ini (default: auto-detect)")
	lookupHashCmd.Flags().BoolVar(&skipArchivedFlag, "skip-archived", true, "Skip archived canvases during discovery")
	lookupHashCmd.Flags().BoolVar(&lowMemoryFlag, "low-memory", false, "Low-memory mode: search files on-demand instead of pre-building catalogs")
}

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover missing assets from the Canvus Server",
	Long:  "Scan the Canvus Server and identify missing asset files by comparing API data with the local filesystem.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadOrPromptConfig()
		if err != nil {
			return fmt.Errorf("configuration: %w", err)
		}
		return commands.NewDiscoverCommand(cfg).Execute()
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the complete workflow (discover, search, report)",
	Long:  "Execute the full DB Solver workflow: discover missing assets, search backups, and generate reports.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadOrPromptConfig()
		if err != nil {
			return fmt.Errorf("configuration: %w", err)
		}
		return commands.NewRunCommand(cfg).Execute(cmd, args)
	},
}

var lookupHashCmd = &cobra.Command{
	Use:   "lookup-hash",
	Short: "Resolve hash values for assets without hash via the Postgres database",
	Long:  "Process assets that lack hash values by querying the PostgreSQL asset_files table, then searching the assets and backup folders.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadOrPromptConfig()
		if err != nil {
			return fmt.Errorf("configuration: %w", err)
		}
		lc := commands.NewLookupHashCommand(cfg)
		lc.SetDryRun(dryRunFlag)
		lc.SetSkipArchived(skipArchivedFlag)
		lc.SetLowMemory(lowMemoryFlag)
		if iniPathFlag != "" {
			lc.SetINIPath(iniPathFlag)
		}
		return lc.Execute(cmd, args)
	},
}

// loadOrPromptConfig loads the config file or interactively prompts the user.
func loadOrPromptConfig() (*config.Config, error) {
	cfg, err := config.LoadConfig(configFileFlag)
	if err == nil && cfg.CanvusServer.Username != "" {
		fmt.Println("Loaded configuration from file")
		if initErr := initLogging(cfg); initErr != nil {
			fmt.Printf("Warning: failed to initialise logging: %v\n", initErr)
		}
		return cfg, nil
	}

	fmt.Println("No configuration file found — prompting for settings...")
	fmt.Println()

	prompts := config.NewInteractivePrompts()
	cfg, err = prompts.PromptForConfig()
	if err != nil {
		return nil, err
	}
	if initErr := initLogging(cfg); initErr != nil {
		fmt.Printf("Warning: failed to initialise logging: %v\n", initErr)
	}
	if saveErr := prompts.PromptForSaveConfig(cfg); saveErr != nil {
		fmt.Printf("Warning: failed to save configuration: %v\n", saveErr)
	}
	return cfg, nil
}

func initLogging(cfg *config.Config) error {
	level := logging.ParseLogLevel(cfg.Logging.Level)
	logFile := ""
	if cfg.Logging.LogToFile {
		logFile = cfg.Logging.LogFile
	}
	return logging.InitLogger(level, cfg.Logging.Verbose, logFile)
}
