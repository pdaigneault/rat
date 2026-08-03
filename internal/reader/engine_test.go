package reader

import (
	"testing"

	"github.com/paul-daigneault/rat/internal/parser"
)

func TestDelayScalesWithWordCount(t *testing.T) {
	one := Delay(Chunk{Words: []string{"a"}}, 300, false)
	two := Delay(Chunk{Words: []string{"a", "b"}}, 300, false)
	if two != 2*one {
		t.Errorf("non-adaptive: two words (%v) should be twice one word (%v)", two, one)
	}
}

func TestDelayNonAdaptiveIgnoresBoundary(t *testing.T) {
	plain := Chunk{Words: []string{"a"}, Boundary: parser.None}
	ending := Chunk{Words: []string{"a."}, Boundary: parser.Sentence}
	if Delay(plain, 300, false) != Delay(ending, 300, false) {
		t.Errorf("non-adaptive delay must not depend on boundary")
	}
}

func TestAdaptiveSentenceLongerThanNonAdaptive(t *testing.T) {
	c := Chunk{Words: []string{"stop."}, Boundary: parser.Sentence}
	if Delay(c, 300, true) <= Delay(c, 300, false) {
		t.Errorf("adaptive sentence-end delay should exceed non-adaptive")
	}
}

func TestAdaptiveSentenceLongerThanClause(t *testing.T) {
	sentence := Chunk{Words: []string{"word"}, Boundary: parser.Sentence}
	clause := Chunk{Words: []string{"word"}, Boundary: parser.Clause}
	none := Chunk{Words: []string{"word"}, Boundary: parser.None}
	if !(Delay(sentence, 300, true) > Delay(clause, 300, true) &&
		Delay(clause, 300, true) > Delay(none, 300, true)) {
		t.Errorf("adaptive delays should order sentence > clause > none")
	}
}

func TestAdaptiveLongWordSlowsDown(t *testing.T) {
	short := Chunk{Words: []string{"cat"}, Boundary: parser.None}
	long := Chunk{Words: []string{"extraordinary"}, Boundary: parser.None}
	if Delay(long, 300, true) <= Delay(short, 300, true) {
		t.Errorf("adaptive long word should take longer than short word")
	}
}
