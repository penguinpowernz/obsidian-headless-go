package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/penguinpowernz/obsidian-headless-go/internal/api"
	"github.com/penguinpowernz/obsidian-headless-go/internal/config"
	"github.com/penguinpowernz/obsidian-headless-go/internal/crypto"
	"github.com/penguinpowernz/obsidian-headless-go/internal/helpers"
	"github.com/penguinpowernz/obsidian-headless-go/internal/ui"
	"github.com/spf13/cobra"
)

var (
	vaultPath      string
	vaultName      string
	vaultID        string
	deviceName     string
	configDirName  string
	encPassword    string
	encryption     string
	region         string
	continuous     bool
	conflictStrat  string
	syncMode       string
	excludedFolders string
	fileTypes      string
	configs        string
)

var syncListRemoteCmd = &cobra.Command{
	Use:   "sync-list-remote",
	Short: "List all remote vaults available to your account",
	RunE:  runSyncListRemote,
}

var syncListLocalCmd = &cobra.Command{
	Use:   "sync-list-local",
	Short: "List locally configured vaults",
	RunE:  runSyncListLocal,
}

var syncCreateRemoteCmd = &cobra.Command{
	Use:   "sync-create-remote",
	Short: "Create a new remote vault",
	RunE:  runSyncCreateRemote,
}

var syncSetupCmd = &cobra.Command{
	Use:   "sync-setup",
	Short: "Setup sync between a local vault and a remote vault",
	RunE:  runSyncSetup,
}

var syncConfigCmd = &cobra.Command{
	Use:   "sync-config",
	Short: "View or change sync settings for a vault",
	RunE:  runSyncConfig,
}

var syncStatusCmd = &cobra.Command{
	Use:   "sync-status",
	Short: "Show sync status and configuration for a vault",
	RunE:  runSyncStatus,
}

var syncUnlinkCmd = &cobra.Command{
	Use:   "sync-unlink",
	Short: "Disconnect a vault from sync",
	RunE:  runSyncUnlink,
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync a vault",
	RunE:  runSync,
}

func init() {
	// sync-create-remote flags
	syncCreateRemoteCmd.Flags().StringVar(&vaultName, "name", "", "Vault name (required)")
	syncCreateRemoteCmd.Flags().StringVar(&encryption, "encryption", "", "Encryption type: standard or e2ee")
	syncCreateRemoteCmd.Flags().StringVar(&encPassword, "password", "", "End-to-end encryption password")
	syncCreateRemoteCmd.Flags().StringVar(&region, "region", "", "Vault region")
	syncCreateRemoteCmd.MarkFlagRequired("name")

	// sync-setup flags
	syncSetupCmd.Flags().StringVar(&vaultID, "vault", "", "Remote vault ID or name (required)")
	syncSetupCmd.Flags().StringVar(&vaultPath, "path", ".", "Local vault path")
	syncSetupCmd.Flags().StringVar(&encPassword, "password", "", "End-to-end encryption password")
	syncSetupCmd.Flags().StringVar(&deviceName, "device-name", "", "Device name")
	syncSetupCmd.Flags().StringVar(&configDirName, "config-dir", ".obsidian", "Config directory name")
	syncSetupCmd.MarkFlagRequired("vault")

	// sync-config flags
	syncConfigCmd.Flags().StringVar(&vaultPath, "path", ".", "Local vault path")
	syncConfigCmd.Flags().StringVar(&syncMode, "mode", "", "Sync mode: bidirectional, pull-only, or mirror-remote")
	syncConfigCmd.Flags().StringVar(&conflictStrat, "conflict-strategy", "", "Conflict strategy: merge or conflict")
	syncConfigCmd.Flags().StringVar(&fileTypes, "file-types", "", "File types to sync (comma-separated)")
	syncConfigCmd.Flags().StringVar(&configs, "configs", "", "Config categories to sync (comma-separated)")
	syncConfigCmd.Flags().StringVar(&excludedFolders, "excluded-folders", "", "Folders to exclude (comma-separated)")
	syncConfigCmd.Flags().StringVar(&deviceName, "device-name", "", "Device name")
	syncConfigCmd.Flags().StringVar(&configDirName, "config-dir", "", "Config directory name")

	// sync-status flags
	syncStatusCmd.Flags().StringVar(&vaultPath, "path", ".", "Local vault path")

	// sync-unlink flags
	syncUnlinkCmd.Flags().StringVar(&vaultPath, "path", ".", "Local vault path")

	// sync flags
	syncCmd.Flags().StringVar(&vaultPath, "path", ".", "Local vault path")
	syncCmd.Flags().BoolVar(&continuous, "continuous", false, "Run in continuous sync mode")
}

