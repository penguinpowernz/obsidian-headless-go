package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/penguinpowernz/obsidian-headless-go/internal/api"
	"github.com/penguinpowernz/obsidian-headless-go/internal/config"
	"github.com/penguinpowernz/obsidian-headless-go/internal/helpers"
	"github.com/penguinpowernz/obsidian-headless-go/internal/ui"
	"github.com/spf13/cobra"
)

var (
	siteID   string
	slug     string
	dryRun   bool
	yes      bool
	all      bool
	includes string
	excludes string
)

var publishListSitesCmd = &cobra.Command{
	Use:   "publish-list-sites",
	Short: "List all publish sites available to your account",
	RunE:  runPublishListSites,
}

var publishCreateSiteCmd = &cobra.Command{
	Use:   "publish-create-site",
	Short: "Create a new publish site",
	RunE:  runPublishCreateSite,
}

var publishSetupCmd = &cobra.Command{
	Use:   "publish-setup",
	Short: "Connect a local vault to a publish site",
	RunE:  runPublishSetup,
}

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish vault changes to a site",
	RunE:  runPublish,
}

var publishConfigCmd = &cobra.Command{
	Use:   "publish-config",
	Short: "View or change publish settings for a vault",
	RunE:  runPublishConfig,
}

var publishUnlinkCmd = &cobra.Command{
	Use:   "publish-unlink",
	Short: "Disconnect a vault from a publish site",
	RunE:  runPublishUnlink,
}

var publishSiteOptionsCmd = &cobra.Command{
	Use:   "publish-site-options",
	Short: "View or update remote site options",
	RunE:  runPublishSiteOptions,
}

func init() {
	// publish-create-site flags
	publishCreateSiteCmd.Flags().StringVar(&slug, "slug", "", "Site slug (required)")
	publishCreateSiteCmd.MarkFlagRequired("slug")

	// publish-setup flags
	publishSetupCmd.Flags().StringVar(&siteID, "site", "", "Site ID or slug (required)")
	publishSetupCmd.Flags().StringVar(&vaultPath, "path", ".", "Local vault path")
	publishSetupCmd.MarkFlagRequired("site")

	// publish flags
	publishCmd.Flags().StringVar(&vaultPath, "path", ".", "Local vault path")
	publishCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show changes without publishing")
	publishCmd.Flags().BoolVar(&yes, "yes", false, "Publish without confirmation")
	publishCmd.Flags().BoolVar(&all, "all", false, "Include files without a publish flag")

	// publish-config flags
	publishConfigCmd.Flags().StringVar(&vaultPath, "path", ".", "Local vault path")
	publishConfigCmd.Flags().StringVar(&includes, "includes", "", "Folders to include (comma-separated)")
	publishConfigCmd.Flags().StringVar(&excludes, "excludes", "", "Folders to exclude (comma-separated)")

	// publish-unlink flags
	publishUnlinkCmd.Flags().StringVar(&vaultPath, "path", ".", "Local vault path")

	// publish-site-options flags
	publishSiteOptionsCmd.Flags().StringVar(&vaultPath, "path", ".", "Local vault path")
}

func runPublishListSites(cmd *cobra.Command, args []string) error {
	token, err := helpers.RequireAuth()
	if err != nil {
		return err
	}

	client := api.NewClient(token)
	response, err := client.ListSites()
	if err != nil {
		return fmt.Errorf("fetch sites: %w", err)
	}

	if len(response.Sites) == 0 && len(response.Shared) == 0 {
		ui.Info("No publish sites found.")
		return nil
	}

	var sites []struct{ ID string }
	for _, s := range response.Sites {
		sites = append(sites, struct{ ID string }{s.ID})
	}
	ui.PrintSiteList(sites, "Sites")

	if len(response.Shared) > 0 {
		var shared []struct{ ID string }
		for _, s := range response.Shared {
			shared = append(shared, struct{ ID string }{s.ID})
		}
		ui.PrintSiteList(shared, "Shared sites")
	}

	return nil
}

