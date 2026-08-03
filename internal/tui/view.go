package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/paul-daigneault/rat/internal/reader"
	"github.com/paul-daigneault/rat/internal/theme"
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

// renderBrowser draws the file picker with a title, a hint, and any transient
// notice (such as an unreadable-file message).
func (m Model) renderBrowser(th theme.Theme) string {
	dim := lipgloss.NewStyle().Foreground(th.Dim)
	title := lipgloss.NewStyle().Foreground(th.Frame).Bold(true).Render("Select a file")
	hint := dim.Render("md · txt   ↑↓ move · enter open · esc cancel")
	parts := []string{title, hint, "", m.picker.View()}
	if m.notice != "" {
		parts = append(parts, "", lipgloss.NewStyle().Foreground(th.Pivot).Render(m.notice))
	}
	return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, strings.Join(parts, "\n"))
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
