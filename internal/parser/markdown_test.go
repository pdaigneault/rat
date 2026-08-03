package parser

import (
	"strings"
	"testing"
)

const fixture = `# Introduction

This is a paragraph with a [link](https://example.com) and some *emphasis*.

Here is a second sentence, with a clause, then done.

- first item
- second item

> A quote to remember.

Some ` + "`inline code`" + ` should vanish.

` + "```go\nfunc main() { fmt.Println(\"skip me\") }\n```" + `

| a | b |
|---|---|
| 1 | 2 |

![alt text](img.png)
`

func texts(tokens []Token) []string {
	out := make([]string, len(tokens))
	for i, t := range tokens {
		out[i] = t.Text
	}
	return out
}

func TestMarkdownStripsSyntax(t *testing.T) {
	tokens, err := Markdown{}.Parse(strings.NewReader(fixture))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	joined := strings.Join(texts(tokens), " ")

	// Code fence body must be gone.
	for _, banned := range []string{"func", "fmt.Println", "skip", "inline", "code"} {
		if strings.Contains(joined, banned) {
			t.Errorf("expected %q to be stripped, got: %q", banned, joined)
		}
	}
	// Table cells must be gone.
	for _, banned := range []string{"| a |", "---"} {
		if strings.Contains(joined, banned) {
			t.Errorf("expected table syntax %q stripped, got: %q", banned, joined)
		}
	}
	// Image alt text and URLs must be gone.
	if strings.Contains(joined, "alt") || strings.Contains(joined, "example.com") || strings.Contains(joined, "img.png") {
		t.Errorf("expected image alt / URLs stripped, got: %q", joined)
	}
	// Link text must be kept.
	if !strings.Contains(joined, "link") {
		t.Errorf("expected link text kept, got: %q", joined)
	}
	// Heading and body prose must survive.
	for _, want := range []string{"Introduction", "paragraph", "emphasis", "first", "second", "quote"} {
		if !strings.Contains(joined, want) {
			t.Errorf("expected %q kept, got: %q", want, joined)
		}
	}
}

func TestMarkdownBoundaries(t *testing.T) {
	tokens, err := Markdown{}.Parse(strings.NewReader("Hello world. Next clause, then stop."))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	// world. -> Sentence, clause, -> Clause, stop. -> Sentence.
	want := map[string]Boundary{
		"world.":  Sentence,
		"clause,": Clause,
		"stop.":   Sentence,
		"Hello":   None,
	}
	got := map[string]Boundary{}
	for _, tok := range tokens {
		got[tok.Text] = tok.Boundary
	}
	for text, b := range want {
		if got[text] != b {
			t.Errorf("token %q: got boundary %v, want %v", text, got[text], b)
		}
	}
}

func TestMarkdownHeadingGetsSentencePause(t *testing.T) {
	// A bare heading has no terminal punctuation, but should still end a block
	// with a Sentence pause so the reader breathes before the body.
	tokens, err := Markdown{}.Parse(strings.NewReader("# Title\n\nBody text here."))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if tokens[0].Text != "Title" {
		t.Fatalf("expected first token 'Title', got %q", tokens[0].Text)
	}
	if tokens[0].Boundary != Sentence {
		t.Errorf("expected heading to end on Sentence boundary, got %v", tokens[0].Boundary)
	}
}

func TestPlainText(t *testing.T) {
	tokens, err := PlainText{}.Parse(strings.NewReader("one two, three."))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(tokens))
	}
	if tokens[1].Boundary != Clause || tokens[2].Boundary != Sentence {
		t.Errorf("boundaries wrong: %+v", tokens)
	}
}
