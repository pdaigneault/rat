package reader

import "testing"

func TestPivotIndex(t *testing.T) {
	cases := []struct {
		word string
		want int
	}{
		{"a", 0},               // length 1
		{"to", 1},              // length 2
		{"there", 1},           // length 5
		{"reader", 2},          // length 6
		{"wonderful", 2},       // length 9
		{"incredible", 3},      // length 10
		{"extraordinar", 3},    // length 12
		{"conversationa", 3},   // length 13
		{"internationally", 4}, // length 15
		{"café", 1},            // 4 runes (accented, still counts as one rune)
		{"naïveté", 2},         // 7 runes
	}
	for _, c := range cases {
		if got := PivotIndex(c.word); got != c.want {
			t.Errorf("PivotIndex(%q) = %d, want %d", c.word, got, c.want)
		}
	}
}

func TestSplitReassembles(t *testing.T) {
	for _, w := range []string{"a", "to", "reader", "internationally", "café"} {
		l, p, r := Split(w)
		if l+p+r != w {
			t.Errorf("Split(%q) = %q+%q+%q, does not reassemble", w, l, p, r)
		}
		if len([]rune(p)) != 1 {
			t.Errorf("Split(%q) pivot %q is not a single rune", w, p)
		}
	}
}

func TestSplitEmpty(t *testing.T) {
	l, p, r := Split("")
	if l != "" || p != "" || r != "" {
		t.Errorf("Split(\"\") = %q,%q,%q; want all empty", l, p, r)
	}
}
