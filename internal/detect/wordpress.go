package detect

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var wpVersionRe = regexp.MustCompile(`\$wp_version\s*=\s*['"]([^'"]+)['"]`)

// WordPress can be installed anywhere: at the web root, in a subdirectory, at
// web/wp under Bedrock, or wherever composer was told to put core. What never
// moves is the shape of core itself, so detection anchors on
// <core>/wp-includes/version.php with a sibling wp-admin rather than on any
// particular location.
func detectWordPress(result *Result, root, path string) {
	core := filepath.Dir(filepath.Dir(path))

	// Core is never legitimately installed inside another install's content
	// directory, so anything under one is a backup or a test fixture.
	if hasAncestor(root, core, "wp-content") {
		return
	}

	version, ok := wordPressAt(path)
	if !ok {
		return
	}

	line, build := wordPressLine(version)
	result.Technologies = append(result.Technologies, Technology{
		Name:     "WordPress",
		Version:  line,
		Kernel:   build,
		Category: "Framework",
		Source:   wordPressSource(root, core),
	})
}

// version.php reports the exact build, but StackDrift tracks WordPress the way
// it tracks a distribution: the release line carries the support dates, and the
// build says which point release is installed. Sending 7.0.2 as the version
// matches no release line, so the line is derived and the full version is kept
// as the build. A version already at line granularity is its own build, which
// is what the catalog lists for an unpatched release.
func wordPressLine(version string) (line, build string) {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return version, version
	}
	return parts[0] + "." + parts[1], version
}

// wordPressAt reads the version out of a wp-includes/version.php, confirming
// first that it belongs to a real install rather than a stray copy of the file.
func wordPressAt(versionFile string) (string, bool) {
	core := filepath.Dir(filepath.Dir(versionFile))
	if !isDir(filepath.Join(core, "wp-admin")) {
		return "", false
	}

	content, ok := readCapped(versionFile)
	if !ok {
		return "", false
	}

	match := wpVersionRe.FindStringSubmatch(content)
	if match == nil {
		return "", false
	}

	return cleanVersion(match[1]), true
}

func isWordPressVersionFile(path, name string) bool {
	return strings.EqualFold(name, "version.php") &&
		strings.EqualFold(filepath.Base(filepath.Dir(path)), "wp-includes")
}

// Uploads holds no core files but does hold whatever the backup plugins put
// there, which is often a complete copy of an older install. It is also the
// largest directory on most sites by a wide margin.
func isWordPressUploads(path string) bool {
	return strings.EqualFold(filepath.Base(path), "uploads") &&
		strings.EqualFold(filepath.Base(filepath.Dir(path)), "wp-content")
}

// wordPressSource names the install by where its core sits, so several
// installs in one tree stay distinguishable in the selection list.
func wordPressSource(root, core string) string {
	const suffix = "wp-includes/version.php"

	rel, err := filepath.Rel(root, core)
	if err != nil || rel == "." {
		return suffix
	}
	return filepath.ToSlash(rel) + "/" + suffix
}

func hasAncestor(root, dir, name string) bool {
	for current := dir; current != root && current != filepath.Dir(current); current = filepath.Dir(current) {
		if strings.EqualFold(filepath.Base(current), name) {
			return true
		}
	}
	return false
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
