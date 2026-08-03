package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/pdaigneault/rat/internal/reader"
	"github.com/pdaigneault/rat/internal/theme"
)

// anchorCol is the fixed column (within the frame) where the ORP pivot rune sits.
// Slightly left of centre matches how the eye naturally lands on a word.
const anchorCol = frameWidth/2 - 2

// View renders the reader. In Bubble Tea v2 the alt-screen and terminal colours
// are set on the returned View rather than via program options.
func (m Model) View() tea.View {
	th := theme.Get(m.themeName)
	v := tea.NewView(m.render(th))
	v.AltScreen = true
	v.BackgroundColor = th.Bg
	v.ForegroundColor = th.Text
	return v
}

// render assembles the whole screen for the current state.
func (m Model) render(th theme.Theme) string {
	if !m.ready {
		return ""
	}
	if m.browsing {
		return m.renderBrowser(th)
	}
	if !m.hasDoc || len(m.chunks) == 0 {
		return m.renderEmpty(th)
	}

	textStyle := lipgloss.NewStyle().Foreground(th.Text)
	pivotStyle := lipgloss.NewStyle().Foreground(th.Pivot).Bold(true)
	frameStyle := lipgloss.NewStyle().Foreground(th.Frame)
	dimStyle := lipgloss.NewStyle().Foreground(th.Dim)

	mark := strings.Repeat(" ", anchorCol)
	topGuide := mark + frameStyle.Render("▼")
	botGuide := mark + frameStyle.Render("▲")
	wordLine := m.renderWord(th, textStyle, pivotStyle)

	fraction := float64(m.idx+1) / float64(len(m.chunks))
	bar := m.prog.ViewAs(fraction)

	lines := []string{
		topGuide,
		wordLine,
		botGuide,
		"",
		bar,
		dimStyle.Render(m.statLine()),
	}
	lines = append(lines, m.footer(dimStyle)...)

	// A fixed-width container keeps the block a constant size so centring never
	// shifts the pivot column between frames.
	frame := lipgloss.NewStyle().Width(frameWidth).Render(strings.Join(lines, "\n"))
	return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, frame)
}

// renderEmpty is the screen shown when rat starts with no file: a centred prompt
// pointing the user at the file picker.
func (m Model) renderEmpty(th theme.Theme) string {
	title := lipgloss.NewStyle().Foreground(th.Text).Bold(true).Render("rat")
	pivot := lipgloss.NewStyle().Foreground(th.Pivot).Bold(true)
	dim := lipgloss.NewStyle().Foreground(th.Dim)
	body := strings.Join([]string{
		title,
		"",
		dim.Render("no file loaded"),
		"",
		"press " + pivot.Render("f") + dim.Render(" to choose a file"),
		dim.Render("q to quit"),
	}, "\n")
	block := lipgloss.NewStyle().Align(lipgloss.Center).Render(body)
	return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, block)
}

// browserWidth is the fixed width of the file-explorer panel.
const browserWidth = 52

// renderBrowser draws the file explorer: the current directory, the filtered
// entry list with a cursor, a key hint, and any transient notice.
func (m Model) renderBrowser(th theme.Theme) string {
	dim := lipgloss.NewStyle().Foreground(th.Dim)
	dirStyle := lipgloss.NewStyle().Foreground(th.Frame).Bold(true)
	cursorStyle := lipgloss.NewStyle().Foreground(th.Pivot).Bold(true)
	pivotStyle := lipgloss.NewStyle().Foreground(th.Pivot)

	e := m.explorer
	header := dirStyle.Render(truncLeft(compactPath(e.dir), browserWidth-2))
	hint := dim.Render("↑↓ move · → open · ← up · enter select · esc cancel")

	// Reserve rows for the header, hint, notice and padding.
	maxRows := m.h - 8
	if maxRows < 3 {
		maxRows = 3
	}
	start, end := scrollWindow(e.cursor, len(e.entries), maxRows)

	var rows []string
	if len(e.entries) == 0 {
		rows = append(rows, dim.Render("(no folders or supported files here)"))
	}
	for i := start; i < end; i++ {
		en := e.entries[i]
		label := en.name
		if en.isDir {
			label += "/"
		}
		style := lipgloss.NewStyle().Foreground(th.Text)
		if en.isDir {
			style = lipgloss.NewStyle().Foreground(th.Frame)
		}
		prefix := "  "
		if i == e.cursor {
			prefix = cursorStyle.Render("▸ ")
			style = style.Bold(true)
		}
		rows = append(rows, prefix+style.Render(label))
	}

	parts := []string{header, "", strings.Join(rows, "\n"), "", hint}
	if e.failed != "" {
		parts = append(parts, pivotStyle.Render(e.failed))
	}
	if m.notice != "" {
		parts = append(parts, pivotStyle.Render(m.notice))
	}
	panel := lipgloss.NewStyle().Width(min(browserWidth, m.w)).Render(strings.Join(parts, "\n"))
	return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, panel)
}

// scrollWindow returns the [start,end) slice of a list of length n to display in
// rows lines, keeping the cursor roughly centred.
func scrollWindow(cursor, n, rows int) (int, int) {
	if n <= rows {
		return 0, n
	}
	start := cursor - rows/2
	if start < 0 {
		start = 0
	}
	if start+rows > n {
		start = n - rows
	}
	return start, start + rows
}

// truncLeft trims s from the left to at most max display columns, prefixing an
// ellipsis — so a long path keeps its most specific (rightmost) part visible.
func truncLeft(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return "…" + string(r[len(r)-(max-1):])
}

// renderWord lays out the current chunk with its pivot rune highlighted and the
// preceding text left-padded so the pivot lands in anchorCol.
func (m Model) renderWord(th theme.Theme, textStyle, pivotStyle lipgloss.Style) string {
	left, pivot, right := reader.Split(m.chunks[m.idx].Text())
	pad := anchorCol - lipgloss.Width(left)
	if pad < 0 {
		pad = 0
	}
	return strings.Repeat(" ", pad) +
		textStyle.Render(left) +
		pivotStyle.Render(pivot) +
		textStyle.Render(right)
}

// statLine is the single-line status: speed, position, percent, theme, mode and
// a play-state glyph.
func (m Model) statLine() string {
	pct := int(float64(m.idx+1) / float64(len(m.chunks)) * 100)
	mode := "steady"
	if m.adaptive {
		mode = "adaptive"
	}
	state := "▶"
	switch {
	case m.finished:
		state = "✓"
	case !m.playing:
		state = "⏸"
	}
	return fmt.Sprintf("%s  %d wpm · %d/%d · %d%% · %s · %s",
		state, m.wpm, m.idx+1, len(m.chunks), pct, m.themeName, mode)
}

// footer returns the help lines. Collapsed, it is a one-line hint; expanded (via
// `?`), it lists every binding.
func (m Model) footer(dim lipgloss.Style) []string {
	if !m.showHelp {
		return []string{dim.Render("? help")}
	}
	parts := make([]string, len(helpEntries))
	for i, e := range helpEntries {
		parts[i] = fmt.Sprintf("%s %s", e.keys, e.desc)
	}
	// Two per line keeps the footer inside the frame width.
	var lines []string
	for i := 0; i < len(parts); i += 2 {
		end := i + 2
		if end > len(parts) {
			end = len(parts)
		}
		lines = append(lines, dim.Render(strings.Join(parts[i:end], "   ")))
	}
	return append([]string{""}, lines...)
}
