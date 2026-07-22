package commands

import (
	"bytes"
	"strings"
	"testing"
)

func completionFixture() []CommandInfo {
	return []CommandInfo{
		{Name: "login", Help: "sign in"},
		{Name: "scan", Help: "detect technologies", Options: []OptionInfo{
			{Name: "--yes", Help: "accept everything"},
		}},
		{Name: "status", Help: "show what is tracked"},
	}
}

func complete(t *testing.T, line string) []string {
	t.Helper()

	var out bytes.Buffer
	if err := CompleteLine(&out, completionFixture(), line); err != nil {
		t.Fatal(err)
	}

	var names []string
	for _, l := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		if l == "" {
			continue
		}
		names = append(names, strings.SplitN(l, "\t", 2)[0])
	}
	return names
}

func TestCompleteLine_AfterTheProgramName_OffersEveryCommand(t *testing.T) {
	got := complete(t, "stackdrift ")

	if len(got) != 3 {
		t.Fatalf("expected every command offered, got %v", got)
	}
}

func TestCompleteLine_HalfTypedCommand_OffersOnlyTheMatches(t *testing.T) {
	got := complete(t, "stackdrift s")

	if len(got) != 2 || got[0] != "scan" || got[1] != "status" {
		t.Fatalf("expected scan and status, got %v", got)
	}
}

func TestCompleteLine_HalfTypedCommand_ExcludesNonMatches(t *testing.T) {
	got := complete(t, "stackdrift lo")

	if len(got) != 1 || got[0] != "login" {
		t.Fatalf("expected only login, got %v", got)
	}
}

func TestCompleteLine_AfterACommand_OffersItsOptions(t *testing.T) {
	// The trailing space is what separates this from completing "scan" itself.
	got := complete(t, "stackdrift scan ")

	if len(got) != 1 || got[0] != "--yes" {
		t.Fatalf("expected the scan options, got %v", got)
	}
}

func TestCompleteLine_HalfTypedOption_OffersTheMatch(t *testing.T) {
	got := complete(t, "stackdrift scan --y")

	if len(got) != 1 || got[0] != "--yes" {
		t.Fatalf("expected --yes, got %v", got)
	}
}

func TestCompleteLine_OptionAlreadyTyped_IsNotOfferedTwice(t *testing.T) {
	if got := complete(t, "stackdrift scan --yes "); len(got) != 0 {
		t.Fatalf("expected nothing left to offer, got %v", got)
	}
}

func TestCompleteLine_CommandWithoutOptions_OffersNothing(t *testing.T) {
	if got := complete(t, "stackdrift status "); len(got) != 0 {
		t.Fatalf("expected no options for status, got %v", got)
	}
}

func TestCompleteLine_UnknownCommand_OffersNothing(t *testing.T) {
	if got := complete(t, "stackdrift nonsense "); len(got) != 0 {
		t.Fatalf("expected nothing for an unknown command, got %v", got)
	}
}

func TestCompleteLine_EmptyLine_OffersEveryCommand(t *testing.T) {
	if got := complete(t, ""); len(got) != 3 {
		t.Fatalf("expected every command offered, got %v", got)
	}
}

func TestCompleteLine_DescribesEachCandidate(t *testing.T) {
	var out bytes.Buffer
	if err := CompleteLine(&out, completionFixture(), "stackdrift sc"); err != nil {
		t.Fatal(err)
	}

	// zsh and fish show the text after the tab beside each candidate.
	if out.String() != "scan\tdetect technologies\n" {
		t.Fatalf("expected the description after a tab, got %q", out.String())
	}
}

func TestCompletion_EverySupportedShell_EmitsAScriptThatAsksTheBinary(t *testing.T) {
	for _, shell := range CompletionShells() {
		var out bytes.Buffer
		if err := Completion(&out, []string{shell}); err != nil {
			t.Fatalf("%s: %v", shell, err)
		}
		// A script that listed the commands itself would go stale on the next
		// release, so each one has to call back into the binary.
		if !strings.Contains(out.String(), "stackdrift __complete") {
			t.Fatalf("%s script does not ask the binary for candidates: %q", shell, out.String())
		}
	}
}

func TestCompletion_ShellNameIsCaseInsensitive(t *testing.T) {
	var out bytes.Buffer
	if err := Completion(&out, []string{"Bash"}); err != nil {
		t.Fatal(err)
	}
	if out.Len() == 0 {
		t.Fatal("expected a script")
	}
}

func TestCompletion_UnknownShell_NamesTheSupportedOnes(t *testing.T) {
	err := Completion(&bytes.Buffer{}, []string{"tcsh"})

	if err == nil || !strings.Contains(err.Error(), "bash") {
		t.Fatalf("expected the supported shells listed, got %v", err)
	}
}

func TestCompletion_NoShell_ReportsTheUsage(t *testing.T) {
	if err := Completion(&bytes.Buffer{}, nil); err == nil {
		t.Fatal("expected an error when no shell is named")
	}
}
