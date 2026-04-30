package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/penguinpowernz/obsidian-headless-go/internal/api"
	"github.com/penguinpowernz/obsidian-headless-go/internal/config"
	"github.com/penguinpowernz/obsidian-headless-go/internal/helpers"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	email    string
	password string
	mfaCode  string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to your Obsidian account",
	Long: `Login to your Obsidian account, or display login status if already logged in.

All options are interactive when omitted — email and password are prompted,
and 2FA is requested automatically if enabled on the account.`,
	RunE: runLogin,
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout and clear stored credentials",
	RunE:  runLogout,
}

func init() {
	loginCmd.Flags().StringVar(&email, "email", "", "Account email")
	loginCmd.Flags().StringVar(&password, "password", "", "Account password")
	loginCmd.Flags().StringVar(&mfaCode, "mfa", "", "MFA code")
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Check if already logged in
	existingToken, err := config.LoadAuthToken()
	if err != nil {
		return fmt.Errorf("load auth token: %w", err)
	}

	// If already logged in and no credentials provided, show status
	if existingToken != nil && email == "" && password == "" {
		fmt.Println("Already logged in")
		if existingToken.Email != "" {
			fmt.Printf("Email: %s\n", existingToken.Email)
		}
		return nil
	}

	// Interactive prompts if not provided
	if email == "" {
		fmt.Print("Email: ")
		reader := bufio.NewReader(os.Stdin)
		emailInput, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("read email: %w", err)
		}
		email = strings.TrimSpace(emailInput)
	}

	if password == "" {
		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("read password: %w", err)
		}
		fmt.Println()
		password = string(passwordBytes)
	}

	// Make login request
	client := api.NewClient("")

	loginResp, err := client.SignIn(email, password, mfaCode)
	if err != nil {
		// Check if MFA is required
		if strings.Contains(err.Error(), "mfa") && mfaCode == "" {
			fmt.Print("MFA code: ")
			reader := bufio.NewReader(os.Stdin)
			mfaInput, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("read MFA code: %w", err)
			}
			mfaCode = strings.TrimSpace(mfaInput)

			loginResp, err = client.SignIn(email, password, mfaCode)
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}
		} else {
			return fmt.Errorf("login failed: %w", err)
		}
	}

	// Save token
	token := &config.AuthToken{
		Token: loginResp.Token,
		Email: email,
	}

	if err := config.SaveAuthToken(token); err != nil {
		return fmt.Errorf("save auth token: %w", err)
	}

	fmt.Println("Login successful!")
	return nil
}

func runLogout(cmd *cobra.Command, args []string) error {
	if err := config.DeleteAuthToken(); err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}

	fmt.Println("Logged out successfully")
	return nil
}

// requireAuth is deprecated - use helpers.RequireAuth instead
// Kept for compatibility
func requireAuth() (string, error) {
	return helpers.RequireAuth()
}
