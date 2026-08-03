package reader

import (
	"time"
	"unicode/utf8"

	"github.com/pdaigneault/rat/internal/parser"
)

// Pacing constants for adaptive mode. Named so they are easy to tune. Each is
// expressed as a multiple of the base per-word time.
const (
	// longWordThreshold: words longer than this (in runes) take extra time.
	longWordThreshold = 8
	// longWordFactor multiplies the whole chunk time when it contains a long word.
	longWordFactor = 1.4
	// clausePause is added on a clause boundary (comma, semicolon, ...).
	clausePause = 0.6
	// sentencePause is added on a sentence boundary (. ! ?).
	sentencePause = 1.2
)

// Delay computes how long a chunk should remain on screen at the given words-
// per-minute rate. base is the time budget for a single word.
//
// Non-adaptive: simply base × number of words.
//
// Adaptive: base × words, then ×longWordFactor if any word is long, then a
// fixed pause added for the chunk's trailing boundary. This makes the reader
// linger on hard words and rest at natural punctuation, which markedly improves
// comprehension over a metronomic pace.
func Delay(c Chunk, wpm int, adaptive bool) time.Duration {
	if wpm < 1 {
		wpm = 1
	}
	base := time.Minute / time.Duration(wpm)

	d := float64(base) * float64(len(c.Words))
	if adaptive {
		if hasLongWord(c.Words) {
			d *= longWordFactor
		}
		switch c.Boundary {
		case parser.Sentence:
			d += float64(base) * sentencePause
		case parser.Clause:
			d += float64(base) * clausePause
		}
	}
	return time.Duration(d)
}

// hasLongWord reports whether any word exceeds longWordThreshold runes.
func hasLongWord(words []string) bool {
	for _, w := range words {
		if utf8.RuneCountInString(w) > longWordThreshold {
			return true
		}
	}
	return false
}
