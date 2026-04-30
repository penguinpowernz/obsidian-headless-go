package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SyncMode defines how sync operates
type SyncMode string

const (
	SyncModeBidirectional SyncMode = "bidirectional"
	SyncModePullOnly      SyncMode = "pull-only"
	SyncModeMirrorRemote  SyncMode = "mirror-remote"
)

// ConflictStrategy defines how conflicts are resolved
type ConflictStrategy string

const (
	ConflictStrategyMerge    ConflictStrategy = "merge"
	ConflictStrategyConflict ConflictStrategy = "conflict"
)

// VaultConfig stores configuration for a synced vault
type VaultConfig struct {
	VaultID           string           `json:"vaultId"`
	VaultName         string           `json:"vaultName"`
	VaultPath         string           `json:"vaultPath"`
	Host              string           `json:"host"`
	EncryptionVersion int              `json:"encryptionVersion"`
	EncryptionKey     string           `json:"encryptionKey"`
	EncryptionSalt    string           `json:"encryptionSalt"`
	ConflictStrategy  ConflictStrategy `json:"conflictStrategy"`
	DeviceName        string           `json:"deviceName,omitempty"`
	ConfigDir         string           `json:"configDir,omitempty"`
	SyncMode          SyncMode         `json:"syncMode,omitempty"`
	IgnoreFolders     []string         `json:"ignoreFolders,omitempty"`
	AllowTypes        []string         `json:"allowTypes,omitempty"`
	AllowSpecialFiles []string         `json:"allowSpecialFiles,omitempty"`
}

// PublishConfig stores configuration for a publish site
type PublishConfig struct {
	SiteID    string   `json:"siteId"`
	Host      string   `json:"host"`
	VaultPath string   `json:"vaultPath"`
	Includes  []string `json:"includes,omitempty"`
	Excludes  []string `json:"excludes,omitempty"`
}

// AuthToken stores authentication credentials
type AuthToken struct {
	Token string `json:"token"`
	Email string `json:"email,omitempty"`
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(home, ".obsidian-headless")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", err
	}
	return configDir, nil
}

// SaveAuthToken saves the authentication token
func SaveAuthToken(token *AuthToken) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	path := filepath.Join(configDir, "auth.json")
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// LoadAuthToken loads the authentication token
func LoadAuthToken() (*AuthToken, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(configDir, "auth.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var token AuthToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// DeleteAuthToken removes the authentication token
func DeleteAuthToken() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	path := filepath.Join(configDir, "auth.json")
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// SaveVaultConfig saves a vault configuration
func SaveVaultConfig(vaultID string, config *VaultConfig) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	vaultsDir := filepath.Join(configDir, "vaults")
	if err := os.MkdirAll(vaultsDir, 0700); err != nil {
		return err
	}

	path := filepath.Join(vaultsDir, fmt.Sprintf("%s.json", vaultID))
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// LoadVaultConfig loads a vault configuration
func LoadVaultConfig(vaultID string) (*VaultConfig, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(configDir, "vaults", fmt.Sprintf("%s.json", vaultID))
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config VaultConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadVaultConfigByPath finds a vault config by its path
func LoadVaultConfigByPath(vaultPath string) (*VaultConfig, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	vaultsDir := filepath.Join(configDir, "vaults")
	entries, err := os.ReadDir(vaultsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no vaults configured")
		}
		return nil, err
	}

	absPath, err := filepath.Abs(vaultPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(vaultsDir, entry.Name()))
		if err != nil {
			continue
		}

		var config VaultConfig
		if err := json.Unmarshal(data, &config); err != nil {
			continue
		}

		configAbsPath, err := filepath.Abs(config.VaultPath)
		if err != nil {
			continue
		}

		if configAbsPath == absPath {
			return &config, nil
		}
	}

	return nil, fmt.Errorf("no vault configuration found for path: %s", vaultPath)
}

// ListVaultConfigs returns all vault configurations
func ListVaultConfigs() ([]*VaultConfig, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	vaultsDir := filepath.Join(configDir, "vaults")
	entries, err := os.ReadDir(vaultsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*VaultConfig{}, nil
		}
		return nil, err
	}

	var configs []*VaultConfig
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(vaultsDir, entry.Name()))
		if err != nil {
			continue
		}

		var config VaultConfig
		if err := json.Unmarshal(data, &config); err != nil {
			continue
		}

		configs = append(configs, &config)
	}

	return configs, nil
}

// DeleteVaultConfig removes a vault configuration
func DeleteVaultConfig(vaultID string) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	path := filepath.Join(configDir, "vaults", fmt.Sprintf("%s.json", vaultID))
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// SavePublishConfig saves a publish configuration
func SavePublishConfig(siteID string, config *PublishConfig) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	sitesDir := filepath.Join(configDir, "sites")
	if err := os.MkdirAll(sitesDir, 0700); err != nil {
		return err
	}

	path := filepath.Join(sitesDir, fmt.Sprintf("%s.json", siteID))
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// LoadPublishConfigByPath finds a publish config by vault path
func LoadPublishConfigByPath(vaultPath string) (*PublishConfig, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	sitesDir := filepath.Join(configDir, "sites")
	entries, err := os.ReadDir(sitesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no sites configured")
		}
		return nil, err
	}

	absPath, err := filepath.Abs(vaultPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(sitesDir, entry.Name()))
		if err != nil {
			continue
		}

		var config PublishConfig
		if err := json.Unmarshal(data, &config); err != nil {
			continue
		}

		configAbsPath, err := filepath.Abs(config.VaultPath)
		if err != nil {
			continue
		}

		if configAbsPath == absPath {
			return &config, nil
		}
	}

	return nil, fmt.Errorf("no publish configuration found for path: %s", vaultPath)
}

// DeletePublishConfig removes a publish configuration
func DeletePublishConfig(siteID string) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	path := filepath.Join(configDir, "sites", fmt.Sprintf("%s.json", siteID))
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
