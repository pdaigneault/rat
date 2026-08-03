// Package theme defines the colour palettes rat can render with and helpers to
// cycle between them live. Colours are resolved through lipgloss so they adapt
// to the terminal's colour profile.
package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Theme is a named palette. Text is the ordinary word colour, Pivot highlights
// the ORP rune, Dim is for secondary chrome (stats, help), Frame is the guide
// marks, Bg is the background, and the Progress colours fill the bar.
type Theme struct {
	Name          string
	Text          color.Color
	Pivot         color.Color
	Dim           color.Color
	Frame         color.Color
	Bg            color.Color
	ProgressFull  color.Color
	ProgressEmpty color.Color
}

// Names is the fixed cycle order for the `t` key. Keeping it as an explicit
// slice (rather than ranging a map) makes cycling deterministic.
var Names = []string{"dark", "light", "solarized", "high-contrast"}

// Builtins holds every palette keyed by name.
var Builtins = map[string]Theme{
	"dark": {
		Name:          "dark",
		Text:          lipgloss.Color("#e0e0e0"),
		Pivot:         lipgloss.Color("#ff5f5f"),
		Dim:           lipgloss.Color("#6c6c6c"),
		Frame:         lipgloss.Color("#5f87af"),
		Bg:            lipgloss.Color("#1c1c1c"),
		ProgressFull:  lipgloss.Color("#5f87af"),
		ProgressEmpty: lipgloss.Color("#3a3a3a"),
	},
	"light": {
		Name:          "light",
		Text:          lipgloss.Color("#1c1c1c"),
		Pivot:         lipgloss.Color("#c00000"),
		Dim:           lipgloss.Color("#8a8a8a"),
		Frame:         lipgloss.Color("#005f87"),
		Bg:            lipgloss.Color("#fdfdfd"),
		ProgressFull:  lipgloss.Color("#005f87"),
		ProgressEmpty: lipgloss.Color("#dadada"),
	},
	"solarized": {
		Name:          "solarized",
		Text:          lipgloss.Color("#93a1a1"),
		Pivot:         lipgloss.Color("#dc322f"),
		Dim:           lipgloss.Color("#586e75"),
		Frame:         lipgloss.Color("#268bd2"),
		Bg:            lipgloss.Color("#002b36"),
		ProgressFull:  lipgloss.Color("#2aa198"),
		ProgressEmpty: lipgloss.Color("#073642"),
	},
	"high-contrast": {
		Name:          "high-contrast",
		Text:          lipgloss.Color("#ffffff"),
		Pivot:         lipgloss.Color("#ffff00"),
		Dim:           lipgloss.Color("#c0c0c0"),
		Frame:         lipgloss.Color("#00ffff"),
		Bg:            lipgloss.Color("#000000"),
		ProgressFull:  lipgloss.Color("#ffff00"),
		ProgressEmpty: lipgloss.Color("#444444"),
	},
}

// Get returns the named theme, falling back to "dark" if the name is unknown
// (e.g. a stale value from an edited config file).
func Get(name string) Theme {
	if t, ok := Builtins[name]; ok {
		return t
	}
	return Builtins["dark"]
}

// Next returns the name of the theme after the given one in cycle order,
// wrapping around. An unknown name resolves to the first theme.
func Next(name string) string {
	for i, n := range Names {
		if n == name {
			return Names[(i+1)%len(Names)]
		}
	}
	return Names[0]
}
