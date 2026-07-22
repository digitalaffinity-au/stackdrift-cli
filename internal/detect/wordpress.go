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

	if !isDir(filepath.Join(core, "wp-admin")) {
		return
	}
	// Core is never legitimately installed inside another install's content
	// directory, so anything under one is a backup or a test fixture.
	if hasAncestor(root, core, "wp-content") {
		return
	}

	content, ok := readCapped(path)
	if !ok {
		return
	}
	match := wpVersionRe.FindStringSubmatch(content)
	if match == nil {
		return
	}

	result.Technologies = append(result.Technologies, Technology{
		Name:     "WordPress",
		Version:  cleanVersion(match[1]),
		Category: "Framework",
		Source:   wordPressSource(root, core),
	})
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
