package ui

import (
	"bufio"
	"strings"
	"testing"
)

func answer(t *testing.T, input string) {
	t.Helper()
	original := reader
	reader = bufio.NewReader(strings.NewReader(input))
	t.Cleanup(func() { reader = original })
}

func TestConfirm_EmptyAnswer_TakesTheDefault(t *testing.T) {
	answer(t, "\n")
	if !Confirm("Proceed?", true) {
		t.Fatal("expected the true default")
	}

	answer(t, "\n")
	if Confirm("Proceed?", false) {
		t.Fatal("expected the false default")
	}
}

func TestConfirm_Yes_IsAccepted(t *testing.T) {
	for _, input := range []string{"y\n", "yes\n", "Y\n", "YES\n", "  y  \n"} {
		answer(t, input)
		if !Confirm("Proceed?", false) {
			t.Fatalf("expected %q to confirm", input)
		}
	}
}

func TestConfirm_No_IsRejected(t *testing.T) {
	for _, input := range []string{"n\n", "no\n", "N\n"} {
		answer(t, input)
		if Confirm("Proceed?", true) {
			t.Fatalf("expected %q to decline even against a true default", input)
		}
	}
}

func TestConfirm_UnrecognisedAnswer_IsNotTakenAsYes(t *testing.T) {
	answer(t, "maybe\n")
	if Confirm("Proceed?", true) {
		t.Fatal("anything that is not yes must not confirm a destructive prompt")
	}
}

func TestAskInt_ValueInRange_IsAccepted(t *testing.T) {
	answer(t, "3\n")
	value, ok := AskInt("Pick: ", 1, 5)
	if !ok || value != 3 {
		t.Fatalf("expected 3, got %d (ok %v)", value, ok)
	}
}

func TestAskInt_Boundaries_AreInclusive(t *testing.T) {
	answer(t, "1\n")
	if value, ok := AskInt("Pick: ", 1, 5); !ok || value != 1 {
		t.Fatalf("expected the minimum to be allowed, got %d (ok %v)", value, ok)
	}

	answer(t, "5\n")
	if value, ok := AskInt("Pick: ", 1, 5); !ok || value != 5 {
		t.Fatalf("expected the maximum to be allowed, got %d (ok %v)", value, ok)
	}
}

func TestAskInt_OutOfRange_IsRejected(t *testing.T) {
	answer(t, "0\n")
	if _, ok := AskInt("Pick: ", 1, 5); ok {
		t.Fatal("expected a value below the minimum to be rejected")
	}

	answer(t, "6\n")
	if _, ok := AskInt("Pick: ", 1, 5); ok {
		t.Fatal("expected a value above the maximum to be rejected")
	}
}

func TestAskInt_NotANumber_IsRejected(t *testing.T) {
	answer(t, "three\n")
	if _, ok := AskInt("Pick: ", 1, 5); ok {
		t.Fatal("expected non-numeric input to be rejected")
	}
}

func TestAskInt_EmptyAnswer_IsRejected(t *testing.T) {
	answer(t, "\n")
	if _, ok := AskInt("Pick: ", 1, 5); ok {
		t.Fatal("expected an empty answer to be rejected")
	}
}

func TestAsk_TrimsSurroundingWhitespace(t *testing.T) {
	answer(t, "  My Project  \n")
	if got := Ask("Name: "); got != "My Project" {
		t.Fatalf("expected trimmed input, got %q", got)
	}
}
