package config

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// InteractivePrompts handles user input for configuration.
type InteractivePrompts struct {
	reader *bufio.Reader
}

// NewInteractivePrompts creates a new InteractivePrompts and registers a SIGTERM
// handler that restores the terminal on interrupt.
func NewInteractivePrompts() *InteractivePrompts {
	p := &InteractivePrompts{
		reader: bufio.NewReader(os.Stdin),
	}
	p.setupSignalHandler()
	return p
}

func (p *InteractivePrompts) setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if term.IsTerminal(int(syscall.Stdin)) {
			fmt.Print("\n")
		}
		os.Exit(1)
	}()
}

// PromptForConfig prompts the user for all required configuration values.
func (p *InteractivePrompts) PromptForConfig() (*Config, error) {
	fmt.Println("Canvus Server DB Solver Configuration")
	fmt.Println("======================================")
	fmt.Println()

	cfg := DefaultConfig()

	fmt.Println("Canvus Server Configuration")
	fmt.Println("---------------------------")
	fmt.Println("Server URL: https://localhost:443 (local access only)")

	cfg.CanvusServer.Username = p.promptString("Username", "", true)
	cfg.CanvusServer.Password = p.promptPassword("Password")
	cfg.CanvusServer.Timeout = p.promptInt("API Timeout (seconds)", 30)
	fmt.Println()

	fmt.Println("Paths Configuration")
	fmt.Println("-------------------")
	cfg.Paths.AssetsFolder = p.promptPath("Assets Folder Path", "", true)
	cfg.Paths.BackupRootFolder = p.promptPath("Backup Root Folder Path", "", true)
	cfg.Paths.OutputFolder = p.promptPath("Output Folder Path", "./output", false)
	fmt.Println()

	fmt.Println("Logging Configuration")
	fmt.Println("---------------------")
	verbose := p.promptBool("Enable verbose logging", false)
	cfg.Logging.Verbose = verbose
	if verbose {
		cfg.Logging.Level = "debug"
	}
	logToFile := p.promptBool("Log to file", false)
	cfg.Logging.LogToFile = logToFile
	if logToFile {
		cfg.Logging.LogFile = p.promptString("Log file path", "canvus-server-db-solver.log", false)
	}
	fmt.Println()

	fmt.Println("Performance Configuration")
	fmt.Println("-------------------------")
	cfg.Performance.MaxConcurrentAPI = p.promptInt("Max concurrent API calls", 10)
	cfg.Performance.MaxConcurrentFiles = p.promptInt("Max concurrent file operations", 20)
	fmt.Println()

	fmt.Println("Validating configuration...")
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	fmt.Println("Configuration validated successfully!")
	fmt.Println()

	return cfg, nil
}

// PromptForSaveConfig asks whether the user wants to save the configuration.
func (p *InteractivePrompts) PromptForSaveConfig(cfg *Config) error {
	if !p.promptBool("Save configuration to file", true) {
		return nil
	}
	configFile := p.promptString("Configuration file path", "config.yaml", false)
	if err := cfg.SaveConfig(configFile); err != nil {
		return fmt.Errorf("save configuration: %w", err)
	}
	fmt.Printf("Configuration saved to: %s\n", configFile)
	return nil
}

// PromptForConfirmation asks the user for a yes/no confirmation.
func (p *InteractivePrompts) PromptForConfirmation(message string) bool {
	return p.promptBool(message, false)
}

// promptString reads a string from stdin, returning defaultValue when the input is empty.
func (p *InteractivePrompts) promptString(prompt, defaultValue string, required bool) string {
	for {
		if defaultValue != "" {
			fmt.Printf("%s [%s]: ", prompt, defaultValue)
		} else {
			fmt.Printf("%s: ", prompt)
		}
		input, err := p.reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}
		input = strings.TrimSpace(input)
		if input == "" {
			if defaultValue != "" {
				return defaultValue
			}
			if required {
				fmt.Println("This field is required. Please enter a value.")
				continue
			}
			return ""
		}
		return input
	}
}

// promptPassword reads a password without echoing.
func (p *InteractivePrompts) promptPassword(prompt string) string {
	for {
		fmt.Printf("%s: ", prompt)
		if term.IsTerminal(int(syscall.Stdin)) {
			password, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				fmt.Printf("\nError reading password: %v\n", err)
				continue
			}
			fmt.Println()
			if len(password) == 0 {
				fmt.Println("Password is required.")
				continue
			}
			return string(password)
		}
		// Non-interactive — read plain text.
		password, err := p.reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading password: %v\n", err)
			continue
		}
		password = strings.TrimSpace(password)
		if password == "" {
			fmt.Println("Password is required.")
			continue
		}
		return password
	}
}

// promptInt reads an integer from stdin.
func (p *InteractivePrompts) promptInt(prompt string, defaultValue int) int {
	for {
		input := p.promptString(prompt, fmt.Sprintf("%d", defaultValue), false)
		if input == "" {
			return defaultValue
		}
		var value int
		if _, err := fmt.Sscanf(input, "%d", &value); err != nil {
			fmt.Println("Please enter a valid integer.")
			continue
		}
		if value < 1 {
			fmt.Println("Value must be at least 1.")
			continue
		}
		return value
	}
}

// promptBool reads a yes/no answer from stdin.
func (p *InteractivePrompts) promptBool(prompt string, defaultValue bool) bool {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}
	for {
		input := p.promptString(prompt+" (y/n)", defaultStr, false)
		if input == "" {
			return defaultValue
		}
		switch strings.ToLower(input) {
		case "y", "yes", "true", "1":
			return true
		case "n", "no", "false", "0":
			return false
		default:
			fmt.Println("Please enter 'y' for yes or 'n' for no.")
		}
	}
}

// promptPath reads a file system path, optionally requiring it to exist.
func (p *InteractivePrompts) promptPath(prompt, defaultValue string, mustExist bool) string {
	for {
		path := p.promptString(prompt, defaultValue, true)
		if strings.HasPrefix(path, "~") {
			home, err := os.UserHomeDir()
			if err != nil {
				fmt.Printf("Error getting home directory: %v\n", err)
				continue
			}
			path = filepath.Join(home, path[1:])
		}
		absPath, err := filepath.Abs(path)
		if err != nil {
			fmt.Printf("Error converting to absolute path: %v\n", err)
			continue
		}
		if mustExist {
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				fmt.Printf("Path does not exist: %s\n", absPath)
				continue
			}
		}
		return absPath
	}
}
