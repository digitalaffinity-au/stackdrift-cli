package commands

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/digitalaffinity-au/stackdrift-cli/internal/api"
	"github.com/digitalaffinity-au/stackdrift-cli/internal/config"
	"github.com/digitalaffinity-au/stackdrift-cli/internal/detect"
)

func TestApplyTechnologies_AddsOnlyTheSelected(t *testing.T) {
	var added []api.AddTechnologyRequest
	client := serve(t, func(w http.ResponseWriter, r *http.Request) {
		var req api.AddTechnologyRequest
		_ = decodeJSON(r, &req)
		added = append(added, req)
		_, _ = w.Write([]byte(`{"id":1,"name":"Demo","technologies":[]}`))
	})

	detected := []detect.Technology{
		{Name: "WordPress", Version: "6.8.3", Category: "Framework"},
		{Name: "Ubuntu", Version: "24.04", Category: "OperatingSystem"},
	}
	cfg := &config.ProjectConfig{ProjectID: 1}

	err := applyTechnologies(client, 1, detected, selected(2, 0), cfg, func() error { return nil })
	if err != nil {
		t.Fatal(err)
	}

	if len(added) != 1 || added[0].Name != "WordPress" {
		t.Fatalf("expected only WordPress added, got %+v", added)
	}
	if len(cfg.Technologies) != 1 || cfg.Technologies[0].Name != "WordPress" {
		t.Fatalf("expected the added technology tracked, got %+v", cfg.Technologies)
	}
}

func TestApplyTechnologies_AlreadyTracked_IsNotAddedTwice(t *testing.T) {
	calls := 0
	client := serve(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = w.Write([]byte(`{"id":1,"name":"Demo","technologies":[]}`))
	})

	detected := []detect.Technology{{Name: "WordPress", Version: "6.8.3", Category: "Framework"}}
	cfg := &config.ProjectConfig{
		ProjectID:    1,
		Technologies: []config.TrackedTechnology{{Name: "WordPress", Version: "6.8.3"}},
	}

	if err := applyTechnologies(client, 1, detected, selected(1, 0), cfg, func() error { return nil }); err != nil {
		t.Fatal(err)
	}

	if calls != 0 {
		t.Fatalf("a re-scan must not re-add a tracked technology, got %d calls", calls)
	}
	if len(cfg.Technologies) != 1 {
		t.Fatalf("expected no duplicate entry, got %+v", cfg.Technologies)
	}
}

func TestApplyTechnologies_ServerRejectsUnknownTechnology_SurfacesTheReason(t *testing.T) {
	// The catalog is closed, so an unseeded name comes back as ProblemDetails.
	client := serve(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"title":"Unknown technology","detail":"StackDrift does not track 'Wordpres'."}`))
	})

	detected := []detect.Technology{{Name: "Wordpres", Version: "6.8", Category: "Framework"}}
	cfg := &config.ProjectConfig{ProjectID: 1}

	err := applyTechnologies(client, 1, detected, selected(1, 0), cfg, func() error { return nil })
	if err == nil || !strings.Contains(err.Error(), "does not track") {
		t.Fatalf("expected the server's reason to reach the user, got %v", err)
	}
	if len(cfg.Technologies) != 0 {
		t.Fatalf("a rejected technology must not be recorded as tracked, got %+v", cfg.Technologies)
	}
}

func TestApplyManifests_UploadsThePrimaryWithItsLock(t *testing.T) {
	var got api.UploadManifestsRequest
	client := serve(t, func(w http.ResponseWriter, r *http.Request) {
		_ = decodeJSON(r, &got)
		_, _ = w.Write([]byte(`{"summary":{"totalCount":1,"groups":[]},"unsupportedFiles":[]}`))
	})

	all := []detect.Manifest{
		{Ecosystem: "Npm", FileName: "package.json", Path: "/app/package.json", Content: "{}", Primary: true},
		{Ecosystem: "Npm", FileName: "package-lock.json", Path: "/app/package-lock.json", Content: "{}", Primary: false},
	}
	cfg := &config.ProjectConfig{ProjectID: 1}

	err := applyManifests(client, 1, "/app", all[:1], all, selected(1, 0), cfg, func() error { return nil })
	if err != nil {
		t.Fatal(err)
	}

	if len(got.Files) != 2 {
		t.Fatalf("expected the lock uploaded alongside the manifest, got %+v", got.Files)
	}
	if len(cfg.DependencyGrp) != 1 {
		t.Fatalf("expected the group tracked, got %+v", cfg.DependencyGrp)
	}
}

func TestApplyManifests_AlreadyTrackedGroup_IsNotUploadedAgain(t *testing.T) {
	calls := 0
	client := serve(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		_, _ = w.Write([]byte(`{"summary":{"totalCount":0,"groups":[]},"unsupportedFiles":[]}`))
	})

	all := []detect.Manifest{
		{Ecosystem: "Npm", FileName: "package.json", Path: "/app/package.json", Content: "{}", Primary: true},
	}
	cfg := &config.ProjectConfig{
		ProjectID:     1,
		DependencyGrp: []config.TrackedDependencyGroup{{Name: groupNameFor("/app", all[0])}},
	}

	if err := applyManifests(client, 1, "/app", all, all, selected(1, 0), cfg, func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatalf("a re-scan must not re-upload a tracked group, got %d calls", calls)
	}
}

func TestResolveProject_LinkedAndStillExists_ReusesItWithoutPrompting(t *testing.T) {
	client := serve(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":7,"name":"Demo","technologies":[]}`))
	})

	project, existing, err := resolveProject(client, linkedDir(t, 7), true)
	if err != nil {
		t.Fatal(err)
	}
	if project == nil || project.ID != 7 {
		t.Fatalf("expected the linked project, got %+v", project)
	}
	if existing == nil || existing.ProjectID != 7 {
		t.Fatalf("expected the saved config returned, got %+v", existing)
	}
}

