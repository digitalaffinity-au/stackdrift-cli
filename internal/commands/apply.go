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

func primaryManifests(result *detect.Result) []detect.Manifest {
	var primaries []detect.Manifest
	for _, m := range result.Manifests {
		if m.Primary {
			primaries = append(primaries, m)
		}
	}
	return primaries
}

func manifestItems(scanDir string, primaries, all []detect.Manifest, existing *config.ProjectConfig) []ui.Item {
	tracked := trackedGroupNames(existing)
	items := make([]ui.Item, len(primaries))
	for i, m := range primaries {
		name := groupNameFor(scanDir, m)
		items[i] = ui.Item{
			Label:    name + " (" + ecosystemLabel(m.Ecosystem) + ")",
			Hint:     manifestHint(scanDir, m, all),
			Selected: !tracked[name],
		}
	}
	return items
}

func manifestHint(scanDir string, primary detect.Manifest, all []detect.Manifest) string {
	files := []string{manifestDisplay(scanDir, primary.Path)}
	for _, supporting := range supportingFor(primary, all) {
		files = append(files, filepath.Base(supporting.Path))
	}
	hint := strings.Join(files, " + ")
	if primary.Ecosystem == "Npm" && len(files) == 1 {
		hint += "  (no lock file, versions not pinned)"
	}
	return hint
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

func applyManifests(client *api.Client, projectID int, dir string, primaries, all []detect.Manifest, chosen []ui.Item, cfg *config.ProjectConfig, save func() error) error {
	tracked := trackedGroupNames(cfg)

	for i, item := range chosen {
		if !item.Selected {
			continue
		}
		primary := primaries[i]
		groupName := groupNameFor(dir, primary)
		if tracked[groupName] {
			continue
		}

		// The lock and central-version files pin the primary manifest's
		// versions, so they are always uploaded alongside it.
		bundle := append([]detect.Manifest{primary}, supportingFor(primary, all)...)
		files := make([]api.ManifestFile, len(bundle))
		names := make([]string, len(bundle))
		for j, m := range bundle {
			files[j] = api.ManifestFile{FileName: m.FileName, Content: m.Content}
			names[j] = manifestDisplay(dir, m.Path)
		}

		_, err := client.UploadManifests(projectID, api.UploadManifestsRequest{
			Ecosystem: primary.Ecosystem,
			GroupName: groupName,
			Files:     files,
		})
		if err != nil {
			return err
		}

		cfg.DependencyGrp = append(cfg.DependencyGrp, config.TrackedDependencyGroup{
			Name:      groupName,
			Ecosystem: primary.Ecosystem,
			Manifests: names,
		})
		tracked[groupName] = true
		if err := save(); err != nil {
			return err
		}
		ui.Println("  uploaded " + groupName + ": " + strings.Join(names, ", "))
	}
	return nil
}

func supportingFor(primary detect.Manifest, all []detect.Manifest) []detect.Manifest {
	primaryDir := filepath.Dir(primary.Path)
	var out []detect.Manifest
	for _, m := range all {
		if m.Primary || m.Ecosystem != primary.Ecosystem {
			continue
		}
		if filepath.Dir(m.Path) == primaryDir {
			out = append(out, m)
			continue
		}
		// Central NuGet package versions apply to every project in the tree.
		if primary.Ecosystem == "NuGet" && strings.EqualFold(m.FileName, "Directory.Packages.props") {
			out = append(out, m)
		}
	}
	return out
}

func trackedGroupNames(cfg *config.ProjectConfig) map[string]bool {
	keys := map[string]bool{}
	if cfg == nil {
		return keys
	}
	for _, group := range cfg.DependencyGrp {
		keys[group.Name] = true
	}
	return keys
}

func groupNameFor(scanDir string, m detect.Manifest) string {
	if strings.HasSuffix(strings.ToLower(m.FileName), ".csproj") {
		return strings.TrimSuffix(m.FileName, filepath.Ext(m.FileName))
	}

	base := filepath.Base(filepath.Dir(m.Path))
	if filepath.Dir(m.Path) == scanDir || base == "." || base == string(filepath.Separator) {
		base = baseName(scanDir)
	}
	return base + " " + ecosystemLabel(m.Ecosystem)
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
