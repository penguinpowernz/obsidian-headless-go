package cmd

import (
	"fmt"
	"syscall"

	"golang.org/x/term"
)

// readPassword reads a password from stdin without echoing
func readPassword() ([]byte, error) {
	fmt.Print("Password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, fmt.Errorf("read password: %w", err)
	}
	fmt.Println()
	return password, nil
}
