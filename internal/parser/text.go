package parser

import "io"

// PlainText treats the whole input as prose: no syntax to strip, just split on
// whitespace and classify boundaries by trailing punctuation. It is selected for
// .txt files and is the simplest demonstration of the Parser seam.
type PlainText struct{}

// Parse reads all input and tokenizes it directly.
func (PlainText) Parse(r io.Reader) ([]Token, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return tokenize(string(b)), nil
}
