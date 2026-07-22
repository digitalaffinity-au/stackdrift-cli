package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func stubRegistry(calls *[]string, err error) map[string]command {
	record := func(name string) func([]string) error {
		return func(args []string) error {
			*calls = append(*calls, name+" "+strings.Join(args, " "))
			return err
		}
	}
	return map[string]command{
		"scan":  {record("scan"), "detect technologies"},
		"check": {record("check"), "report CVE status"},
	}
}

func TestRun_KnownCommand_SucceedsAndForwardsArguments(t *testing.T) {
	var calls []string
	var stdout, stderr bytes.Buffer

	code := run([]string{"scan", "--yes"}, stubRegistry(&calls, nil), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if len(calls) != 1 || calls[0] != "scan --yes" {
		t.Fatalf("expected the flags forwarded to the command, got %v", calls)
	}
}

func TestRun_CommandFails_ExitsNonZeroAndReportsOnStderr(t *testing.T) {
	var calls []string
	var stdout, stderr bytes.Buffer

	code := run([]string{"check"}, stubRegistry(&calls, errors.New("2 technology CVEs found")), &stdout, &stderr)

	// This is the CI gate: a failing check has to exit non-zero.
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "2 technology CVEs found") {
		t.Fatalf("expected the reason on stderr, got %q", stderr.String())
	}
}

func TestRun_NoArguments_ShowsUsageAndFails(t *testing.T) {
	var calls []string
	var stdout, stderr bytes.Buffer

	code := run(nil, stubRegistry(&calls, nil), &stdout, &stderr)

	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Usage: stackdrift") {
		t.Fatalf("expected usage on stdout, got %q", stdout.String())
	}
	if len(calls) != 0 {
		t.Fatalf("expected no command to run, got %v", calls)
	}
}

func TestRun_UnknownCommand_NamesItAndFails(t *testing.T) {
	var calls []string
	var stdout, stderr bytes.Buffer

	code := run([]string{"scna"}, stubRegistry(&calls, nil), &stdout, &stderr)

	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unknown command: scna") {
		t.Fatalf("expected the typo echoed back, got %q", stderr.String())
	}
	if len(calls) != 0 {
		t.Fatalf("expected no command to run, got %v", calls)
	}
}

func TestRun_HelpFlags_ShowUsageAndSucceed(t *testing.T) {
	for _, arg := range []string{"help", "-h", "--help"} {
		var calls []string
		var stdout, stderr bytes.Buffer

		code := run([]string{arg}, stubRegistry(&calls, nil), &stdout, &stderr)

		if code != 0 {
			t.Fatalf("expected %q to exit 0, got %d", arg, code)
		}
		if !strings.Contains(stdout.String(), "Usage: stackdrift") {
			t.Fatalf("expected usage for %q, got %q", arg, stdout.String())
		}
	}
}

func TestRegistry_ListsEveryCommandTheUsageAdvertises(t *testing.T) {
	// usage() prints from a hardcoded order list, so a command added to one and
	// not the other would show a blank description.
	advertised := []string{"login", "scan", "status", "check", "remove", "whoami", "logout", "update", "version"}
	actual := registry()

	for _, name := range advertised {
		if _, ok := actual[name]; !ok {
			t.Fatalf("usage advertises %q but the registry has no such command", name)
		}
	}
	if len(actual) != len(advertised) {
		t.Fatalf("expected the registry and usage list to match, registry has %d and usage lists %d", len(actual), len(advertised))
	}
}

func TestUsage_DescribesEveryCommand(t *testing.T) {
	var out bytes.Buffer
	usage(&out, registry())

	for name, cmd := range registry() {
		if !strings.Contains(out.String(), name) {
			t.Fatalf("expected %q in the usage text", name)
		}
		if !strings.Contains(out.String(), cmd.help) {
			t.Fatalf("expected the description for %q in the usage text", name)
		}
	}
}
