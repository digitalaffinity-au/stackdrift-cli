package config

import (
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
	loaded, err := LoadProject(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if loaded != nil {
		t.Fatal("expected nil for a directory with no config")
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
