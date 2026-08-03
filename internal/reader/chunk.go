package reader

import (
	"strings"

	"github.com/pdaigneault/rat/internal/parser"
)

// Chunk is a group of one to three words flashed together. Boundary is the
// strongest trailing boundary among its words, so the engine can pace the pause
// that follows the chunk. Text is the space-joined words, used for ORP and
// display.
type Chunk struct {
	Words    []string
	Boundary parser.Boundary
}

// Text returns the chunk's words joined by single spaces.
func (c Chunk) Text() string {
	return strings.Join(c.Words, " ")
}

// Group packs tokens into chunks of at most size words (size is clamped to
// 1..3). A chunk never straddles a sentence: on hitting a Sentence boundary the
// chunk closes early, so a full stop always lands at the end of a flash. Each
// chunk's Boundary is the strongest boundary of the tokens it contains.
func Group(tokens []parser.Token, size int) []Chunk {
	if size < 1 {
		size = 1
	}
	if size > 3 {
		size = 3
	}

	var chunks []Chunk
	var cur Chunk
	flush := func() {
		if len(cur.Words) > 0 {
			chunks = append(chunks, cur)
			cur = Chunk{}
		}
	}

	for _, tok := range tokens {
		cur.Words = append(cur.Words, tok.Text)
		if tok.Boundary > cur.Boundary {
			cur.Boundary = tok.Boundary
		}
		// Close the chunk when it is full, or early at a sentence end so a chunk
		// never spans two sentences.
		if len(cur.Words) >= size || tok.Boundary == parser.Sentence {
			flush()
		}
	}
	flush()
	return chunks
}
