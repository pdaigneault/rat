// Command rat plays a markdown or text file back as a speed reader: instead of
// dumping the file like cat, it flashes it one chunk at a time using RSVP with
// an Optimal Recognition Point highlight.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"runtime/debug"

	tea "charm.land/bubbletea/v2"

	"github.com/paul-daigneault/rat/internal/config"
	"github.com/paul-daigneault/rat/internal/parser"
	"github.com/paul-daigneault/rat/internal/tui"
)

// version is set at build time via -ldflags "-X main.version=..." (GoReleaser
// and the Makefile do this). It defaults to "dev"; see versionString for the
// fallback used by `go install`-based builds.
var version = "dev"

// versionString resolves the version to report. A build-time ldflag wins; when
// it's absent (a plain `go install github.com/.../rat@v1.2.3`), we recover the
// module version Go recorded in the binary's build info instead of "dev".
func versionString() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "rat:", err)
		os.Exit(1)
	}
}

func run() error {
	// Flags override persisted config. We detect "was it set" with a sentinel so
	// an unset flag leaves the config value untouched.
	wpm := flag.Int("wpm", 0, "words per minute (100-1000)")
	chunk := flag.Int("chunk", 0, "words per flash (1-3)")
	themeName := flag.String("theme", "", "colour theme: dark, light, solarized, high-contrast")
	adaptive := flag.String("adaptive", "", "adaptive pacing: true or false")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Usage = usage
	flag.Parse()

	if *showVersion {
		fmt.Println("rat", versionString())
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	applyFlags(&cfg, *wpm, *chunk, *themeName, *adaptive)

	src, name, ttyInput, err := resolveInput(flag.Arg(0))
	if err != nil {
		return err
	}
	defer src.Close()

	tokens, err := selectParser(name).Parse(src)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", name, err)
	}
	if len(tokens) == 0 {
		return fmt.Errorf("no readable text found in %s", displayName(name))
	}

	opts := []tea.ProgramOption{}
	if ttyInput != nil {
		// Content came from a pipe, so stdin is the document, not the keyboard.
		// Read keystrokes from the controlling terminal instead.
		opts = append(opts, tea.WithInput(ttyInput))
	}
	p := tea.NewProgram(tui.New(tokens, cfg), opts...)
	_, err = p.Run()
	return err
}

// resolveInput decides where the document comes from and returns a reader for
// it, its name (for parser selection), and — when the document arrived on a pipe
// — an open /dev/tty to read keystrokes from. Callers must Close the reader.
func resolveInput(arg string) (src io.ReadCloser, name string, ttyInput *os.File, err error) {
	if arg != "" {
		f, err := os.Open(arg)
		if err != nil {
			return nil, "", nil, err
		}
		return f, arg, nil, nil
	}

	// No file argument: read from stdin only if it is piped (not a terminal).
	info, statErr := os.Stdin.Stat()
	if statErr != nil {
		return nil, "", nil, statErr
	}
	if info.Mode()&os.ModeCharDevice != 0 {
		// stdin is a terminal and no file was given: nothing to read.
		usage()
		return nil, "", nil, fmt.Errorf("no input: provide a file or pipe text in")
	}

	// Piped input: buffer stdin as the document, then take over the keyboard via
	// the controlling terminal so the TUI stays interactive.
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return nil, "", nil, fmt.Errorf("piped input needs a terminal for controls: %w", err)
	}
	return os.Stdin, "", tty, nil
}

// selectParser picks a Parser by file extension, defaulting to markdown. Piped
// input (empty name) is treated as markdown too, which degrades gracefully to
// plain prose for non-markdown text.
func selectParser(name string) parser.Parser {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".txt", ".text":
		return parser.PlainText{}
	default:
		return parser.Markdown{}
	}
}

// applyFlags overlays any explicitly-set flags onto the config, clamping to the
// valid ranges so bad flag values cannot break the reader.
func applyFlags(cfg *config.Config, wpm, chunk int, themeName, adaptive string) {
	if wpm != 0 {
		cfg.WPM = clamp(wpm, config.MinWPM, config.MaxWPM)
	}
	if chunk != 0 {
		cfg.ChunkSize = clamp(chunk, config.MinChunk, config.MaxChunk)
	}
	if themeName != "" {
		cfg.Theme = themeName
	}
	switch strings.ToLower(adaptive) {
	case "true", "1", "on", "yes":
		cfg.Adaptive = true
	case "false", "0", "off", "no":
		cfg.Adaptive = false
	}
}

func displayName(name string) string {
	if name == "" {
		return "stdin"
	}
	return name
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func usage() {
	fmt.Fprint(os.Stderr, `rat — a speed-reading pager for markdown and text

Usage:
  rat [flags] [file]
  cat file.md | rat [flags]

Flags:
  --wpm N        words per minute (100-1000)
  --chunk N      words per flash (1-3)
  --theme NAME   dark | light | solarized | high-contrast
  --adaptive B   true or false — longer pauses at punctuation and long words

Controls (in-reader, press ? for the full list):
  space play/pause · ←/→ seek · ↑/↓ speed · [ ] chunk · t theme · a adaptive · q quit
`)
}