func runSyncListRemote(cmd *cobra.Command, args []string) error {
	token, err := helpers.RequireAuth()
	if err != nil {
		return err
	}

	client := api.NewClient(token)
	response, err := client.ListVaults(3) // encryption version 3
	if err != nil {
		return fmt.Errorf("fetch vaults: %w", err)
	}

	if len(response.Vaults) == 0 && len(response.Shared) == 0 {
		ui.Info("No vaults found.")
		return nil
	}

	// Convert to simple struct for UI
	var vaults []struct{ ID, Name, Region string }
	for _, v := range response.Vaults {
		vaults = append(vaults, struct{ ID, Name, Region string }{v.ID, v.Name, v.Region})
	}
	ui.PrintVaultList(vaults, "Vaults")

	if len(response.Shared) > 0 {
		var shared []struct{ ID, Name, Region string }
		for _, v := range response.Shared {
			shared = append(shared, struct{ ID, Name, Region string }{v.ID, v.Name, v.Region})
		}
		ui.PrintVaultList(shared, "Shared vaults")
	}

	return nil
}

func runSyncListLocal(cmd *cobra.Command, args []string) error {
	vaults, err := config.ListVaultConfigs()
	if err != nil {
		return fmt.Errorf("list vaults: %w", err)
	}

	if len(vaults) == 0 {
		fmt.Println("No vaults configured.")
		return nil
	}

	fmt.Println("Configured vaults:")
	for _, v := range vaults {
		fmt.Printf("  %s\n", v.VaultName)
		fmt.Printf("    Path: %s\n", v.VaultPath)
		fmt.Printf("    Host: %s\n", v.Host)
	}

	return nil
}

func runSyncCreateRemote(cmd *cobra.Command, args []string) error {
	token, err := requireAuth()
	if err != nil {
		return err
	}

	client := api.NewClient(token)

	var salt string
	var encVersion int
	if encryption != "standard" && encPassword != "" {
		// Generate salt for E2EE
		salt, err = crypto.GenerateSalt()
		if err != nil {
			return fmt.Errorf("generate salt: %w", err)
		}
		encVersion = 1 // Placeholder for encryption version
	}

	createReq := map[string]interface{}{
		"name":   vaultName,
		"region": region,
	}

	if salt != "" {
		createReq["salt"] = salt
		createReq["encryption_version"] = encVersion
	}

	var vault api.Vault
	err = client.Request("POST", "/vaults", createReq, &vault)
	if err != nil {
		return fmt.Errorf("create vault: %w", err)
	}

	fmt.Println("\nVault created successfully!")
	fmt.Printf("  Vault ID: %s\n", vault.ID)
	fmt.Printf("  Vault name: %s\n", vault.Name)
	fmt.Printf("  Region: %s\n", vault.Region)
	if encPassword != "" {
		fmt.Println("  Encryption: end-to-end")
	} else {
		fmt.Println("  Encryption: managed")
	}
	fmt.Printf("\nRun 'ob sync-setup --vault \"%s\"' to configure sync.\n", vault.ID)

	return nil
}

