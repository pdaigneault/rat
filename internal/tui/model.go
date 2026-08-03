// Package tui is the Bubble Tea front end: it drives the RSVP playback loop,
// handles keys, persists preference changes, and renders the reader.
package tui

import (
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"

	"github.com/paul-daigneault/rat/internal/config"
	"github.com/paul-daigneault/rat/internal/parser"
	"github.com/paul-daigneault/rat/internal/reader"
	"github.com/paul-daigneault/rat/internal/theme"
)

// wpmStep is how much ↑/↓ change the speed by.
const wpmStep = 25

// frameWidth is the fixed width of the reading frame. A constant width keeps the
// ORP pivot in the same screen column for every word — the whole point of a
// gaze anchor.
const frameWidth = 44

// tickMsg advances playback. It carries the epoch it was scheduled under so the
// model can discard ticks left over from before a pause, seek, or speed change.
type tickMsg struct{ epoch int }

// Model is the reader's state. tokens is retained so the stream can be re-chunked
// live when the chunk size changes.
type Model struct {
	tokens []parser.Token
	chunks []reader.Chunk

	idx       int
	playing   bool
	finished  bool
	wpm       int
	chunkSize int
	adaptive  bool
	themeName string
	showHelp  bool

	epoch int // bumped whenever timing is invalidated; guards stale ticks

	prog  progress.Model
	w, h  int
	ready bool
	cfg   config.Config
}

// New builds a Model from a parsed token stream and the loaded config.
func New(tokens []parser.Token, cfg config.Config) Model {
	m := Model{
		tokens:    tokens,
		idx:       0,
		playing:   true,
		wpm:       cfg.WPM,
		chunkSize: cfg.ChunkSize,
		adaptive:  cfg.Adaptive,
		themeName: cfg.Theme,
		cfg:       cfg,
	}
	m.chunks = reader.Group(tokens, m.chunkSize)
	m.prog = newProgress(theme.Get(m.themeName))
	return m
}

// newProgress builds a static (non-animated) progress bar coloured for the
// theme. We drive it with ViewAs, so no spring animation competes with our word
// ticks.
func newProgress(th theme.Theme) progress.Model {
	return progress.New(
		progress.WithColors(th.ProgressFull),
		progress.WithFillCharacters('█', '░'),
		progress.WithoutPercentage(),
	)
}

// Init starts playback: the first tick is scheduled immediately so the reader
// begins as soon as the first WindowSizeMsg gives us a canvas.
func (m Model) Init() tea.Cmd {
	return m.scheduleTick()
}

// scheduleTick returns a command that fires after the current chunk's delay,
// tagged with the current epoch. It returns nil when playback is not running,
// so a paused or finished reader schedules nothing.
func (m Model) scheduleTick() tea.Cmd {
	if !m.playing || m.finished || len(m.chunks) == 0 {
		return nil
	}
	d := reader.Delay(m.chunks[m.idx], m.wpm, m.adaptive)
	ep := m.epoch
	return tea.Tick(d, func(time.Time) tea.Msg { return tickMsg{epoch: ep} })
}

// invalidate bumps the epoch so any in-flight tick is ignored, then returns a
// freshly scheduled tick (or nil if not playing). Call this after any change
// that affects timing: pause/resume, seek, speed, chunk size, adaptive.
func (m *Model) invalidate() tea.Cmd {
	m.epoch++
	return m.scheduleTick()
}

// Update is the Bubble Tea reducer.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		m.ready = true
		m.prog.SetWidth(min(frameWidth, m.w))
		return m, nil

	case tickMsg:
		if msg.epoch != m.epoch {
			return m, nil // stale tick from before a state change
		}
		return m.advance()

	case tea.KeyPressMsg:
		return m.handleKey(msg.String())
	}
	return m, nil
}

// advance moves to the next chunk, stopping (and pausing) at the end.
func (m Model) advance() (tea.Model, tea.Cmd) {
	if m.idx+1 >= len(m.chunks) {
		m.idx = len(m.chunks) - 1
		m.finished = true
		m.playing = false
		return m, nil
	}
	m.idx++
	return m, m.scheduleTick()
}

