package config

import (
	"os"
	"runtime"
	"testing"
)

func TestBaseURL_FromEnv_OverridesDefault(t *testing.T) {
	t.Setenv("STACKDRIFT_URL", "http://192.168.1.47/")
	if got := BaseURL(); got != "http://192.168.1.47" {
		t.Fatalf("expected trimmed env url, got %q", got)
	}
}

func TestBaseURL_NoEnv_UsesDefault(t *testing.T) {
	t.Setenv("STACKDRIFT_URL", "")
	if got := BaseURL(); got != DefaultBaseURL {
		t.Fatalf("expected default %q, got %q", DefaultBaseURL, got)
	}
}

func TestSaveAndLoadProject_RoundTrips(t *testing.T) {
	t.Setenv("STACKDRIFT_HOME", t.TempDir())

	dir := t.TempDir()
	cfg := &ProjectConfig{
		ProjectID:   7,
		ProjectName: "Demo",
		Technologies: []TrackedTechnology{
			{Name: "Ubuntu", Version: "24.04", Category: "OperatingSystem"},
		},
	}

	if err := SaveProject(dir, cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	if loaded == nil {
		t.Fatal("expected a config")
	}
	if loaded.ProjectID != 7 || loaded.ProjectName != "Demo" {
		t.Fatalf("unexpected project: %+v", loaded)
	}
	if len(loaded.Technologies) != 1 || loaded.Technologies[0].Name != "Ubuntu" {
		t.Fatalf("unexpected technologies: %+v", loaded.Technologies)
	}
	if loaded.Version != 1 {
		t.Fatalf("expected version defaulted to 1, got %d", loaded.Version)
	}
}

func TestLoadProject_Missing_ReturnsNil(t *testing.T) {
	t.Setenv("STACKDRIFT_HOME", t.TempDir())

	loaded, err := LoadProject(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if loaded != nil {
		t.Fatal("expected nil for a directory with no config")
	}
}

func TestSaveProject_WritesNothingIntoTheScannedDirectory(t *testing.T) {
	t.Setenv("STACKDRIFT_HOME", t.TempDir())

	dir := t.TempDir()
	if err := SaveProject(dir, &ProjectConfig{ProjectID: 3, ProjectName: "Site"}); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("a scanned directory can be a public web root, expected it untouched, got %+v", entries)
	}
}

func TestSaveProject_LinkIsNotWorldReadable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix permission bits")
	}
	t.Setenv("STACKDRIFT_HOME", t.TempDir())

	if err := SaveProject(t.TempDir(), &ProjectConfig{ProjectID: 4}); err != nil {
		t.Fatal(err)
	}

	path, err := ProjectFilePath(4)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Fatalf("expected 0600, got %o", mode)
	}
}

func TestSaveProject_WithoutProjectID_Fails(t *testing.T) {
	t.Setenv("STACKDRIFT_HOME", t.TempDir())

	if err := SaveProject(t.TempDir(), &ProjectConfig{ProjectName: "No id"}); err == nil {
		t.Fatal("expected an error rather than a link stored under project 0")
	}
}

func TestLoadProject_LegacyFileInScanDir_IsMovedOut(t *testing.T) {
	t.Setenv("STACKDRIFT_HOME", t.TempDir())

	dir := t.TempDir()
	legacy := LegacyProjectFilePath(dir)
	body := `{"version":1,"projectId":9,"projectName":"Old","technologies":[{"name":"WordPress","version":"6.8.3","category":"Framework"}],"dependencyGroups":[]}`
	if err := os.WriteFile(legacy, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	if loaded == nil || loaded.ProjectID != 9 {
		t.Fatalf("expected the legacy link to be read, got %+v", loaded)
	}
	if !loaded.Migrated {
		t.Fatal("expected the move to be reported to the caller")
	}
	if len(loaded.Technologies) != 1 {
		t.Fatalf("expected tracked technologies to survive, got %+v", loaded.Technologies)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatal("expected the file to be removed from the scanned directory")
	}

	again, err := LoadProject(dir)
	if err != nil {
		t.Fatal(err)
	}
	if again == nil || again.ProjectID != 9 {
		t.Fatalf("expected the moved link to resolve, got %+v", again)
	}
	if again.Migrated {
		t.Fatal("expected the second load to be a plain read")
	}
}

func TestLoadProject_SameProjectFromTwoDirectories_ResolvesForBoth(t *testing.T) {
	t.Setenv("STACKDRIFT_HOME", t.TempDir())

	first, second := t.TempDir(), t.TempDir()
	cfg := &ProjectConfig{ProjectID: 11, ProjectName: "Shared"}

	if err := SaveProject(first, cfg); err != nil {
		t.Fatal(err)
	}
	if err := SaveProject(second, cfg); err != nil {
		t.Fatal(err)
	}

	for _, dir := range []string{first, second} {
		loaded, err := LoadProject(dir)
		if err != nil {
			t.Fatal(err)
		}
		if loaded == nil || loaded.ProjectID != 11 {
			t.Fatalf("expected %s to resolve to project 11, got %+v", dir, loaded)
		}
	}
}

func TestLoadProject_UnrelatedDirectory_DoesNotMatchAnotherProject(t *testing.T) {
	t.Setenv("STACKDRIFT_HOME", t.TempDir())

	if err := SaveProject(t.TempDir(), &ProjectConfig{ProjectID: 12}); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadProject(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if loaded != nil {
		t.Fatalf("expected no link for an unscanned directory, got %+v", loaded)
	}
}

func TestCredentialRoundTrip_IsolatedPerServer(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	if err := SaveCredential(Credential{BaseURL: "http://a", Token: "tok-a", Email: "a@x"}); err != nil {
		t.Fatal(err)
	}
	if err := SaveCredential(Credential{BaseURL: "http://b", Token: "tok-b"}); err != nil {
		t.Fatal(err)
	}

	a, err := LoadCredential("http://a")
	if err != nil || a == nil || a.Token != "tok-a" {
		t.Fatalf("expected tok-a, got %+v (err %v)", a, err)
	}

	if err := DeleteCredential("http://a"); err != nil {
		t.Fatal(err)
	}

	gone, err := LoadCredential("http://a")
	if err != nil {
		t.Fatal(err)
	}
	if gone != nil {
		t.Fatal("expected credential removed")
	}

	stillThere, err := LoadCredential("http://b")
	if err != nil || stillThere == nil || stillThere.Token != "tok-b" {
		t.Fatalf("expected tok-b to survive, got %+v", stillThere)
	}
}
