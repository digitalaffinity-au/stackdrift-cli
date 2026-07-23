package detect

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The host web root search reads real directories, so it is off by default in
// tests. Otherwise every Scan test would walk whatever /var and /usr happen to
// hold on the machine running them, which is both slow and not reproducible.
func TestMain(m *testing.M) {
	webRootSearchRoots = nil
	os.Exit(m.Run())
}

func useSearchRoots(t *testing.T, roots ...string) {
	t.Helper()
	previous := webRootSearchRoots
	webRootSearchRoots = roots
	t.Cleanup(func() { webRootSearchRoots = previous })
}

func TestWebRootFindsWordPressOutsideTheScannedDirectory(t *testing.T) {
	server := t.TempDir()
	writeWordPress(t, server, filepath.Join("www", "html"), "6.8.1")
	useSearchRoots(t, server)

	result := &Result{}
	scanWebRoots(result)

	tech := onlyWordPress(t, result)
	if tech.Version != "6.8" || tech.Kernel != "6.8.1" {
		t.Fatalf("expected line 6.8 build 6.8.1, got %q / %q", tech.Version, tech.Kernel)
	}
	if !strings.Contains(tech.Source, "wp-includes/version.php") {
		t.Fatalf("expected the source to name the install, got %q", tech.Source)
	}
}

// A web root install describes the machine, not the directory that was scanned,
// so it must be offered unticked exactly like the OS and kernel are.
func TestWebRootWordPressIsAHostDetection(t *testing.T) {
	server := t.TempDir()
	writeWordPress(t, server, filepath.Join("www", "html"), "6.8.1")

	result := &Result{}
	searchWebRoot(result, server)

	if !IsHostSource(onlyWordPress(t, result).Source) {
		t.Fatal("a web root install must count as a host detection")
	}
}

func TestWebRootFindsHostingPanelLayout(t *testing.T) {
	server := t.TempDir()
	writeWordPress(t, server, filepath.Join("www", "vhosts", "example.com", "httpdocs"), "6.7.2")

	result := &Result{}
	searchWebRoot(result, server)

	tech := onlyWordPress(t, result)
	if tech.Version != "6.7" || tech.Kernel != "6.7.2" {
		t.Fatalf("expected line 6.7 build 6.7.2, got %q / %q", tech.Version, tech.Kernel)
	}
}

func TestWebRootStopsAtTheDepthLimit(t *testing.T) {
	server := t.TempDir()
	deep := "."
	for i := 0; i < webRootMaxDepth+2; i++ {
		deep = filepath.Join(deep, "d")
	}
	writeWordPress(t, server, deep, "6.8.1")

	result := &Result{}
	searchWebRoot(result, server)

	if len(result.Technologies) != 0 {
		t.Fatalf("expected nothing past the depth limit, got %d", len(result.Technologies))
	}
}

func TestWebRootSkipsHugeSystemTrees(t *testing.T) {
	if !webRootSkipPaths["/var/lib"] || !webRootSkipPaths["/usr/lib"] {
		t.Fatal("the trees that dominate a walk must stay pruned")
	}
	if webRootSkipPaths["/var/www"] || webRootSkipPaths["/usr/share"] {
		t.Fatal("web roots must not be pruned")
	}
}

func TestWebRootIgnoresVersionFileWithoutWpAdmin(t *testing.T) {
	server := t.TempDir()
	write(t, server, filepath.Join("www", "wp-includes", "version.php"), "<?php\n$wp_version = '6.8.1';\n")

	result := &Result{}
	searchWebRoot(result, server)

	if len(result.Technologies) != 0 {
		t.Fatal("a stray version.php is not an install")
	}
}

func TestWebRootIgnoresBackupCopyUnderUploads(t *testing.T) {
	server := t.TempDir()
	writeWordPress(t, server, filepath.Join("www", "html"), "6.8.1")
	writeWordPress(t, server, filepath.Join("www", "html", "wp-content", "uploads", "backup"), "5.2.0")

	result := &Result{}
	searchWebRoot(result, server)

	if onlyWordPress(t, result).Kernel != "6.8.1" {
		t.Fatal("a backup under uploads must not be reported")
	}
}

func TestWebRootMissingDirectoryIsNotAnError(t *testing.T) {
	result := &Result{}
	searchWebRoot(result, filepath.Join(t.TempDir(), "does-not-exist"))

	if len(result.Technologies) != 0 {
		t.Fatal("expected nothing from a missing root")
	}
}