// handleKey applies a keypress. Mutating keys persist the config so a choice
// survives even an abrupt exit.
func (m Model) handleKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case keyQuit, keyQuitCtrl, keyQuitEsc:
		m.save()
		return m, tea.Quit

	case keyPlayPause:
		if m.finished {
			// Restart from the top when finished.
			m.idx = 0
			m.finished = false
		}
		m.playing = !m.playing
		return m, m.invalidate()

	case keySeekBack:
		m.seek(m.idx - 1)
		return m, m.invalidate()
	case keySeekFwd:
		m.seek(m.idx + 1)
		return m, m.invalidate()
	case keySentenceBack:
		m.seek(m.prevSentence())
		return m, m.invalidate()
	case keySentenceFwd:
		m.seek(m.nextSentence())
		return m, m.invalidate()
	case keyRestart:
		m.seek(0)
		return m, m.invalidate()

	case keyFaster:
		m.wpm = clamp(m.wpm+wpmStep, config.MinWPM, config.MaxWPM)
		m.save()
		return m, m.invalidate()
	case keySlower:
		m.wpm = clamp(m.wpm-wpmStep, config.MinWPM, config.MaxWPM)
		m.save()
		return m, m.invalidate()

	case keyChunkDown:
		return m.rechunk(m.chunkSize - 1)
	case keyChunkUp:
		return m.rechunk(m.chunkSize + 1)

	case keyTheme:
		m.themeName = theme.Next(m.themeName)
		m.prog = newProgress(theme.Get(m.themeName))
		m.prog.SetWidth(min(frameWidth, m.w))
		m.save()
		return m, nil

	case keyAdaptive:
		m.adaptive = !m.adaptive
		m.save()
		return m, m.invalidate()

	case keyHelp:
		m.showHelp = !m.showHelp
		return m, nil
	}
	return m, nil
}

// seek jumps to a chunk index, clamping to range and clearing the finished flag.
func (m *Model) seek(to int) {
	m.idx = clamp(to, 0, len(m.chunks)-1)
	m.finished = false
}

// rechunk regroups the token stream at a new chunk size, preserving the reader's
// place by word offset, then reschedules. Size is clamped to 1..3.
func (m Model) rechunk(size int) (tea.Model, tea.Cmd) {
	size = clamp(size, config.MinChunk, config.MaxChunk)
	if size == m.chunkSize {
		return m, nil // no change (already at a bound)
	}
	word := wordOffset(m.chunks, m.idx)
	m.chunkSize = size
	m.chunks = reader.Group(m.tokens, size)
	m.idx = clamp(chunkAtWord(m.chunks, word), 0, len(m.chunks)-1)
	m.finished = false
	m.save()
	return m, m.invalidate()
}

// prevSentence returns the index of the nearest chunk before the current one
// that ends a sentence (so playback resumes at the next sentence's start), or 0.
func (m Model) prevSentence() int {
	for i := m.idx - 1; i > 0; i-- {
		if m.chunks[i-1].Boundary == parser.Sentence {
			return i
		}
	}
	return 0
}

// nextSentence returns the index of the first chunk after the current sentence's
// end, or the last chunk if none follows.
func (m Model) nextSentence() int {
	for i := m.idx; i < len(m.chunks); i++ {
		if m.chunks[i].Boundary == parser.Sentence {
			if i+1 < len(m.chunks) {
				return i + 1
			}
			return len(m.chunks) - 1
		}
	}
	return len(m.chunks) - 1
}

// save mirrors live state into the config and writes it. Errors are ignored: a
// failed preference write must never interrupt reading.
func (m *Model) save() {
	m.cfg.WPM = m.wpm
	m.cfg.ChunkSize = m.chunkSize
	m.cfg.Theme = m.themeName
	m.cfg.Adaptive = m.adaptive
	_ = m.cfg.Save()
}

// wordOffset returns the number of words before chunk idx.
func wordOffset(chunks []reader.Chunk, idx int) int {
	n := 0
	for i := 0; i < idx && i < len(chunks); i++ {
		n += len(chunks[i].Words)
	}
	return n
}

// chunkAtWord returns the index of the chunk containing the given word offset.
func chunkAtWord(chunks []reader.Chunk, word int) int {
	n := 0
	for i, c := range chunks {
		n += len(c.Words)
		if word < n {
			return i
		}
	}
	return len(chunks) - 1
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
