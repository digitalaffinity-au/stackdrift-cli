package detect

import (
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"
)

// WordPress is served from a web root rather than sitting in whatever directory
// the scan happens to be started from, so a scan run from a home directory would
// otherwise miss the site the machine exists to serve. These trees are searched
// for core installs no matter where the scan runs.
var webRootSearchRoots = []string{"/srv", "/var", "/usr", "/opt"}

// Deep enough for the layouts hosting panels produce, such as
// /var/www/vhosts/example.com/httpdocs/wp-includes, without opening the door to
// walking an entire filesystem.
const webRootMaxDepth = 6

// Trees that never hold a web root but can be enormous. Walking /var/lib/docker
// or /usr/lib on a busy machine costs more than the rest of the search put
// together, and finds nothing. Matched by absolute path rather than by directory
// name so that a harmless "lib" inside a site is still searched.
var webRootSkipPaths = map[string]bool{
	"/var/lib":          true,
	"/var/cache":        true,
	"/var/log":          true,
	"/var/spool":        true,
	"/var/tmp":          true,
	"/var/backups":      true,
	"/var/crash":        true,
	"/usr/bin":          true,
	"/usr/sbin":         true,
	"/usr/include":      true,
	"/usr/lib":          true,
	"/usr/lib32":        true,
	"/usr/lib64":        true,
	"/usr/libexec":      true,
	"/usr/src":          true,
	"/usr/share/man":    true,
	"/usr/share/doc":    true,
	"/usr/share/locale": true,
	"/usr/share/icons":  true,
	"/usr/share/fonts":  true,
}

// scanWebRoots looks for WordPress core anywhere under the machine's web roots.
// Findings are marked as host detections, because they describe the machine
// rather than the directory that was scanned, so they are listed but left
// unticked until the user says otherwise.
func scanWebRoots(result *Result) {
	if runtime.GOOS != "linux" {
		return
	}

	for _, root := range webRootSearchRoots {
		searchWebRoot(result, root)
	}
}

func searchWebRoot(result *Result, root string) {
	rootDepth := strings.Count(filepath.Clean(root), string(filepath.Separator))

	// Errors are skipped rather than surfaced: a scan run without root will hit
	// directories it cannot read, and that must not stop it finding the rest.
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() {
			if isWordPressVersionFile(path, d.Name()) {
				addWebRootWordPress(result, path)
			}
			return nil
		}

		if path == root {
			return nil
		}

		if webRootSkipPaths[path] || skipDirs[d.Name()] || strings.HasPrefix(d.Name(), ".") ||
			isWordPressUploads(path) || isVendorDir(path) {
			return filepath.SkipDir
		}

		if strings.Count(path, string(filepath.Separator))-rootDepth >= webRootMaxDepth {
			return filepath.SkipDir
		}

		return nil
	})
}

func addWebRootWordPress(result *Result, versionFile string) {
	version, ok := wordPressAt(versionFile)
	if !ok {
		return
	}

	result.Technologies = append(result.Technologies, Technology{
		Name:     "WordPress",
		Version:  version,
		Category: "Framework",
		Source:   SourceHostPrefix + versionFile,
	})
}
