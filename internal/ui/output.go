package ui

import (
	"fmt"
	"strings"

	"github.com/penguinpowernz/obsidian-headless-go/internal/config"
)

// PrintVaultConfig displays vault configuration in a readable format
func PrintVaultConfig(cfg *config.VaultConfig) {
	fmt.Printf("  Vault: %s (%s)\n", cfg.VaultName, cfg.VaultID)
	fmt.Printf("  Location: %s\n", cfg.VaultPath)

	mode := cfg.SyncMode
	if mode == "" {
		mode = config.SyncModeBidirectional
	}
	fmt.Printf("  Sync mode: %s\n", mode)
	fmt.Printf("  Conflict strategy: %s\n", cfg.ConflictStrategy)

	if cfg.DeviceName != "" {
		fmt.Printf("  Device name: %s\n", cfg.DeviceName)
	}
	if cfg.ConfigDir != "" {
		fmt.Printf("  Config directory: %s\n", cfg.ConfigDir)
	}
	if len(cfg.AllowTypes) > 0 {
		fmt.Printf("  File types: %s\n", strings.Join(cfg.AllowTypes, ", "))
	}
	if len(cfg.AllowSpecialFiles) > 0 {
		fmt.Printf("  Configs: %s\n", strings.Join(cfg.AllowSpecialFiles, ", "))
	}
	if len(cfg.IgnoreFolders) > 0 {
		fmt.Printf("  Excluded folders: %s\n", strings.Join(cfg.IgnoreFolders, ", "))
	}
}

// PrintPublishConfig displays publish configuration in a readable format
func PrintPublishConfig(cfg *config.PublishConfig) {
	fmt.Printf("  Site: %s\n", cfg.SiteID)
	fmt.Printf("  Host: %s\n", cfg.Host)
	fmt.Printf("  Location: %s\n", cfg.VaultPath)

	if len(cfg.Includes) > 0 {
		fmt.Printf("  Includes: %s\n", strings.Join(cfg.Includes, ", "))
	}
	if len(cfg.Excludes) > 0 {
		fmt.Printf("  Excludes: %s\n", strings.Join(cfg.Excludes, ", "))
	}
}

// PrintVaultList displays a list of vaults
func PrintVaultList(vaults []struct{ ID, Name, Region string }, title string) {
	if len(vaults) == 0 {
		return
	}

	fmt.Printf("\n%s:\n", title)
	for _, v := range vaults {
		if v.Region != "" {
			fmt.Printf("  %s  \"%s\"  (%s)\n", v.ID, v.Name, v.Region)
		} else {
			fmt.Printf("  %s  \"%s\"\n", v.ID, v.Name)
		}
	}
}

// PrintSiteList displays a list of publish sites
func PrintSiteList(sites []struct{ ID string }, title string) {
	if len(sites) == 0 {
		return
	}

	fmt.Printf("\n%s:\n", title)
	for _, s := range sites {
		fmt.Printf("  %s\n", s.ID)
	}
}

// Success prints a success message
func Success(msg string) {
	fmt.Println(msg)
}

// Info prints an informational message
func Info(msg string) {
	fmt.Println(msg)
}

// Warning prints a warning message
func Warning(msg string) {
	fmt.Printf("Warning: %s\n", msg)
}
