package ui

import (
	"strings"
	"testing"
)

func TestViewRows_LeavesRoomForChrome(t *testing.T) {
	// The drawn block must fit the screen, or the cursor-up redraw runs off
	// the top and the display corrupts.
	rows := viewRows(40, 24)
	if rows+chromeLines > 24 {
		t.Fatalf("block of %d rows plus chrome exceeds a 24 line screen", rows)
	}
}

func TestViewRows_ShortListIsNotPadded(t *testing.T) {
	if rows := viewRows(3, 50); rows != 3 {
		t.Fatalf("expected 3 rows for a 3 item list, got %d", rows)
	}
}

func TestViewRows_TinyTerminal_StillShowsSomething(t *testing.T) {
	if rows := viewRows(40, 3); rows < 1 {
		t.Fatalf("expected at least one visible row, got %d", rows)
	}
}

func TestScrollTo_CursorBelowWindow_ScrollsDown(t *testing.T) {
	// 40 items, 10 visible, cursor on the last one.
	if top := scrollTo(0, 39, 10, 40); top != 30 {
		t.Fatalf("expected the window to end on the last item, got top %d", top)
	}
}

func TestScrollTo_CursorAboveWindow_ScrollsUp(t *testing.T) {
	if top := scrollTo(30, 5, 10, 40); top != 5 {
		t.Fatalf("expected the window to start at the cursor, got top %d", top)
	}
}

func TestScrollTo_CursorInsideWindow_DoesNotMove(t *testing.T) {
	if top := scrollTo(10, 15, 10, 40); top != 10 {
		t.Fatalf("expected a stationary window, got top %d", top)
	}
}

func TestScrollTo_EverythingFits_StaysAtTop(t *testing.T) {
	if top := scrollTo(0, 3, 10, 4); top != 0 {
		t.Fatalf("expected no scrolling when the list fits, got top %d", top)
	}
}

func TestScrollTo_NeverScrollsPastTheEnd(t *testing.T) {
	top := scrollTo(38, 39, 10, 40)
	if top+10 > 40 {
		t.Fatalf("window runs past the end, top %d", top)
	}
}

func TestScrollTo_WrapFromTopToBottom_ShowsTheEnd(t *testing.T) {
	// Pressing up on the first item wraps to the last.
	if top := scrollTo(0, 39, 10, 40); top != 30 {
		t.Fatalf("expected the wrap to reveal the last item, got top %d", top)
	}
}

func TestFitLine_LongHintIsTrimmedBeforeTheLabel(t *testing.T) {
	label, hint := fitLine("web npm", strings.Repeat("x", 200), 40)

	if label != "web npm" {
		t.Fatalf("the label identifies the item and should survive, got %q", label)
	}
	if len([]rune(label))+len([]rune(hint))+2 > 40 {
		t.Fatalf("label plus hint still exceeds the budget: %q %q", label, hint)
	}
}

func TestFitLine_LabelLongerThanTheBudget_IsTruncated(t *testing.T) {
	label, hint := fitLine(strings.Repeat("y", 100), "some hint", 20)

	if len([]rune(label)) > 20 {
		t.Fatalf("expected the label trimmed to 20, got %d", len([]rune(label)))
	}
	if hint != "" {
		t.Fatalf("expected no room left for a hint, got %q", hint)
	}
}

func TestFitLine_ShortEnough_IsUntouched(t *testing.T) {
	label, hint := fitLine("app npm", "package.json", 60)
	if label != "app npm" || hint != "package.json" {
		t.Fatalf("expected both kept intact, got %q %q", label, hint)
	}
}

func TestClampWidth_PlainTextIsCutToWidth(t *testing.T) {
	if got := clampWidth(strings.Repeat("a", 100), 10); len([]rune(got)) != 10 {
		t.Fatalf("expected 10 visible runes, got %d (%q)", len([]rune(got)), got)
	}
}

func TestClampWidth_ColourCodesDoNotCountTowardWidth(t *testing.T) {
	// Escapes take no columns, so a short styled string must survive whole.
	styled := "\x1b[36m" + "hello" + "\x1b[0m"

	got := clampWidth(styled, 10)

	if !strings.Contains(got, "hello") {
		t.Fatalf("expected the text kept, got %q", got)
	}
}

func TestClampWidth_CutInsideStyledText_ResetsColour(t *testing.T) {
	styled := "\x1b[36m" + strings.Repeat("z", 50)

	got := clampWidth(styled, 10)

	if !strings.HasSuffix(got, "\x1b[0m") {
		t.Fatalf("expected a colour reset so the rest of the screen is not tinted, got %q", got)
	}
}

func TestClampWidth_ZeroWidth_IsLeftAlone(t *testing.T) {
	// A terminal that reports no size must not blank the output entirely.
	if got := clampWidth("hello", 0); got != "hello" {
		t.Fatalf("expected the line untouched, got %q", got)
	}
}
