package commands

import "testing"

func TestCveFoundError_Message(t *testing.T) {
	err := &CveFoundError{Technology: 2, Dependency: 3}
	if err.Error() != "2 technology CVEs and 3 dependency CVEs found" {
		t.Fatalf("unexpected message: %q", err.Error())
	}
}

func TestHasFlag(t *testing.T) {
	if !hasFlag([]string{"--yes"}, "--yes", "-y") {
		t.Fatal("expected --yes to match")
	}
	if !hasFlag([]string{"-y"}, "--yes", "-y") {
		t.Fatal("expected -y to match")
	}
	if hasFlag([]string{"scan"}, "--yes", "-y") {
		t.Fatal("did not expect a match")
	}
}
