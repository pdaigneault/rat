package parser

import (
	"io"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

// Markdown extracts readable prose from a markdown document, discarding syntax
// that carries no spoken content: code blocks and spans, tables, images, bare
// URLs and raw HTML. Link text is kept while the destination URL is dropped
// (which happens naturally — the URL lives on the node, not in a text child).
type Markdown struct{}

// Parse reads the whole document, walks its AST, and returns the prose as a flat
// token stream. The final token of each block (paragraph, heading, list item)
// is promoted to at least a Sentence boundary so the reader pauses between
// blocks even when they lack terminal punctuation, such as a bare heading.
func (Markdown) Parse(r io.Reader) ([]Token, error) {
	source, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// The Table extension makes goldmark emit table nodes we can skip wholesale;
	// without it, table rows would leak through as pipe-laden paragraph text.
	md := goldmark.New(goldmark.WithExtensions(extension.Table))
	doc := md.Parser().Parse(text.NewReader(source))

	var tokens []Token
	var block strings.Builder

	flush := func() {
		prose := strings.TrimSpace(block.String())
		block.Reset()
		if prose == "" {
			return
		}
		bt := tokenize(prose)
		if n := len(bt); n > 0 && bt[n-1].Boundary < Sentence {
			bt[n-1].Boundary = Sentence
		}
		tokens = append(tokens, bt...)
	}

	walkErr := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch node := n.(type) {
		case *ast.FencedCodeBlock, *ast.CodeBlock, *ast.CodeSpan,
			*ast.Image, *ast.HTMLBlock, *ast.RawHTML, *ast.AutoLink,
			*east.Table:
			// Skip the node and its entire subtree: no spoken content here.
			return ast.WalkSkipChildren, nil
		case *ast.Text:
			if entering {
				block.Write(node.Value(source))
				if node.SoftLineBreak() || node.HardLineBreak() {
					block.WriteByte(' ')
				}
			}
		case *ast.String:
			if entering {
				block.Write(node.Value)
			}
		case *ast.Paragraph, *ast.Heading, *ast.TextBlock:
			// Block-level prose containers: flush their accumulated text on exit.
			if !entering {
				flush()
			}
		}
		return ast.WalkContinue, nil
	})
	if walkErr != nil {
		return nil, walkErr
	}
	// Defensive: flush any prose from a block type not handled above.
	flush()
	return tokens, nil
}