func runPublishCreateSite(cmd *cobra.Command, args []string) error {
	token, err := requireAuth()
	if err != nil {
		return err
	}

	client := api.NewClient(token)

	site, err := client.CreateSite()
	if err != nil {
		return fmt.Errorf("create site: %w", err)
	}

	// Set the slug
	err = client.SetSlug(site.ID, site.Host, slug)
	if err != nil {
		return fmt.Errorf("set slug: %w", err)
	}

	fmt.Println("\nSite created successfully!")
	fmt.Printf("  Site ID: %s\n", site.ID)
	fmt.Printf("  Slug: %s\n", slug)
	fmt.Printf("  Host: %s\n", site.Host)
	fmt.Printf("\nRun 'ob publish-setup --site \"%s\"' to connect a vault.\n", site.ID)

	return nil
}

func runPublishSetup(cmd *cobra.Command, args []string) error {
	token, err := requireAuth()
	if err != nil {
		return err
	}

	client := api.NewClient(token)

	absPath, err := filepath.Abs(vaultPath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	// Fetch site info
	fmt.Println("Fetching sites...")
	siteList, err := client.ListSites()
	if err != nil {
		return fmt.Errorf("fetch sites: %w", err)
	}

	allSites := append(siteList.Sites, siteList.Shared...)

	// Find site by ID
	var selectedSite *api.Site
	for i, s := range allSites {
		if s.ID == siteID {
			selectedSite = &allSites[i]
			break
		}
	}

	if selectedSite == nil {
		return fmt.Errorf("site %q not found", siteID)
	}

	// Save configuration
	publishConfig := &config.PublishConfig{
		SiteID:    selectedSite.ID,
		Host:      selectedSite.Host,
		VaultPath: absPath,
	}

	if err := config.SavePublishConfig(selectedSite.ID, publishConfig); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Println("\nSite connected successfully!")
	printPublishConfig(publishConfig)

	// Create directory if needed
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		if err := os.MkdirAll(absPath, 0755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
	}

	fmt.Println("\nRun 'ob publish --dry-run' to preview changes.")
	return nil
}

func runPublish(cmd *cobra.Command, args []string) error {
	publishConfig, err := helpers.LoadPublishByPath(vaultPath)
	if err != nil {
		return err
	}

	token, err := helpers.RequireAuth()
	if err != nil {
		return err
	}

	_ = token
	_ = publishConfig

	// TODO: Implement actual publish logic
	// This would involve:
	// 1. Scanning vault for files to publish
	// 2. Reading frontmatter to check publish: true/false
	// 3. Computing file hashes
	// 4. Comparing with remote state
	// 5. Uploading new/changed files
	// 6. Removing deleted files

	ui.Info("Publish functionality not yet fully implemented.")
	ui.Info("This requires reverse-engineering the Obsidian publish protocol from the obfuscated JavaScript.")

	return nil
}

func runPublishConfig(cmd *cobra.Command, args []string) error {
	publishConfig, err := helpers.LoadPublishByPath(vaultPath)
	if err != nil {
		return err
	}

	updater := &helpers.PublishConfigUpdater{Config: publishConfig}
	updater.SetIncludes(includes)
	updater.SetExcludes(excludes)

	if updater.Updated {
		if err := config.SavePublishConfig(publishConfig.SiteID, publishConfig); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
		ui.Info("Configuration updated:")
	} else {
		ui.Info("Current configuration:")
	}

	ui.PrintPublishConfig(publishConfig)
	return nil
}

func runPublishUnlink(cmd *cobra.Command, args []string) error {
	publishConfig, err := helpers.LoadPublishByPath(vaultPath)
	if err != nil {
		return err
	}

	if err := config.DeletePublishConfig(publishConfig.SiteID); err != nil {
		return fmt.Errorf("delete config: %w", err)
	}

	ui.Success(fmt.Sprintf("Publish configuration removed for %s", publishConfig.VaultPath))
	return nil
}

func runPublishSiteOptions(cmd *cobra.Command, args []string) error {
	absPath, err := filepath.Abs(vaultPath)
	if err != nil {
		return fmt.Errorf("resolve path: %w", err)
	}

	publishConfig, err := config.LoadPublishConfigByPath(absPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	token, err := requireAuth()
	if err != nil {
		return err
	}

	_ = token
	_ = publishConfig

	fmt.Println("Site options functionality not yet fully implemented.")
	return nil
}

// printPublishConfig is deprecated - use ui.PrintPublishConfig instead
// Kept for now to avoid breaking changes
func printPublishConfig(cfg *config.PublishConfig) {
	ui.PrintPublishConfig(cfg)
}
