package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/digitalaffinity-au/stackdrift-cli/internal/api"
	"github.com/digitalaffinity-au/stackdrift-cli/internal/config"
	"github.com/digitalaffinity-au/stackdrift-cli/internal/detect"
	"github.com/digitalaffinity-au/stackdrift-cli/internal/ui"
)

func Scan(args []string) error {
	assumeYes := hasFlag(args, "--yes", "-y")

	client, _, err := authenticatedClient()
	if err != nil {
		return err
	}

	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	project, existing, err := resolveProject(client, dir, assumeYes)
	if err != nil {
		return err
	}

	ui.Println()
	ui.Println("Scanning " + dir + " ...")
	result, err := detect.Scan(dir)
	if err != nil {
		return err
	}

	if len(result.Technologies) == 0 && len(result.Manifests) == 0 {
		ui.Println("No supported technologies or dependency manifests were found here.")
		return nil
	}

	primaries := primaryManifests(result)
	techItems := technologyItems(result, existing)
	manifestItems := manifestItems(dir, primaries, existing)

	var chosenTechs, chosenManifests []ui.Item
	if assumeYes {
		// Accept the recommended defaults without prompting. Host-machine
		// detections stay off, so a project scan never adds the dev machine's OS.
		chosenTechs = techItems
		chosenManifests = manifestItems
	} else {
		chosenTechs = ui.ToggleList("Technologies detected:", techItems)
		chosenManifests = ui.ToggleList("Dependency projects detected:", manifestItems)
	}

	cfg := configFor(project, existing)

	// Persist the project link before mutating, so an interrupted run still
	// remembers the project and does not re-add what already succeeded.
	if err := config.SaveProject(dir, cfg); err != nil {
		return err
	}

	saveConfig := func() error { return config.SaveProject(dir, cfg) }

	if err := applyTechnologies(client, project.ID, result.Technologies, chosenTechs, cfg, saveConfig); err != nil {
		return err
	}
	if err := applyManifests(client, project.ID, dir, primaries, result.Manifests, chosenManifests, cfg, saveConfig); err != nil {
		return err
	}

	if err := config.SaveProject(dir, cfg); err != nil {
		return err
	}

	ui.Println()
	ui.Println("Saved " + config.ProjectFileName + " and updated project '" + project.Name + "'.")
	return nil
}

func resolveProject(client *api.Client, dir string, assumeYes bool) (*api.Project, *config.ProjectConfig, error) {
	existing, err := config.LoadProject(dir)
	if err != nil {
		return nil, nil, err
	}

	if existing != nil && existing.ProjectID != 0 {
		project, err := client.GetProject(existing.ProjectID)
		if err == nil {
			ui.Println("Using project '" + project.Name + "' from " + config.ProjectFileName + ".")
			return project, existing, nil
		}
		if !isNotFound(err) {
			return nil, nil, err
		}
		ui.Println("The project recorded in " + config.ProjectFileName + " no longer exists. Choose another.")
		existing.Technologies = nil
		existing.DependencyGrp = nil
	}

	if assumeYes {
		return nil, nil, fmt.Errorf("no %s here yet, run scan once without --yes to choose a project", config.ProjectFileName)
	}

	project, err := chooseProject(client)
	if err != nil {
		return nil, nil, err
	}
	return project, existing, nil
}

func hasFlag(args []string, names ...string) bool {
	for _, arg := range args {
		for _, name := range names {
			if arg == name {
				return true
			}
		}
	}
	return false
}

func configFor(project *api.Project, existing *config.ProjectConfig) *config.ProjectConfig {
	cfg := existing
	if cfg == nil {
		cfg = &config.ProjectConfig{}
	}
	cfg.ProjectID = project.ID
	cfg.ProjectName = project.Name
	return cfg
}

func manifestDisplay(dir, path string) string {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return path
	}
	return rel
}
