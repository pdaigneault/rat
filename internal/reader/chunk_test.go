package reader

import (
	"testing"

	"github.com/paul-daigneault/rat/internal/parser"
)

func toks(specs ...struct {
	t string
	b parser.Boundary
}) []parser.Token {
	out := make([]parser.Token, len(specs))
	for i, s := range specs {
		out[i] = parser.Token{Text: s.t, Boundary: s.b}
	}
	return out
}

type ts = struct {
	t string
	b parser.Boundary
}

func TestGroupSizeOne(t *testing.T) {
	in := toks(ts{"a", parser.None}, ts{"b", parser.None}, ts{"c", parser.None})
	got := Group(in, 1)
	if len(got) != 3 {
		t.Fatalf("size 1: expected 3 chunks, got %d", len(got))
	}
	for i, c := range got {
		if len(c.Words) != 1 {
			t.Errorf("chunk %d has %d words, want 1", i, len(c.Words))
		}
	}
}

func TestGroupSizeThree(t *testing.T) {
	in := toks(ts{"a", parser.None}, ts{"b", parser.None}, ts{"c", parser.None},
		ts{"d", parser.None}, ts{"e", parser.None})
	got := Group(in, 3)
	// abc | de
	if len(got) != 2 || len(got[0].Words) != 3 || len(got[1].Words) != 2 {
		t.Fatalf("size 3: got %#v", got)
	}
}

func TestGroupBreaksOnSentence(t *testing.T) {
	in := toks(ts{"a", parser.None}, ts{"b.", parser.Sentence}, ts{"c", parser.None})
	got := Group(in, 3)
	// Even though size is 3, the sentence end after "b." closes the chunk early.
	if len(got) != 2 {
		t.Fatalf("expected early break into 2 chunks, got %d: %#v", len(got), got)
	}
	if len(got[0].Words) != 2 || got[0].Boundary != parser.Sentence {
		t.Errorf("first chunk should be [a b.] with Sentence boundary, got %#v", got[0])
	}
}

func TestGroupBoundaryPropagation(t *testing.T) {
	in := toks(ts{"a", parser.None}, ts{"b,", parser.Clause}, ts{"c", parser.None})
	got := Group(in, 3)
	// One chunk of all three; its boundary is the strongest = Clause.
	if len(got) != 1 || got[0].Boundary != parser.Clause {
		t.Fatalf("expected single Clause chunk, got %#v", got)
	}
}

func TestGroupClampsSize(t *testing.T) {
	in := toks(ts{"a", parser.None}, ts{"b", parser.None}, ts{"c", parser.None},
		ts{"d", parser.None})
	// size 9 clamps to 3.
	got := Group(in, 9)
	if len(got[0].Words) != 3 {
		t.Errorf("size 9 should clamp to 3, first chunk had %d words", len(got[0].Words))
	}
}
