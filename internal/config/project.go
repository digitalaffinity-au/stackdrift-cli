package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const ProjectFileName = ".stackdrift"

type TrackedTechnology struct {
	Name     string `json:"name"`
	Version  string `json:"version,omitempty"`
	Category string `json:"category"`
	ID       int    `json:"id,omitempty"`
}

type TrackedDependencyGroup struct {
	Name      string   `json:"name"`
	Ecosystem string   `json:"ecosystem"`
	Manifests []string `json:"manifests"`
}

type ProjectConfig struct {
	Version       int                      `json:"version"`
	ProjectID     int                      `json:"projectId"`
	ProjectName   string                   `json:"projectName"`
	Technologies  []TrackedTechnology      `json:"technologies"`
	DependencyGrp []TrackedDependencyGroup `json:"dependencyGroups"`
}

func ProjectFilePath(dir string) string {
	return filepath.Join(dir, ProjectFileName)
}

func LoadProject(dir string) (*ProjectConfig, error) {
	data, err := os.ReadFile(ProjectFilePath(dir))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg ProjectConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveProject(dir string, cfg *ProjectConfig) error {
	if cfg.Version == 0 {
		cfg.Version = 1
	}
	if cfg.Technologies == nil {
		cfg.Technologies = []TrackedTechnology{}
	}
	if cfg.DependencyGrp == nil {
		cfg.DependencyGrp = []TrackedDependencyGroup{}
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ProjectFilePath(dir), append(data, '\n'), 0o644)
}