// The directory being scanned is the project, so its own detection must keep the
// relative source and stay ticked by default even when the same install is also
// reachable from a web root.
func TestScannedDirectoryDetectionWinsOverTheWebRootCopy(t *testing.T) {
	tree := []Technology{{Name: "WordPress", Version: "6.8.1", Source: "wp-includes/version.php"}}
	web := []Technology{{Name: "WordPress", Version: "6.8.1", Source: SourceHostPrefix + "/var/www/html/wp-includes/version.php"}}

	merged := dedupeTechnologies(append(tree, web...))

	if len(merged) != 1 {
		t.Fatalf("expected one entry, got %d", len(merged))
	}
	if IsHostSource(merged[0].Source) {
		t.Fatal("the scanned directory detection must win")
	}
}

func onlyWordPress(t *testing.T, result *Result) Technology {
	t.Helper()
	var found []Technology
	for _, tech := range result.Technologies {
		if tech.Name == "WordPress" {
			found = append(found, tech)
		}
	}
	if len(found) != 1 {
		t.Fatalf("expected exactly one WordPress detection, got %d", len(found))
	}
	return found[0]
}

// The build is what makes a WordPress finding actionable: without it the server
// scores the site against the base of its line and reports every advisory
// already fixed in that line as affecting a fully patched install.
func TestWebRootReportsTheExactBuildNotJustTheLine(t *testing.T) {
	server := t.TempDir()
	writeWordPress(t, server, filepath.Join("www", "html"), "6.8.3")

	result := &Result{}
	searchWebRoot(result, server)

	tech := onlyWordPress(t, result)
	if tech.Version != "6.8" {
		t.Fatalf("expected the 6.8 line, got %q", tech.Version)
	}
	if tech.Kernel != "6.8.3" {
		t.Fatalf("expected the 6.8.3 build, got %q", tech.Kernel)
	}
}

// An auto-update leaves an extracted core under wp-content/upgrade. The tree
// scanner has always ignored those; the web root search must agree or it invents
// an install that is not running anywhere.
func TestWebRootIgnoresCoreUnderWpContent(t *testing.T) {
	server := t.TempDir()
	writeWordPress(t, server, filepath.Join("www", "html"), "6.8.3")
	writeWordPress(t, server, filepath.Join("www", "html", "wp-content", "upgrade", "wordpress"), "5.2.0")

	result := &Result{}
	searchWebRoot(result, server)

	if onlyWordPress(t, result).Kernel != "6.8.3" {
		t.Fatal("core under wp-content is not a running install")
	}
}

// The same install reached both ways must land on ONE row, because two rows
// sharing a name and version also share a tracking key: unticking either would
// delete the technology the other still represents.
func TestSameInstallFoundBothWaysCollapsesToOneRow(t *testing.T) {
	tree := []Technology{{Name: "WordPress", Version: "6.8", Kernel: "6.8.3", Source: "wp-includes/version.php"}}
	web := []Technology{{Name: "WordPress", Version: "6.8", Kernel: "6.8.3", Source: SourceHostPrefix + "/var/www/html/wp-includes/version.php"}}

	merged := dedupeTechnologies(append(tree, web...))

	if len(merged) != 1 {
		t.Fatalf("expected one row, got %d: %+v", len(merged), merged)
	}
	if IsHostSource(merged[0].Source) {
		t.Fatal("the scanned directory detection must win")
	}
}

// Whichever detection knows the build wins, so a Dockerfile naming the host's
// own release cannot shadow /etc/os-release and drop the running kernel.
func TestDedupeKeepsTheBuildFromWhicheverDetectionHasIt(t *testing.T) {
	merged := dedupeTechnologies([]Technology{
		{Name: "Ubuntu", Version: "24.04", Source: "Dockerfile"},
		{Name: "Ubuntu", Version: "24.04", Kernel: "6.8.0-60", Source: SourceOsRelease},
	})

	if len(merged) != 1 {
		t.Fatalf("expected one row, got %d", len(merged))
	}
	if merged[0].Source != "Dockerfile" {
		t.Fatalf("the tree detection must still win the source, got %q", merged[0].Source)
	}
	if merged[0].Kernel != "6.8.0-60" {
		t.Fatalf("the running kernel must survive, got %q", merged[0].Kernel)
	}
}
