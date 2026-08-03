package tui

import (
	"os"
	"path/filepath"
	"testing"
)

// buildTree makes a temp dir with a mix of supported files, an unsupported file,
// a hidden file, and a subdirectory. It returns the root.
func buildTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writes := map[string]string{
		"notes.md":    "# hi",
		"plain.txt":   "hello",
		"image.png":   "binary",
		".secret.md":  "hidden",
		"sub/deep.md": "# deep",
	}
	for rel, content := range writes {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func names(e explorer) []string {
	out := make([]string, len(e.entries))
	for i, en := range e.entries {
		out[i] = en.name
	}
	return out
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

func TestExplorerFiltersEntries(t *testing.T) {
	root := buildTree(t)
	e := newExplorer(root)
	got := names(e)

	// Supported files and the subdirectory are shown; unsupported and hidden are not.
	for _, want := range []string{"..", "sub", "notes.md", "plain.txt"} {
		if !contains(got, want) {
			t.Errorf("expected entry %q, got %v", want, got)
		}
	}
	for _, banned := range []string{"image.png", ".secret.md"} {
		if contains(got, banned) {
			t.Errorf("entry %q should be filtered out, got %v", banned, got)
		}
	}
}

func TestExplorerDescendAndSelect(t *testing.T) {
	root := buildTree(t)
	e := newExplorer(root)

	// Move to "sub" and descend into it.
	for {
		cur := e.entries[e.cursor]
		if cur.name == "sub" {
			break
		}
		e.moveDown()
	}
	if _, sel := e.enter(); sel {
		t.Fatal("entering a directory should not select a file")
	}
	if filepath.Base(e.dir) != "sub" {
		t.Fatalf("expected to be in sub, got %s", e.dir)
	}
	// deep.md is the only supported file in sub.
	if !contains(names(e), "deep.md") {
		t.Fatalf("expected deep.md in sub, got %v", names(e))
	}
	// Select deep.md.
	for e.entries[e.cursor].name != "deep.md" {
		e.moveDown()
	}
	path, sel := e.enter()
	if !sel || filepath.Base(path) != "deep.md" {
		t.Fatalf("expected to select deep.md, got sel=%v path=%s", sel, path)
	}
}

func TestExplorerAscend(t *testing.T) {
	root := buildTree(t)
	sub := filepath.Join(root, "sub")
	e := newExplorer(sub)
	e.ascend()
	if e.dir != root {
		t.Errorf("ascend should move to parent %s, got %s", root, e.dir)
	}
	// The ".." row lets us go up even further (above the start dir).
	if e.entries[0].name != ".." {
		t.Errorf("expected a '..' entry at the top, got %v", names(e))
	}
}
