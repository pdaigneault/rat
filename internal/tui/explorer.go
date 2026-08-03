package tui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pdaigneault/rat/internal/parser"
)

// fileEntry is one row in the explorer: a directory or a supported file.
type fileEntry struct {
	name  string
	path  string
	isDir bool
}

// explorer is a minimal file browser that lists only directories and files rat
// can read. It supports navigating into subdirectories and up to parents
// (including above the starting directory, via the ".." row or the left key).
type explorer struct {
	dir     string
	entries []fileEntry
	cursor  int
	failed  string // set when the current directory can't be read
}

// newExplorer opens the explorer at dir.
func newExplorer(dir string) explorer {
	e := explorer{dir: dir}
	e.load()
	return e
}

// load reads the current directory, keeping only directories and supported
// files, sorted dirs-first then alphabetically. A ".." row is prepended unless
// we're already at the filesystem root, so going up is always discoverable.
func (e *explorer) load() {
	e.entries = nil
	e.cursor = 0
	e.failed = ""

	if parent := filepath.Dir(e.dir); parent != e.dir {
		e.entries = append(e.entries, fileEntry{name: "..", path: parent, isDir: true})
	}

	dirEntries, err := os.ReadDir(e.dir)
	if err != nil {
		e.failed = "can't open " + compactPath(e.dir)
		return
	}

	var dirs, files []fileEntry
	for _, de := range dirEntries {
		name := de.Name()
		if strings.HasPrefix(name, ".") {
			continue // skip hidden entries
		}
		full := filepath.Join(e.dir, name)
		switch {
		case de.IsDir():
			dirs = append(dirs, fileEntry{name: name, path: full, isDir: true})
		case isSupported(name):
			files = append(files, fileEntry{name: name, path: full})
		}
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].name < dirs[j].name })
	sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })
	e.entries = append(e.entries, dirs...)
	e.entries = append(e.entries, files...)
}

// isSupported reports whether a filename has a readable extension.
func isSupported(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	for _, s := range parser.SupportedExtensions {
		if ext == s {
			return true
		}
	}
	return false
}

func (e *explorer) moveUp() {
	if e.cursor > 0 {
		e.cursor--
	}
}

func (e *explorer) moveDown() {
	if e.cursor < len(e.entries)-1 {
		e.cursor++
	}
}

// enter acts on the highlighted row: descend into a directory (returns
// selected=false) or choose a file (returns its path with selected=true).
func (e *explorer) enter() (path string, selected bool) {
	if e.cursor < 0 || e.cursor >= len(e.entries) {
		return "", false
	}
	cur := e.entries[e.cursor]
	if cur.isDir {
		e.dir = cur.path
		e.load()
		return "", false
	}
	return cur.path, true
}

// ascend navigates to the parent directory, stopping at the filesystem root.
func (e *explorer) ascend() {
	if parent := filepath.Dir(e.dir); parent != e.dir {
		e.dir = parent
		e.load()
	}
}

// startDir is where the explorer opens: the current working directory, or "."
// if that can't be determined.
func startDir() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "."
}

// compactPath abbreviates the user's home directory to "~" for display.
func compactPath(p string) string {
	if home, err := os.UserHomeDir(); err == nil && home != "" && strings.HasPrefix(p, home) {
		return "~" + p[len(home):]
	}
	return p
}
