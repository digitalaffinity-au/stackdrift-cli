package detect

import (
	"os"
	"path/filepath"
	"testing"
)

func writeWordPress(t *testing.T, dir, core, version string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, core, "wp-admin"), 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, dir, filepath.Join(core, "wp-includes", "version.php"),
		"<?php\n$wp_version = '"+version+"';\n$wp_db_version = 60421;\n")
}

func findAll(techs []Technology, name string) []Technology {
	var out []Technology
	for _, tech := range techs {
		if tech.Name == name {
			out = append(out, tech)
		}
	}
	return out
}

func TestScan_WordPressAtRoot_DetectsVersion(t *testing.T) {
	dir := t.TempDir()
	writeWordPress(t, dir, ".", "6.8.3")

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	tech, ok := findTech(result.Technologies, "WordPress")
	if !ok {
		t.Fatal("expected WordPress")
	}
	if tech.Version != "6.8" {
		t.Fatalf("expected the 6.8 line, got %q", tech.Version)
	}
	if tech.Kernel != "6.8.3" {
		t.Fatalf("expected the 6.8.3 build, got %q", tech.Kernel)
	}
	if tech.Category != "Framework" {
		t.Fatalf("expected Framework, got %q", tech.Category)
	}
}

func TestScan_WordPressInSubdirectory_IsFound(t *testing.T) {
	dir := t.TempDir()
	writeWordPress(t, dir, "blog", "6.9.5")

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	tech, ok := findTech(result.Technologies, "WordPress")
	if !ok {
		t.Fatal("expected WordPress in a subdirectory install")
	}
	if tech.Version != "6.9" || tech.Kernel != "6.9.5" {
		t.Fatalf("expected line 6.9 build 6.9.5, got %q / %q", tech.Version, tech.Kernel)
	}
	if tech.Source != "blog/wp-includes/version.php" {
		t.Fatalf("expected the source to name the install location, got %q", tech.Source)
	}
}

func TestScan_BedrockLayout_IsFound(t *testing.T) {
	dir := t.TempDir()
	writeWordPress(t, dir, filepath.Join("web", "wp"), "7.0.2")
	write(t, dir, "composer.json", `{"require":{"roots/bedrock-autoloader":"^1.0"}}`)

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	tech, ok := findTech(result.Technologies, "WordPress")
	if !ok {
		t.Fatal("expected WordPress under web/wp")
	}
	if tech.Version != "7.0" || tech.Kernel != "7.0.2" {
		t.Fatalf("expected line 7.0 build 7.0.2, got %q / %q", tech.Version, tech.Kernel)
	}
}

func TestScan_VersionFileWithoutWpAdmin_IsIgnored(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, filepath.Join("plugin", "tests", "fixtures", "wp-includes", "version.php"),
		"<?php\n$wp_version = '4.1';\n")

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := findTech(result.Technologies, "WordPress"); ok {
		t.Fatal("a version.php with no sibling wp-admin is a fixture, not an install")
	}
}

func TestScan_BackupCopyUnderUploads_IsIgnored(t *testing.T) {
	dir := t.TempDir()
	writeWordPress(t, dir, ".", "6.8.3")
	writeWordPress(t, dir, filepath.Join("wp-content", "uploads", "backup", "site"), "5.4.1")

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	found := findAll(result.Technologies, "WordPress")
	if len(found) != 1 {
		t.Fatalf("expected only the live install, got %d: %+v", len(found), found)
	}
	if found[0].Kernel != "6.8.3" {
		t.Fatalf("expected the live 6.8.3 build, got %q", found[0].Kernel)
	}
}

func TestScan_CoreUnderWpContent_IsIgnored(t *testing.T) {
	dir := t.TempDir()
	writeWordPress(t, dir, ".", "6.8.3")
	writeWordPress(t, dir, filepath.Join("wp-content", "plugins", "bundled"), "5.9")

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	found := findAll(result.Technologies, "WordPress")
	if len(found) != 1 {
		t.Fatalf("core inside wp-content is never a real install, got %d: %+v", len(found), found)
	}
}

func TestScan_TwoInstallsAtDifferentVersions_ReportsBoth(t *testing.T) {
	dir := t.TempDir()
	writeWordPress(t, dir, "current", "6.9.5")
	writeWordPress(t, dir, "wp-old", "5.4.1")

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	found := findAll(result.Technologies, "WordPress")
	if len(found) != 2 {
		t.Fatalf("expected both installs, got %d: %+v", len(found), found)
	}

	sources := map[string]string{}
	for _, tech := range found {
		sources[tech.Kernel] = tech.Source
	}
	if sources["5.4.1"] != "wp-old/wp-includes/version.php" {
		t.Fatalf("expected the stale copy to name its path, got %q", sources["5.4.1"])
	}
	if sources["6.9.5"] != "current/wp-includes/version.php" {
		t.Fatalf("expected the live install to name its path, got %q", sources["6.9.5"])
	}
}

func TestScan_TwoInstallsAtSameVersion_CollapseToOne(t *testing.T) {
	dir := t.TempDir()
	writeWordPress(t, dir, "site-a", "6.8.3")
	writeWordPress(t, dir, "site-b", "6.8.3")

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	if found := findAll(result.Technologies, "WordPress"); len(found) != 1 {
		t.Fatalf("same version should dedupe to one entry, got %d", len(found))
	}
}

func TestScan_PrereleaseVersion_UsesNumericLead(t *testing.T) {
	dir := t.TempDir()
	writeWordPress(t, dir, ".", "7.1-beta2")

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	tech, ok := findTech(result.Technologies, "WordPress")
	if !ok {
		t.Fatal("expected WordPress")
	}
	if tech.Version != "7.1" {
		t.Fatalf("expected 7.1, got %q", tech.Version)
	}
}

func TestScan_WordPressIsNotAHostSource(t *testing.T) {
	dir := t.TempDir()
	writeWordPress(t, dir, ".", "6.8.3")

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	tech, _ := findTech(result.Technologies, "WordPress")
	if IsHostSource(tech.Source) {
		t.Fatal("a WordPress install under the scan root is the project, not the host")
	}
}

// StackDrift tracks WordPress by release line with the exact build alongside, the
// same shape a distribution uses, so version.php's build has to be split rather
// than sent as the version. A version already at line granularity is its own
// build, which is what the catalog lists for an unpatched release.
func TestWordPressLineSplit(t *testing.T) {
	cases := []struct{ detected, line, build string }{
		{"7.0.2", "7.0", "7.0.2"},
		{"6.8.3", "6.8", "6.8.3"},
		{"6.8", "6.8", "6.8"},
		{"7", "7", "7"},
	}
	for _, c := range cases {
		line, build := wordPressLine(c.detected)
		if line != c.line || build != c.build {
			t.Fatalf("%s: expected %s / %s, got %s / %s", c.detected, c.line, c.build, line, build)
		}
	}
}

// Two installs on the same line collapse, because a project tracks one entry per
// line and apply keys tracking on name and version. Reporting both would add only
// the first anyway, so the merge happens here where it is visible.
func TestScan_TwoInstallsOnTheSameLine_CollapseToOne(t *testing.T) {
	dir := t.TempDir()
	writeWordPress(t, dir, "site-a", "6.8.3")
	writeWordPress(t, dir, "site-b", "6.8.1")

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	found := findAll(result.Technologies, "WordPress")
	if len(found) != 1 {
		t.Fatalf("expected one entry for the 6.8 line, got %d: %+v", len(found), found)
	}
}
