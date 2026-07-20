package commands

import (
	"path/filepath"
	"strings"

	"github.com/digitalaffinity-au/stackdrift-cli/internal/api"
	"github.com/digitalaffinity-au/stackdrift-cli/internal/config"
	"github.com/digitalaffinity-au/stackdrift-cli/internal/detect"
	"github.com/digitalaffinity-au/stackdrift-cli/internal/ui"
)

func technologyItems(result *detect.Result, existing *config.ProjectConfig) []ui.Item {
	tracked := trackedTechKeys(existing)
	items := make([]ui.Item, len(result.Technologies))
	for i, t := range result.Technologies {
		label := t.Name
		if t.Version != "" {
			label += " " + t.Version
		}
		// Host-machine detections describe where the CLI runs, not the project,
		// so they are offered but not selected by default.
		items[i] = ui.Item{
			Label:    label,
			Hint:     t.Source,
			Selected: !tracked[techKey(t.Name, t.Version)] && !detect.IsHostSource(t.Source),
		}
	}
	return items
}

func manifestItems(result *detect.Result, existing *config.ProjectConfig) []ui.Item {
	tracked := trackedManifests(existing)
	items := make([]ui.Item, len(result.Manifests))
	for i, m := range result.Manifests {
		key := manifestKey(m.Ecosystem, m.FileName)
		items[i] = ui.Item{
			Label:    m.FileName + " (" + ecosystemLabel(m.Ecosystem) + ")",
			Hint:     m.Path,
			Selected: !tracked[key],
		}
	}
	return items
}

func applyTechnologies(client *api.Client, projectID int, detected []detect.Technology, chosen []ui.Item, cfg *config.ProjectConfig, save func() error) error {
	tracked := trackedTechKeys(cfg)
	for i, item := range chosen {
		if !item.Selected {
			continue
		}
		t := detected[i]
		if tracked[techKey(t.Name, t.Version)] {
			continue
		}

		_, err := client.AddTechnology(projectID, api.AddTechnologyRequest{
			Name:     t.Name,
			Version:  t.Version,
			Category: t.Category,
		})
		if err != nil {
			return err
		}

		cfg.Technologies = append(cfg.Technologies, config.TrackedTechnology{
			Name:     t.Name,
			Version:  t.Version,
			Category: t.Category,
		})
		tracked[techKey(t.Name, t.Version)] = true
		if err := save(); err != nil {
			return err
		}
		ui.Println("  added technology: " + label(t.Name, t.Version))
	}
	return nil
}

func applyManifests(client *api.Client, projectID int, dir string, detected []detect.Manifest, chosen []ui.Item, cfg *config.ProjectConfig, save func() error) error {
	tracked := trackedManifests(cfg)

	byEcosystem := map[string][]detect.Manifest{}
	for i, item := range chosen {
		if !item.Selected {
			continue
		}
		m := detected[i]
		if tracked[manifestKey(m.Ecosystem, m.FileName)] {
			continue
		}
		byEcosystem[m.Ecosystem] = append(byEcosystem[m.Ecosystem], m)
	}

	for ecosystem, manifests := range byEcosystem {
		files := make([]api.ManifestFile, len(manifests))
		names := make([]string, len(manifests))
		for i, m := range manifests {
			files[i] = api.ManifestFile{FileName: m.FileName, Content: m.Content}
			names[i] = manifestDisplay(dir, m.Path)
		}

		groupName := defaultGroupName(dir, ecosystem)
		_, err := client.UploadManifests(projectID, api.UploadManifestsRequest{
			Ecosystem: ecosystem,
			GroupName: groupName,
			Files:     files,
		})
		if err != nil {
			return err
		}

		cfg.DependencyGrp = append(cfg.DependencyGrp, config.TrackedDependencyGroup{
			Name:      groupName,
			Ecosystem: ecosystem,
			Manifests: names,
		})
		if err := save(); err != nil {
			return err
		}
		ui.Println("  uploaded " + ecosystemLabel(ecosystem) + " manifests: " + strings.Join(names, ", "))
	}
	return nil
}

func trackedManifests(cfg *config.ProjectConfig) map[string]bool {
	keys := map[string]bool{}
	if cfg == nil {
		return keys
	}
	for _, group := range cfg.DependencyGrp {
		for _, name := range group.Manifests {
			keys[manifestKey(group.Ecosystem, filepath.Base(name))] = true
		}
	}
	return keys
}

func manifestKey(ecosystem, fileName string) string {
	return ecosystem + "|" + strings.ToLower(fileName)
}

func trackedTechKeys(cfg *config.ProjectConfig) map[string]bool {
	keys := map[string]bool{}
	if cfg == nil {
		return keys
	}
	for _, t := range cfg.Technologies {
		keys[techKey(t.Name, t.Version)] = true
	}
	return keys
}

func techKey(name, version string) string {
	return strings.ToLower(name) + "|" + version
}

func label(name, version string) string {
	if version == "" {
		return name
	}
	return name + " " + version
}

func ecosystemLabel(ecosystem string) string {
	if ecosystem == "Npm" {
		return "npm"
	}
	return ecosystem
}

func defaultGroupName(dir, ecosystem string) string {
	base := baseName(dir)
	return base + " " + ecosystemLabel(ecosystem)
}
