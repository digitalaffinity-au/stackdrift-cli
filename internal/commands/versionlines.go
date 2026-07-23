package commands

import (
	"strings"

	"github.com/digitalaffinity-au/stackdrift-cli/internal/api"
	"github.com/digitalaffinity-au/stackdrift-cli/internal/detect"
)

// The catalog decides what a release line looks like, and the shape is not
// consistent between technologies: WordPress lines are 7.0, .NET lines are 8.0,
// and Laravel lines are 11 for current releases but 5.8 for older ones. A rule
// hardcoded here is wrong for whichever technology it was not written for, and
// sending a version that matches no line leaves the project showing something the
// catalog cannot reason about. So the line is resolved against what the server
// actually holds.
//
// Only the version is adjusted. Whether a build is known is the detector's
// business: version.php reports the exact build a site runs, while a composer
// constraint or a target framework only ever names the line.
func resolveVersionLines(client *api.Client, result *detect.Result) {
	known := make(map[string][]string)

	for i := range result.Technologies {
		tech := &result.Technologies[i]
		if tech.Version == "" {
			continue
		}

		versions, looked := known[tech.Name]
		if !looked {
			// A lookup failure must not fail the scan. Leaving the detected
			// version alone is the same behaviour as before this existed.
			versions, _ = client.GetVersions(tech.Name)
			known[tech.Name] = versions
		}

		if line := matchVersionLine(tech.Version, versions); line != "" {
			tech.Version = line
		}
	}
}

// matchVersionLine mirrors how the server matches a version to a release: an
// exact hit wins, otherwise the longest known line the detected version sits
// under. A version under no known line is left alone, because that is a genuine
// signal that the release is not one the catalog tracks any more.
func matchVersionLine(version string, known []string) string {
	best := ""
	for _, candidate := range known {
		if strings.EqualFold(candidate, version) {
			return candidate
		}
		if strings.HasPrefix(version, candidate+".") && len(candidate) > len(best) {
			best = candidate
		}
	}
	return best
}
