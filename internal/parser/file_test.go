package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParserFor(t *testing.T) {
	if _, ok := ParserFor("notes.md").(Markdown); !ok {
		t.Error(".md should use Markdown")
	}
	if _, ok := ParserFor("notes.markdown").(Markdown); !ok {
		t.Error(".markdown should use Markdown")
	}
	if _, ok := ParserFor("notes.txt").(PlainText); !ok {
		t.Error(".txt should use PlainText")
	}
	if _, ok := ParserFor("NOTES.TEXT").(PlainText); !ok {
		t.Error(".TEXT (any case) should use PlainText")
	}
	if _, ok := ParserFor("noextension").(Markdown); !ok {
		t.Error("unknown extension should default to Markdown")
	}
}

func TestParseFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.txt")
	if err := os.WriteFile(p, []byte("one two, three."), 0o644); err != nil {
		t.Fatal(err)
	}
	toks, err := ParseFile(p)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if len(toks) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(toks))
	}
	if toks[2].Boundary != Sentence {
		t.Errorf("last token should end a sentence, got %v", toks[2].Boundary)
	}
}

func TestParseFileMissing(t *testing.T) {
	if _, err := ParseFile(filepath.Join(t.TempDir(), "nope.md")); err == nil {
		t.Error("expected an error for a missing file")
	}
}
