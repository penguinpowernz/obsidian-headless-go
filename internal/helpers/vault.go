package helpers

import (
	"fmt"
	"path/filepath"

	"github.com/penguinpowernz/obsidian-headless-go/internal/api"
	"github.com/penguinpowernz/obsidian-headless-go/internal/config"
)

// ResolveVaultPath resolves a vault path to an absolute path
func ResolveVaultPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	return absPath, nil
}

// LoadVaultByPath loads a vault config by path
func LoadVaultByPath(path string) (*config.VaultConfig, error) {
	absPath, err := ResolveVaultPath(path)
	if err != nil {
		return nil, err
	}

	vaultConfig, err := config.LoadVaultConfigByPath(absPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return vaultConfig, nil
}

// FindVaultByIDOrName searches for a vault by ID or name
func FindVaultByIDOrName(vaults []api.Vault, identifier string) (*api.Vault, error) {
	var matches []api.Vault

	for i, v := range vaults {
		if v.ID == identifier {
			return &vaults[i], nil
		}
		if v.Name == identifier {
			matches = append(matches, vaults[i])
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("vault %q not found", identifier)
	}

	if len(matches) > 1 {
		return nil, fmt.Errorf("multiple vaults named %q found, use vault ID instead", identifier)
	}

	return &matches[0], nil
}

// UpdateVaultConfigFromFlags updates config based on command flags
type VaultConfigUpdater struct {
	Config  *config.VaultConfig
	Updated bool
}

// SetSyncMode sets the sync mode if provided
func (u *VaultConfigUpdater) SetSyncMode(mode string) error {
	if mode == "" {
		return nil
	}

	switch mode {
	case "bidirectional":
		u.Config.SyncMode = config.SyncModeBidirectional
	case "pull-only":
		u.Config.SyncMode = config.SyncModePullOnly
	case "mirror-remote":
		u.Config.SyncMode = config.SyncModeMirrorRemote
	default:
		return fmt.Errorf("invalid sync mode: %s", mode)
	}

	u.Updated = true
	return nil
}

// SetConflictStrategy sets the conflict strategy if provided
func (u *VaultConfigUpdater) SetConflictStrategy(strategy string) error {
	if strategy == "" {
		return nil
	}

	switch strategy {
	case "merge":
		u.Config.ConflictStrategy = config.ConflictStrategyMerge
	case "conflict":
		u.Config.ConflictStrategy = config.ConflictStrategyConflict
	default:
		return fmt.Errorf("invalid conflict strategy: %s", strategy)
	}

	u.Updated = true
	return nil
}

// SetStringSlice sets a string slice field if provided (empty string clears)
func (u *VaultConfigUpdater) SetStringSlice(value string, target *[]string) {
	if value == "" {
		return // Not provided
	}

	if value == "" {
		*target = nil
	} else {
		*target = splitCSV(value)
	}

	u.Updated = true
}

// SetString sets a string field if provided
func (u *VaultConfigUpdater) SetString(value string, target *string) {
	if value == "" {
		return
	}

	*target = value
	u.Updated = true
}

// splitCSV splits comma-separated values
func splitCSV(value string) []string {
	if value == "" {
		return nil
	}

	var result []string
	for _, part := range stringsSplit(value, ",") {
		trimmed := stringsTrim(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// Helper functions to avoid importing strings package issues
func stringsSplit(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

func stringsTrim(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n') {
		end--
	}

	return s[start:end]
}
