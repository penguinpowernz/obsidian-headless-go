package helpers

import (
	"fmt"

	"github.com/penguinpowernz/obsidian-headless-go/internal/config"
)

// RequireAuth ensures the user is authenticated and returns the token
func RequireAuth() (string, error) {
	token, err := config.LoadAuthToken()
	if err != nil {
		return "", fmt.Errorf("load auth token: %w", err)
	}

	if token == nil || token.Token == "" {
		return "", fmt.Errorf("not logged in. Run 'ob login' first")
	}

	return token.Token, nil
}
