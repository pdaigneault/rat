// Package reader turns the parser's token stream into a timed, positioned
// display: it groups tokens into chunks, finds each chunk's Optimal Recognition
// Point (ORP), and computes how long each chunk should stay on screen.
package reader

// MaxPivot is the largest pivot index the stepped table below can return. The
// view uses it to reserve enough left padding that the pivot column never moves
// regardless of word length.
const MaxPivot = 4

// PivotIndex returns the Optimal Recognition Point for a word: the rune index
// the eye should land on, slightly left of centre. The mapping is the classic
// Spritz stepped table, keyed on rune length (not byte length, so multi-byte
// characters count as one):
//
//	1        -> 0
//	2..5     -> 1
//	6..9     -> 2
//	10..13   -> 3
//	>=14     -> 4
func PivotIndex(word string) int {
	n := len([]rune(word))
	switch {
	case n <= 1:
		return 0
	case n <= 5:
		return 1
	case n <= 9:
		return 2
	case n <= 13:
		return 3
	default:
		return MaxPivot
	}
}

// Split divides a word at its pivot rune, returning the runes before the pivot,
// the pivot rune itself, and the runes after. The view styles the pivot rune in
// the highlight colour and left-pads left so the pivot lands in a fixed column.
// An empty word yields three empty strings.
func Split(word string) (left, pivot, right string) {
	runes := []rune(word)
	if len(runes) == 0 {
		return "", "", ""
	}
	p := PivotIndex(word)
	return string(runes[:p]), string(runes[p : p+1]), string(runes[p+1:])
}
