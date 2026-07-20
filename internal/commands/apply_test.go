package commands

import (
	"path/filepath"
	"testing"

	"github.com/digitalaffinity-au/stackdrift-cli/internal/detect"
)

func TestGroupNameFor_Csproj_UsesProjectName(t *testing.T) {
	m := detect.Manifest{Ecosystem: "NuGet", FileName: "ProjA.csproj", Path: "/sln/ProjA/ProjA.csproj", Primary: true}
	if got := groupNameFor("/sln", m); got != "ProjA" {
		t.Fatalf("expected ProjA, got %q", got)
	}
}

func TestGroupNameFor_PackageJsonInSubdir_UsesDirAndEcosystem(t *testing.T) {
	m := detect.Manifest{Ecosystem: "Npm", FileName: "package.json", Path: "/repo/web/package.json", Primary: true}
	if got := groupNameFor("/repo", m); got != "web npm" {
		t.Fatalf("expected 'web npm', got %q", got)
	}
}

func TestPrimaryManifests_FiltersOutLockAndProps(t *testing.T) {
	result := &detect.Result{Manifests: []detect.Manifest{
		{FileName: "package.json", Primary: true},
		{FileName: "package-lock.json", Primary: false},
		{FileName: "A.csproj", Primary: true},
		{FileName: "packages.lock.json", Primary: false},
	}}
	primaries := primaryManifests(result)
	if len(primaries) != 2 {
		t.Fatalf("expected 2 primaries, got %d", len(primaries))
	}
}

func TestSupportingFor_PackageJson_BundlesSiblingLock(t *testing.T) {
	all := []detect.Manifest{
		{Ecosystem: "Npm", FileName: "package.json", Path: "/app/package.json", Primary: true},
		{Ecosystem: "Npm", FileName: "package-lock.json", Path: "/app/package-lock.json", Primary: false},
		{Ecosystem: "Npm", FileName: "package-lock.json", Path: "/other/package-lock.json", Primary: false},
	}
	support := supportingFor(all[0], all)
	if len(support) != 1 || support[0].Path != "/app/package-lock.json" {
		t.Fatalf("expected only the sibling lock, got %+v", support)
	}
}

func TestSupportingFor_Csproj_BundlesSiblingLockAndTreeWideProps(t *testing.T) {
	all := []detect.Manifest{
		{Ecosystem: "NuGet", FileName: "A.csproj", Path: "/sln/A/A.csproj", Primary: true},
		{Ecosystem: "NuGet", FileName: "packages.lock.json", Path: "/sln/A/packages.lock.json", Primary: false},
		{Ecosystem: "NuGet", FileName: "Directory.Packages.props", Path: "/sln/Directory.Packages.props", Primary: false},
		{Ecosystem: "NuGet", FileName: "B.csproj", Path: "/sln/B/B.csproj", Primary: true},
	}
	support := supportingFor(all[0], all)
	paths := map[string]bool{}
	for _, m := range support {
		paths[filepath.Base(m.Path)] = true
	}
	if !paths["packages.lock.json"] || !paths["Directory.Packages.props"] {
		t.Fatalf("expected lock + props bundled, got %+v", support)
	}
	if len(support) != 2 {
		t.Fatalf("expected exactly 2 supporting files, got %d", len(support))
	}
}