func runSyncSetup(cmd *cobra.Command, args []string) error {
	token, err := requireAuth()
	if err != nil {
		return err
	}

	client := api.NewClient(token)

	// Resolve vault path
	absPath, err := filepath.Abs(vaultPath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	// Fetch vault info
	fmt.Println("Fetching vault info...")
	vaultList, err := client.ListVaults(3)
	if err != nil {
		return fmt.Errorf("fetch vaults: %w", err)
	}

	allVaults := append(vaultList.Vaults, vaultList.Shared...)

	// Find vault by ID or name
	var selectedVault *api.Vault
	for i, v := range allVaults {
		if v.ID == vaultID || v.Name == vaultID {
			selectedVault = &allVaults[i]
			break
		}
	}

	if selectedVault == nil {
		return fmt.Errorf("vault %q not found", vaultID)
	}

	// Handle encryption password
	if selectedVault.Salt != "" && encPassword == "" {
		fmt.Print("End-to-end encryption password: ")
		pwBytes, err := readPassword()
		if err != nil {
			return err
		}
		encPassword = string(pwBytes)
	}

	var encKey []byte
	if encPassword != "" {
		encKey, err = crypto.DeriveKey(encPassword, selectedVault.Salt)
		if err != nil {
			return fmt.Errorf("derive key: %w", err)
		}
	}

	// Save configuration
	vaultConfig := &config.VaultConfig{
		VaultID:           selectedVault.ID,
		VaultName:         selectedVault.Name,
		VaultPath:         absPath,
		Host:              selectedVault.Host,
		EncryptionVersion: selectedVault.EncryptionVersion,
		EncryptionSalt:    selectedVault.Salt,
		ConflictStrategy:  config.ConflictStrategyMerge,
		DeviceName:        deviceName,
		ConfigDir:         configDirName,
	}

	if encKey != nil {
		vaultConfig.EncryptionKey = crypto.EncodeKey(encKey)
	}

	if err := config.SaveVaultConfig(selectedVault.ID, vaultConfig); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Println("\nVault configured successfully!")
	printVaultConfig(vaultConfig)

	// Create directory if needed
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		if err := os.MkdirAll(absPath, 0755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
	}

	fmt.Println("\nRun 'ob sync' to start syncing.")
	return nil
}

func runSyncConfig(cmd *cobra.Command, args []string) error {
	vaultConfig, err := helpers.LoadVaultByPath(vaultPath)
	if err != nil {
		return err
	}

	updater := &helpers.VaultConfigUpdater{Config: vaultConfig}

	if err := updater.SetSyncMode(syncMode); err != nil {
		return err
	}

	if err := updater.SetConflictStrategy(conflictStrat); err != nil {
		return err
	}

	updater.SetStringSlice(excludedFolders, &vaultConfig.IgnoreFolders)
	updater.SetStringSlice(fileTypes, &vaultConfig.AllowTypes)
	updater.SetStringSlice(configs, &vaultConfig.AllowSpecialFiles)
	updater.SetString(deviceName, &vaultConfig.DeviceName)
	updater.SetString(configDirName, &vaultConfig.ConfigDir)

	if updater.Updated {
		if err := config.SaveVaultConfig(vaultConfig.VaultID, vaultConfig); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
		ui.Info("Configuration updated:")
	} else {
		ui.Info("Current configuration:")
	}

	ui.PrintVaultConfig(vaultConfig)
	return nil
}

func runSyncStatus(cmd *cobra.Command, args []string) error {
	vaultConfig, err := helpers.LoadVaultByPath(vaultPath)
	if err != nil {
		return err
	}

	ui.Info("Sync Configuration:")
	ui.PrintVaultConfig(vaultConfig)
	return nil
}

func runSyncUnlink(cmd *cobra.Command, args []string) error {
	vaultConfig, err := helpers.LoadVaultByPath(vaultPath)
	if err != nil {
		return err
	}

	if err := config.DeleteVaultConfig(vaultConfig.VaultID); err != nil {
		return fmt.Errorf("delete config: %w", err)
	}

	ui.Success(fmt.Sprintf("Sync configuration removed for %s", vaultConfig.VaultPath))
	return nil
}

func runSync(cmd *cobra.Command, args []string) error {
	vaultConfig, err := helpers.LoadVaultByPath(vaultPath)
	if err != nil {
		return err
	}

	token, err := helpers.RequireAuth()
	if err != nil {
		return err
	}

	ui.Info("Starting sync:")
	ui.PrintVaultConfig(vaultConfig)

	// TODO: Implement actual sync logic
	// This would involve:
	// 1. Connecting to the sync server
	// 2. Downloading remote changes
	// 3. Uploading local changes
	// 4. Handling conflicts based on strategy
	// 5. If continuous mode, watch for file changes

	_ = token
	ui.Info("\nSync functionality not yet fully implemented.")
	ui.Info("This requires reverse-engineering the Obsidian sync protocol from the obfuscated JavaScript.")

	return nil
}

// printVaultConfig is deprecated - use ui.PrintVaultConfig instead
// Kept for now to avoid breaking changes
func printVaultConfig(cfg *config.VaultConfig) {
	ui.PrintVaultConfig(cfg)
}
