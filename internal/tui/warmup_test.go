package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/pdaigneault/rat/internal/config"
	"github.com/pdaigneault/rat/internal/parser"
)

func manyTokens(n int) []parser.Token {
	toks := make([]parser.Token, n)
	for i := range toks {
		toks[i] = parser.Token{Text: "word"}
	}
	return toks
}

func adv(m Model) Model {
	next, _ := m.advance()
	return next.(Model)
}

func TestWarmupStartsAtDefault(t *testing.T) {
	cfg := config.Defaults()
	cfg.WPM = 500
	cfg.Warmup = true
	m := New(manyTokens(60), cfg)
	if !m.ramping {
		t.Fatal("expected to be ramping when warmup on and target > default")
	}
	if got := m.effectiveWPM(); got != config.DefaultWPM {
		t.Errorf("ramp should start at %d, got %d", config.DefaultWPM, got)
	}
}

func TestWarmupNoRampWhenNotNeeded(t *testing.T) {
	cases := []struct {
		name   string
		wpm    int
		warmup bool
	}{
		{"target equals default", 300, true},
		{"target below default", 200, true},
		{"warmup disabled", 500, false},
	}
	for _, c := range cases {
		cfg := config.Defaults()
		cfg.WPM = c.wpm
		cfg.Warmup = c.warmup
		m := New(manyTokens(10), cfg)
		if m.ramping {
			t.Errorf("%s: should not be ramping", c.name)
		}
		if got := m.effectiveWPM(); got != c.wpm {
			t.Errorf("%s: effective wpm should be %d, got %d", c.name, c.wpm, got)
		}
	}
}

func TestWarmupRampsToTarget(t *testing.T) {
	cfg := config.Defaults()
	cfg.WPM = 500
	cfg.Warmup = true
	m := New(manyTokens(80), cfg)

	// After the first rampStepEvery advances, speed rises by one rampStep.
	for i := 0; i < rampStepEvery; i++ {
		m = adv(m)
	}
	if want := config.DefaultWPM + rampStep; m.effectiveWPM() != want {
		t.Errorf("after one interval expected %d wpm, got %d", want, m.effectiveWPM())
	}

	// Keep advancing until the ramp must have completed.
	steps := ((500-config.DefaultWPM)/rampStep + 1) * rampStepEvery
	for i := 0; i < steps; i++ {
		m = adv(m)
	}
	if m.ramping {
		t.Error("ramp should have finished")
	}
	if m.effectiveWPM() != 500 {
		t.Errorf("effective wpm should reach target 500, got %d", m.effectiveWPM())
	}
}

func TestWarmupStopsWhenTargetLoweredBelowRamp(t *testing.T) {
	cfg := config.Defaults()
	cfg.WPM = 800
	cfg.Warmup = true
	m := New(manyTokens(200), cfg)

	// Advance one interval so the ramp is above the default (325 wpm).
	for i := 0; i < rampStepEvery; i++ {
		m = adv(m)
	}
	// Lower the target below the current ramp speed with the ↓ key; the ramp
	// should collapse to the new target immediately.
	down := tea.KeyPressMsg{Code: tea.KeyDown}
	for m.wpm > config.DefaultWPM {
		next, _ := m.Update(down)
		m = next.(Model)
	}
	if m.ramping {
		t.Error("lowering the target below the ramp speed should end warm-up")
	}
	if m.effectiveWPM() != m.wpm {
		t.Errorf("effective wpm should equal target %d, got %d", m.wpm, m.effectiveWPM())
	}
}
