package ui

import "testing"

func TestNextAction(t *testing.T) {
	cases := []struct {
		name         string
		in           []byte
		wantAction   keyAction
		wantConsumed int
		wantMore     bool
	}{
		{"empty needs more", nil, keyIgnore, 0, true},
		{"up arrow", []byte{0x1b, '[', 'A'}, keyUp, 3, false},
		{"down arrow", []byte{0x1b, '[', 'B'}, keyDown, 3, false},
		{"unknown csi consumed", []byte{0x1b, '[', 'C'}, keyIgnore, 3, false},
		{"partial esc needs more", []byte{0x1b}, keyIgnore, 0, true},
		{"esc bracket needs more", []byte{0x1b, '['}, keyIgnore, 0, true},
		{"esc then non-bracket", []byte{0x1b, 'O'}, keyIgnore, 1, false},
		{"arrow then buffered key", []byte{0x1b, '[', 'B', ' '}, keyDown, 3, false},
		{"k", []byte{'k'}, keyUp, 1, false},
		{"j", []byte{'j'}, keyDown, 1, false},
		{"space", []byte{' '}, keyToggle, 1, false},
		{"enter cr", []byte{'\r'}, keyConfirm, 1, false},
		{"enter lf", []byte{'\n'}, keyConfirm, 1, false},
		{"ctrl-c", []byte{0x03}, keyCancel, 1, false},
		{"q", []byte{'q'}, keyCancel, 1, false},
		{"a", []byte{'a'}, keySelectAll, 1, false},
		{"n", []byte{'n'}, keySelectNone, 1, false},
		{"i", []byte{'i'}, keyInvert, 1, false},
		{"upper A", []byte{'A'}, keySelectAll, 1, false},
		{"unknown", []byte{'z'}, keyIgnore, 1, false},
	}
	for _, c := range cases {
		a, consumed, more := nextAction(c.in)
		if a != c.wantAction || consumed != c.wantConsumed || more != c.wantMore {
			t.Errorf("%s: nextAction(%v) = (%d,%d,%v), want (%d,%d,%v)",
				c.name, c.in, a, consumed, more, c.wantAction, c.wantConsumed, c.wantMore)
		}
	}
}

func items(sel ...bool) []Item {
	out := make([]Item, len(sel))
	for i, s := range sel {
		out[i] = Item{Label: "x", Selected: s}
	}
	return out
}

func selection(items []Item) []bool {
	out := make([]bool, len(items))
	for i, it := range items {
		out[i] = it.Selected
	}
	return out
}

func TestApplyKey_CursorWraps(t *testing.T) {
	it := items(false, false, false)
	c, _, _ := applyKey(it, 0, keyUp)
	if c != 2 {
		t.Fatalf("up from 0 = %d, want 2 (wrap)", c)
	}
	c, _, _ = applyKey(it, 2, keyDown)
	if c != 0 {
		t.Fatalf("down from last = %d, want 0 (wrap)", c)
	}
	c, _, _ = applyKey(it, 1, keyDown)
	if c != 2 {
		t.Fatalf("down from 1 = %d, want 2", c)
	}
}

func TestApplyKey_Toggle(t *testing.T) {
	it := items(false, false)
	applyKey(it, 1, keyToggle)
	if !it[1].Selected || it[0].Selected {
		t.Fatalf("toggle at 1 gave %v", selection(it))
	}
}

func TestApplyKey_SelectAllNoneInvert(t *testing.T) {
	it := items(true, false, true)

	applyKey(it, 0, keySelectAll)
	if got := selection(it); !(got[0] && got[1] && got[2]) {
		t.Fatalf("select all gave %v", got)
	}

	applyKey(it, 0, keySelectNone)
	if got := selection(it); got[0] || got[1] || got[2] {
		t.Fatalf("select none gave %v", got)
	}

	it = items(true, false, true)
	applyKey(it, 0, keyInvert)
	if got := selection(it); !(!got[0] && got[1] && !got[2]) {
		t.Fatalf("invert gave %v, want [false true false]", got)
	}
}

func TestApplyKey_ConfirmAndCancel(t *testing.T) {
	it := items(true)
	if _, done, cancelled := applyKey(it, 0, keyConfirm); !done || cancelled {
		t.Fatalf("confirm: done=%v cancelled=%v", done, cancelled)
	}
	if _, done, cancelled := applyKey(it, 0, keyCancel); !done || !cancelled {
		t.Fatalf("cancel: done=%v cancelled=%v", done, cancelled)
	}
	if _, done, _ := applyKey(it, 0, keyIgnore); done {
		t.Fatalf("ignore should not finish")
	}
}
