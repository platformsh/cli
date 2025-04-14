package alt

import (
	"context"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/platformsh/cli/internal/config"
	"github.com/platformsh/cli/internal/state"
	"github.com/platformsh/cli/internal/version"
)

// ShouldUpdate returns whether the Update function may be run on configuration.
func ShouldUpdate(cnf *config.Config) bool {
	return cnf.Updates.Check &&
		cnf.SourceFile != "" &&
		cnf.Metadata.URL != ""
}

// Update checks for configuration updates, when appropriate.
// The "cnf" pointer will NOT be updated with the new configuration.
func Update(ctx context.Context, cnf *config.Config, debugLog func(fmt string, i ...any)) error {
	s, err := state.Load(cnf)
	if err != nil {
		return err
	}
	interval := time.Second * time.Duration(cnf.Updates.CheckInterval)
	lastChecked := time.Unix(s.ConfigUpdates.LastChecked, 0)
	if time.Since(lastChecked) < interval {
		debugLog("Config updates checked recently (%v ago)", time.Since(lastChecked).Truncate(time.Second))
		return nil
	}

	if cnf.SourceFile == "" {
		return fmt.Errorf("no config file path available")
	}
	if cnf.Metadata.URL == "" {
		return fmt.Errorf("no config URL available")
	}

	stat, err := os.Stat(cnf.SourceFile)
	if err != nil {
		return fmt.Errorf("could not stat config file %s: %w", cnf.SourceFile, err)
	}
	if time.Since(stat.ModTime()) < interval {
		debugLog("Config file updated recently (%v ago): %s",
			time.Since(stat.ModTime()).Truncate(time.Second), cnf.SourceFile)
		return nil
	}

	defer func() {
		s.ConfigUpdates.LastChecked = time.Now().Unix()
		if err := state.Save(s, cnf); err != nil {
			debugLog("Error saving state: %s", err)
		}
	}()

	debugLog("Checking for config updates from URL: %s", cnf.Metadata.URL)
	newCnfNode, newCnfStruct, err := FetchConfig(ctx, cnf.Metadata.URL)
	if err != nil {
		return err
	}
	if !newCnfStruct.Metadata.UpdatedAt.IsZero() &&
		!newCnfStruct.Metadata.UpdatedAt.After(cnf.Metadata.UpdatedAt) {
		debugLog("Config is already up to date (updated at %v)", cnf.Metadata.UpdatedAt.Format(time.RFC3339))
		return nil
	}
	if newCnfStruct.Metadata.Version != "" {
		cmp, err := version.Compare(cnf.Metadata.Version, newCnfStruct.Metadata.Version)
		if err != nil {
			return fmt.Errorf("could not compare config versions: %w", err)
		}
		if cmp >= 0 {
			debugLog("Config is already up to date (version %s)", cnf.Metadata.Version)
			return nil
		}
	}
	b, err := yaml.Marshal(newCnfNode)
	if err != nil {
		return err
	}

	if err := writeFile(cnf.SourceFile, b, 0, 0o644); err != nil {
		return err
	}
	debugLog("Automatically updated config file: %s", cnf.SourceFile)

	return nil
}
