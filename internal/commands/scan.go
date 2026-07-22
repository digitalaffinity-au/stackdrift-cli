package commands

import (
	"errors"
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

	return scan(client, dir, assumeYes)
}

func scan(client *api.Client, dir string, assumeYes bool) error {
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

	cfg := configFor(project, existing)

	// The project can be edited on the website between scans, so what the
	// server holds decides what is tracked. Without this a technology removed
	// there stays listed locally, is shown as already tracked, and is silently
	// skipped instead of being added back.
	if err := reconcileTracked(client, project, cfg); err != nil {
		return err
	}

	primaries := primaryManifests(result)
	techItems := technologyItems(result, cfg)
	manifestItems := manifestItems(dir, primaries, result.Manifests, cfg)

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
	ui.Println("Linked this directory to project '" + project.Name + "' and updated it.")
	if path, err := config.ProjectFilePath(project.ID); err == nil {
		ui.Println("Link saved to " + path)
	}
	return nil
}

func resolveProject(client *api.Client, dir string, assumeYes bool) (*api.Project, *config.ProjectConfig, error) {
	existing, err := config.LoadProject(dir)
	if err != nil {
		return nil, nil, err
	}

	if existing != nil && existing.Migrated {
		ui.Println("Moved the project link out of " + dir + " so it is not exposed by a web server.")
		ui.Println("The old " + config.ProjectFileName + " file was removed. If it is in git, commit that deletion.")
	}

	if existing != nil && existing.ProjectID != 0 {
		project, err := client.GetProject(existing.ProjectID)
		if err == nil {
			ui.Println("Using linked project '" + project.Name + "'.")
			return project, existing, nil
		}
		if !isNotFound(err) {
			return nil, nil, err
		}
		ui.Println("The linked project no longer exists. Choose another.")
		existing.Technologies = nil
		existing.DependencyGrp = nil
	}

	if assumeYes {
		return nil, nil, errors.New("this directory is not linked to a project yet, run scan once without --yes to choose one")
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
