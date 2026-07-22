package commands

import (
	"fmt"
	"os"

	"github.com/digitalaffinity-au/stackdrift-cli/internal/config"
	"github.com/digitalaffinity-au/stackdrift-cli/internal/ui"
)

type CveFoundError struct {
	Technology int
	Dependency int
}

func (e *CveFoundError) Error() string {
	return fmt.Sprintf("%d technology CVEs and %d dependency CVEs found", e.Technology, e.Dependency)
}

func Check(args []string) error {
	client, _, err := authenticatedClient()
	if err != nil {
		return err
	}

	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := config.LoadProject(dir)
	if err != nil {
		return err
	}
	if cfg == nil || cfg.ProjectID == 0 {
		return errNoProjectLink
	}

	stats, err := client.GetProjectStats(cfg.ProjectID)
	if err != nil {
		return err
	}

	ui.Println("Project: " + cfg.ProjectName)
	ui.Printf("Technologies: %d (%d past end of life)\n", stats.TechnologyCount, stats.EndOfLifeCount)
	ui.Printf("Technology CVEs: %d\n", stats.TechnologyCveCount)
	ui.Printf("Dependencies: %d (%d vulnerable)\n", stats.DependencyCount, stats.VulnerableDependencyCount)
	ui.Printf("Dependency CVEs: %d\n", stats.DependencyCveCount)

	if stats.TechnologyCveCount > 0 || stats.DependencyCveCount > 0 {
		return &CveFoundError{Technology: stats.TechnologyCveCount, Dependency: stats.DependencyCveCount}
	}

	ui.Println("No known CVEs.")
	return nil
}
