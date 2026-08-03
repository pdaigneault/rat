package parser

import (
	"os"
	"path/filepath"
	"strings"
)

// SupportedExtensions lists the file types rat can read. It drives the file
// picker's filter so only readable files are selectable.
var SupportedExtensions = []string{".md", ".markdown", ".txt", ".text"}

// ParserFor returns the Parser appropriate for a filename, chosen by extension.
// Plain-text extensions use PlainText; everything else defaults to Markdown
// (which also degrades gracefully to plain prose for unknown formats).
func ParserFor(name string) Parser {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".txt", ".text":
		return PlainText{}
	default:
		return Markdown{}
	}
}

// ParseFile opens path and parses it with the extension-appropriate parser. It
// is the single entry point both the CLI and the in-app file picker use to turn
// a chosen path into tokens.
func ParseFile(path string) ([]Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParserFor(path).Parse(f)
}
