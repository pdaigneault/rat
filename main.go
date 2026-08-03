// Command rat plays a markdown or text file back as a speed reader: instead of
// dumping the file like cat, it flashes it one chunk at a time using RSVP with
// an Optimal Recognition Point highlight.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/pdaigneault/rat/internal/config"
	"github.com/pdaigneault/rat/internal/parser"
	"github.com/pdaigneault/rat/internal/tui"
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

	tokens, ttyInput, err := loadInput(flag.Arg(0))
	if err != nil {
		return err
	}

	var opts []tea.ProgramOption
	if ttyInput != nil {
		// The document came from a pipe, so stdin is not the keyboard; read
		// keystrokes from the controlling terminal instead.
		defer ttyInput.Close()
		opts = append(opts, tea.WithInput(ttyInput))
	}
	// tokens may be nil: rat then starts empty and the user picks a file with f.
	p := tea.NewProgram(tui.New(tokens, cfg), opts...)
	_, err = p.Run()
	return err
}

// loadInput resolves the document to read. It returns the parsed tokens — nil
// when rat is launched interactively with no file, so the in-app picker takes
// over — and, when the document arrived on a pipe, an open /dev/tty for
// keystrokes (the caller owns closing it).
func loadInput(arg string) (tokens []parser.Token, ttyInput *os.File, err error) {
	if arg != "" {
		tokens, err = parser.ParseFile(arg)
		if err != nil {
			return nil, nil, err
		}
		if len(tokens) == 0 {
			return nil, nil, fmt.Errorf("no readable text found in %s", arg)
		}
		return tokens, nil, nil
	}

	info, err := os.Stdin.Stat()
	if err != nil {
		return nil, nil, err
	}
	if info.Mode()&os.ModeCharDevice != 0 {
		// Interactive terminal with no file: start empty; the user presses f.
		return nil, nil, nil
	}

	// Piped input: read stdin as the document, then take the keyboard from the
	// controlling terminal so the reader stays interactive.
	tokens, err = parser.Markdown{}.Parse(os.Stdin)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing stdin: %w", err)
	}
	if len(tokens) == 0 {
		return nil, nil, fmt.Errorf("no readable text found on stdin")
	}
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return nil, nil, fmt.Errorf("piped input needs a terminal for controls: %w", err)
	}
	return tokens, tty, nil
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

With no file and no piped input, rat opens a file picker — press f to choose.

Flags:
  --wpm N        words per minute (100-1000)
  --chunk N      words per flash (1-3)
  --theme NAME   dark | light | solarized | high-contrast
  --adaptive B   true or false — longer pauses at punctuation and long words

Controls (in-reader, press ? for the full list):
  space play/pause · ←/→ seek · ↑/↓ speed · [ ] chunk · t theme · a adaptive · f file · q quit
`)
}
