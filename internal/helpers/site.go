package helpers

import (
	"fmt"
	"path/filepath"

	"github.com/penguinpowernz/obsidian-headless-go/internal/api"
	"github.com/penguinpowernz/obsidian-headless-go/internal/config"
)

// LoadPublishByPath loads a publish config by path
func LoadPublishByPath(path string) (*config.PublishConfig, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	publishConfig, err := config.LoadPublishConfigByPath(absPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return publishConfig, nil
}

// FindSiteByID finds a site by ID from a list
func FindSiteByID(sites []api.Site, siteID string) (*api.Site, error) {
	for i, s := range sites {
		if s.ID == siteID {
			return &sites[i], nil
		}
	}
	return nil, fmt.Errorf("site %q not found", siteID)
}

// UpdatePublishConfigFromFlags updates publish config based on flags
type PublishConfigUpdater struct {
	Config  *config.PublishConfig
	Updated bool
}

// SetIncludes sets the includes list if provided
func (u *PublishConfigUpdater) SetIncludes(value string) {
	if value == "" {
		return // Not provided
	}

	if value == "" {
		u.Config.Includes = nil
	} else {
		u.Config.Includes = splitCSV(value)
	}

	u.Updated = true
}

// SetExcludes sets the excludes list if provided
func (u *PublishConfigUpdater) SetExcludes(value string) {
	if value == "" {
		return // Not provided
	}

	if value == "" {
		u.Config.Excludes = nil
	} else {
		u.Config.Excludes = splitCSV(value)
	}

	u.Updated = true
}
