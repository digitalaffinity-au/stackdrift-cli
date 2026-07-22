package detect

import (
	"path/filepath"
	"testing"
)

func primaryPaths(result *Result) []string {
	var paths []string
	for _, m := range result.Manifests {
		if m.Primary {
			paths = append(paths, m.Path)
		}
	}
	return paths
}

func TestScan_LibManClientLibraries_AreNotProjects(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "App.csproj", "<Project><TargetFramework>net8.0</TargetFramework></Project>")
	for _, lib := range []string{"bootstrap", "apexcharts", "bootstrap-table"} {
		write(t, dir, filepath.Join("wwwroot", "lib", lib, "package.json"), `{"name":"`+lib+`","version":"1.0.0"}`)
	}

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range primaryPaths(result) {
		if filepath.Base(path) == "package.json" {
			t.Fatalf("a vendored client library is not a project, got %q", path)
		}
	}
}

func TestScan_ScopedClientLibraries_AreNotProjects(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, filepath.Join("wwwroot", "lib", "@fullcalendar", "core", "package.json"), `{"name":"@fullcalendar/core"}`)
	write(t, dir, filepath.Join("wwwroot", "lib", "@highlightjs", "cdn-assets", "es", "package.json"), `{"type":"module"}`)

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	if paths := primaryPaths(result); len(paths) != 0 {
		t.Fatalf("expected nothing under wwwroot/lib, got %v", paths)
	}
}

func TestScan_ApplicationManifestBesideVendoredLibs_Survives(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "package.json", `{"name":"my-app"}`)
	write(t, dir, "package-lock.json", `{"lockfileVersion":3}`)
	write(t, dir, filepath.Join("wwwroot", "lib", "bootstrap", "package.json"), `{"name":"bootstrap"}`)

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	paths := primaryPaths(result)
	if len(paths) != 1 {
		t.Fatalf("expected only the application manifest, got %v", paths)
	}
	if filepath.Dir(paths[0]) != dir {
		t.Fatalf("expected the manifest at the scan root, got %q", paths[0])
	}
	if !hasManifest(result.Manifests, "package-lock.json", "Npm") {
		t.Fatal("expected the application lock still bundled")
	}
}

func TestScan_PlainLibDirectory_IsStillSource(t *testing.T) {
	// lib is an ordinary source folder outside wwwroot, so skipping it by name
	// alone would hide real projects.
	dir := t.TempDir()
	write(t, dir, filepath.Join("lib", "package.json"), `{"name":"my-lib"}`)

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(primaryPaths(result)) != 1 {
		t.Fatal("expected a plain lib directory to still be scanned")
	}
}

func TestScan_BowerComponents_AreNotProjects(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, filepath.Join("bower_components", "jquery", "package.json"), `{"name":"jquery"}`)

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	if paths := primaryPaths(result); len(paths) != 0 {
		t.Fatalf("expected bower_components skipped, got %v", paths)
	}
}

func TestIsVendorDir_MatchesRegardlessOfCase(t *testing.T) {
	if !isVendorDir(filepath.Join("app", "wwwroot", "lib")) {
		t.Fatal("expected wwwroot/lib to be vendored")
	}
	if !isVendorDir(filepath.Join("app", "wwwroot", "Lib")) {
		t.Fatal("expected wwwroot/Lib to be vendored")
	}
	if !isVendorDir(filepath.Join("app", "Content", "lib")) {
		t.Fatal("expected Content/lib to be vendored")
	}
	if !isVendorDir(filepath.Join("app", "Bower_Components")) {
		t.Fatal("expected bower_components to be vendored regardless of case")
	}
	if isVendorDir(filepath.Join("app", "src", "lib")) {
		t.Fatal("a lib folder outside wwwroot is source, not vendored")
	}
}

func TestScan_PublicVendors_AreNotProjects(t *testing.T) {
	dir := t.TempDir()
	for _, lib := range []string{"bootstrap", "echarts", "Chart.js"} {
		write(t, dir, filepath.Join("public", "vendors", lib, "package.json"), `{"name":"`+lib+`"}`)
	}
	write(t, dir, "package.json", `{"name":"my-app"}`)

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	paths := primaryPaths(result)
	if len(paths) != 1 || filepath.Dir(paths[0]) != dir {
		t.Fatalf("expected only the application manifest, got %v", paths)
	}
}

func TestScan_WordPressPluginsAndThemes_AreNotProjects(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, filepath.Join("wp-content", "plugins", "woocommerce", "package.json"), `{"name":"woocommerce"}`)
	write(t, dir, filepath.Join("wp-content", "themes", "mytheme", "package.json"), `{"name":"mytheme"}`)

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	if paths := primaryPaths(result); len(paths) != 0 {
		t.Fatalf("plugin and theme build tooling is not the site's dependencies, got %v", paths)
	}
}

func TestScan_WordPressCoreStillDetected_WithPluginsSkipped(t *testing.T) {
	dir := t.TempDir()
	writeWordPress(t, dir, ".", "6.8.3")
	write(t, dir, filepath.Join("wp-content", "plugins", "woocommerce", "package.json"), `{"name":"woocommerce"}`)

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := findTech(result.Technologies, "WordPress"); !ok {
		t.Fatal("skipping plugins must not affect core detection")
	}
}

func TestScan_DocsAndPublishDirectories_AreSkipped(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, filepath.Join("docs", "template", "package.json"), `{"name":"template"}`)
	write(t, dir, filepath.Join("publish", "package.json"), `{"name":"output"}`)
	write(t, dir, "package.json", `{"name":"my-app"}`)

	result, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	paths := primaryPaths(result)
	if len(paths) != 1 || filepath.Dir(paths[0]) != dir {
		t.Fatalf("expected only the application manifest, got %v", paths)
	}
}

func TestScan_GenericVendorConventions_AreSkipped(t *testing.T) {
	for _, vendor := range []string{
		filepath.Join("assets", "vendor"),
		filepath.Join("assets", "plugins"),
		filepath.Join("assets", "libs"),
		filepath.Join("static", "vendor"),
		filepath.Join("public", "vendor"),
		"third_party",
		"thirdparty",
		"external",
	} {
		dir := t.TempDir()
		write(t, dir, filepath.Join(vendor, "somelib", "package.json"), `{"name":"somelib"}`)

		result, err := Scan(dir)
		if err != nil {
			t.Fatal(err)
		}
		if paths := primaryPaths(result); len(paths) != 0 {
			t.Fatalf("expected %s to be treated as vendored, got %v", vendor, paths)
		}
	}
}
