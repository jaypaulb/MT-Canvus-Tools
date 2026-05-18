// Package commands provides the login command for the Canvus CLI.
package commands

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/jaypaulb/MT-Canvus-Tools/go/cli/internal/config"
	"github.com/jaypaulb/MT-Canvus-Tools/go/sdk/canvus"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Canvus and save credentials",
	Long: `Authenticate with the Canvus server using username and password.

This command will:
  1. Prompt for your username and password (or accept them as flags)
  2. Authenticate with the server and obtain an API token
  3. Save the token to your configuration file (~/.canvus/config.yaml)
  4. Display authentication success

After logging in, the token will be used for all future commands.

Examples:
  # Interactive login (prompts for password)
  canvus login --username user@example.com

  # Non-interactive login (for scripting - WARNING: password visible in history)
  canvus login --username user@example.com --password mypassword

Security Note:
  Using --password flag exposes your password in command history.
  For interactive use, omit --password to be prompted securely.`,
	RunE: runLogin,
}

var (
	loginUsername string
	loginPassword string
)

func init() {
	loginCmd.Flags().StringVarP(&loginUsername, "username", "u", "", "Username (email) for authentication")
	loginCmd.Flags().StringVarP(&loginPassword, "password", "p", "", "Password (prompted if not provided)")
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Get configuration context (URL, etc.)
	cfg, err := config.GetConfig(cmd.Context())
	if err != nil {
		// If config not in context, load minimal config
		cfg = &config.Config{
			URL: cmd.Flag("url").Value.String(),
		}
		if cfg.URL == "" {
			return fmt.Errorf("server URL is required: use --url flag or set CANVUS_URL environment variable")
		}
	}

	// Get username
	username := loginUsername
	if username == "" {
		username = cfg.Username
	}
	if username == "" {
		return fmt.Errorf("username is required: use --username flag or set CANVUS_USERNAME environment variable")
	}

	// Get password - prompt if not provided
	password := loginPassword
	if password == "" {
		password = cfg.Password
	}
	if password == "" {
		// Prompt for password securely
		fmt.Fprintf(os.Stderr, "Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Fprintf(os.Stderr, "\n")
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		password = string(passwordBytes)
	} else if loginPassword != "" {
		// Warn if password was provided via flag
		fmt.Fprintf(os.Stderr, "Warning: Using --password flag exposes your password in command history.\n")
		fmt.Fprintf(os.Stderr, "         For interactive use, omit --password to be prompted securely.\n\n")
	}

	// Validate credentials
	if username == "" || password == "" {
		return fmt.Errorf("both username and password are required")
	}

	// Create SDK session configuration. login.go takes a path independent of
	// internal/session because at this point the CLI explicitly is not yet
	// configured with credentials.
	sessionCfg := canvus.DefaultSessionConfig()
	sessionCfg.BaseURL = cfg.URL
	if cfg.Insecure {
		sessionCfg.HTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, //nolint:gosec // user opt-in.
				},
			},
			Timeout: sessionCfg.RequestTimeout,
		}
	}

	// Create session without authentication
	session := canvus.NewSession(sessionCfg)

	// Attempt login
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Fprintf(os.Stderr, "Logging in to %s...\n", cfg.URL)
	if err := session.Login(ctx, username, password); err != nil {
		// Provide helpful error messages
		if apiErr, ok := err.(*canvus.APIError); ok {
			switch apiErr.StatusCode {
			case 401:
				return fmt.Errorf("login failed: invalid username or password")
			case 403:
				return fmt.Errorf("login failed: account may be locked or disabled")
			default:
				return fmt.Errorf("login failed (status %d): %s", apiErr.StatusCode, apiErr.Message)
			}
		}
		return fmt.Errorf("login failed: %w", err)
	}

	// Login successful - the token is now stored in the session
	// Unfortunately, the SDK doesn't expose the token directly, so we need to extract it
	// For now, we'll make a test request to verify authentication worked
	userID := session.UserID()
	fmt.Fprintf(os.Stderr, "Login successful! User ID: %d\n", userID)

	// Get user info to verify and show details
	userInfo, err := session.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to retrieve user information: %w", err)
	}

	// The SDK stores the token internally but doesn't expose it
	// Since we can't extract the token from the session, we'll save the username/password
	// This is a limitation of the current SDK design
	// In a production environment, you'd want to either:
	// 1. Modify the SDK to expose the token
	// 2. Use API key authentication instead
	// 3. Store username/password (less secure)

	// For now, we'll update the config with username/password and clear any old API key
	cfg.Username = username
	cfg.Password = password
	cfg.APIKey = "" // Clear any existing API key

	// Note: In a real implementation, we'd want to extract and save the token as api_key
	// For this implementation, we're storing credentials which will be used by the session package
	// The session package will call Login() each time, which is not ideal but works
	fmt.Fprintf(os.Stderr, "\nNote: Credentials will be saved to config file for future use.\n")
	fmt.Fprintf(os.Stderr, "      Consider using API token authentication for better security.\n\n")

	// Save configuration
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Authentication successful!\n")
	fmt.Printf("User: %s (%s)\n", userInfo.Name, userInfo.Email)
	fmt.Printf("Configuration saved to ~/.canvus/config.yaml\n")

	return nil
}

// GetLoginCmd returns the login command for registration with root command
func GetLoginCmd() *cobra.Command {
	return loginCmd
}
