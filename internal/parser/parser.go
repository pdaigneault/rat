package parser

import (
	"io"
	"strings"
)

// Parser extracts a flat stream of prose tokens from a document. Implementations
// exist per format; new formats plug in here without touching the reader or TUI.
type Parser interface {
	Parse(r io.Reader) ([]Token, error)
}

// tokenize splits already-extracted prose into Tokens, tagging each with the
// boundary implied by its trailing punctuation. It is shared by every Parser
// implementation so boundary rules stay consistent across formats.
func tokenize(prose string) []Token {
	fields := strings.Fields(prose)
	tokens := make([]Token, 0, len(fields))
	for _, f := range fields {
		tokens = append(tokens, Token{Text: f, Boundary: classify(f)})
	}
	return tokens
}
