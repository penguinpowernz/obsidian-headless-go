package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ob",
	Short: "Obsidian Headless - CLI for Obsidian Sync and Publish",
	Long: `Headless client for Obsidian Sync and Obsidian Publish.
Sync and publish your vaults from the command line without the desktop app.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(syncListRemoteCmd)
	rootCmd.AddCommand(syncListLocalCmd)
	rootCmd.AddCommand(syncCreateRemoteCmd)
	rootCmd.AddCommand(syncSetupCmd)
	rootCmd.AddCommand(syncConfigCmd)
	rootCmd.AddCommand(syncStatusCmd)
	rootCmd.AddCommand(syncUnlinkCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(publishListSitesCmd)
	rootCmd.AddCommand(publishCreateSiteCmd)
	rootCmd.AddCommand(publishSetupCmd)
	rootCmd.AddCommand(publishCmd)
	rootCmd.AddCommand(publishConfigCmd)
	rootCmd.AddCommand(publishUnlinkCmd)
	rootCmd.AddCommand(publishSiteOptionsCmd)
}
