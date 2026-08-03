// Package parser turns source documents into a flat stream of prose tokens
// that the reader can flash one chunk at a time. Each format (markdown, plain
// text, and later PDF/EPUB) implements the Parser interface, keeping the rest
// of the program independent of how the prose was extracted.
package parser

// Boundary records the kind of pause that should follow a token, derived from
// its trailing punctuation. The timing engine uses it to slow down at natural
// resting points so the reader breathes with the sentence structure.
type Boundary int

const (
	// None means the token flows straight into the next with no extra pause.
	None Boundary = iota
	// Clause marks a mid-sentence break: comma, semicolon, colon, dash.
	Clause
	// Sentence marks a full stop: period, question mark, exclamation mark.
	Sentence
)

// Token is a single whitespace-delimited word plus the boundary implied by its
// trailing punctuation. The punctuation is kept in Text (we flash it with the
// word); Boundary is a classification layered on top for pacing.
type Token struct {
	Text     string
	Boundary Boundary
}

// classify inspects a word's trailing punctuation and returns the strongest
// boundary it implies. Sentence-ending marks win over clause marks.
func classify(word string) Boundary {
	// Walk backwards over trailing punctuation; closing quotes/brackets may sit
	// after the real terminator (e.g. `done."`), so we look past them.
	strongest := None
	for _, r := range reverseRunes(word) {
		switch r {
		case '.', '!', '?':
			return Sentence
		case ',', ';', ':', '—', '–', '-':
			if strongest < Clause {
				strongest = Clause
			}
		case '"', '\'', ')', ']', '}', '»', '”', '’':
			// Trailing closer: keep scanning for the real terminator underneath.
			continue
		default:
			// First non-punctuation rune from the end: stop scanning.
			return strongest
		}
	}
	return strongest
}

// reverseRunes returns the runes of s in reverse order.
func reverseRunes(s string) []rune {
	rs := []rune(s)
	for i, j := 0, len(rs)-1; i < j; i, j = i+1, j-1 {
		rs[i], rs[j] = rs[j], rs[i]
	}
	return rs
}