func TestResolveProject_LinkedButDeleted_CannotContinueUnattended(t *testing.T) {
	client := serve(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, _, err := resolveProject(client, linkedDir(t, 7), true)

	if err == nil {
		t.Fatal("expected --yes to refuse to pick a project on its own")
	}
}

func TestResolveProject_UnlinkedWithYes_RefusesRatherThanGuessing(t *testing.T) {
	client := serve(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("expected no request before a project is chosen")
	})

	if _, _, err := resolveProject(client, unlinkedDir(t), true); err == nil {
		t.Fatal("expected an error telling the user to run scan interactively first")
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	full := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestScan_WithYes_AddsDetectedProjectTechnologies(t *testing.T) {
	var added []api.AddTechnologyRequest
	client := serve(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/technologies") {
			var req api.AddTechnologyRequest
			_ = decodeJSON(r, &req)
			added = append(added, req)
		}
		_, _ = w.Write([]byte(`{"id":7,"name":"Demo","technologies":[]}`))
	})

	dir := linkedDir(t, 7)
	if err := os.MkdirAll(filepath.Join(dir, "wp-admin"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, dir, filepath.Join("wp-includes", "version.php"), "<?php\n$wp_version = '6.8.3';\n")

	if err := scan(client, dir, true); err != nil {
		t.Fatal(err)
	}

	if len(added) != 1 || added[0].Name != "WordPress" || added[0].Version != "6.8.3" {
		t.Fatalf("expected WordPress 6.8.3 added, got %+v", added)
	}
}

func TestScan_WithYes_DoesNotAddTheMachineOperatingSystem(t *testing.T) {
	var added []api.AddTechnologyRequest
	client := serve(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/technologies") {
			var req api.AddTechnologyRequest
			_ = decodeJSON(r, &req)
			added = append(added, req)
		}
		_, _ = w.Write([]byte(`{"id":7,"name":"Demo","technologies":[]}`))
	})

	dir := linkedDir(t, 7)
	writeFile(t, dir, "composer.json", `{"require":{"laravel/framework":"^11.9"}}`)

	if err := scan(client, dir, true); err != nil {
		t.Fatal(err)
	}

	// An unattended scan must never record the developer's own machine as part
	// of the project.
	for _, req := range added {
		if req.Category == "OperatingSystem" {
			t.Fatalf("expected no host operating system added, got %+v", req)
		}
	}
	if len(added) != 1 || added[0].Name != "Laravel" {
		t.Fatalf("expected only Laravel added, got %+v", added)
	}
}

func TestScan_NothingDetected_MakesNoChanges(t *testing.T) {
	posted := false
	client := serve(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			posted = true
		}
		_, _ = w.Write([]byte(`{"id":7,"name":"Demo","technologies":[]}`))
	})

	dir := linkedDir(t, 7)
	writeFile(t, dir, "README.md", "nothing to see")

	if err := scan(client, dir, true); err != nil {
		t.Fatal(err)
	}
	if posted {
		t.Fatal("expected no changes when nothing is detected")
	}
}
