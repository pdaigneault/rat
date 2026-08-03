package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/paul-daigneault/rat/internal/config"
	"github.com/paul-daigneault/rat/internal/parser"
)

// spaceKey is the message Bubble Tea delivers for a space keypress.
var spaceKey = tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}

// TestSpaceKeyString pins our binding to the library's actual output. Bubble
// Tea v2 stringifies the space key as "space" (not " " as in v1); if that ever
// changes, this fails loudly rather than silently breaking play/pause.
func TestSpaceKeyString(t *testing.T) {
	if got := spaceKey.String(); got != keyPlayPause {
		t.Fatalf("space key stringifies to %q, but keyPlayPause is %q", got, keyPlayPause)
	}
}

// TestSpaceTogglesPlaying drives the real Update path to confirm space pauses
// and resumes playback.
func TestSpaceTogglesPlaying(t *testing.T) {
	tokens := []parser.Token{{Text: "one"}, {Text: "two"}, {Text: "three"}}
	m := New(tokens, config.Defaults())
	if !m.playing {
		t.Fatal("reader should start playing")
	}

	paused, _ := m.Update(spaceKey)
	if paused.(Model).playing {
		t.Error("space should have paused playback")
	}

	resumed, _ := paused.(Model).Update(spaceKey)
	if !resumed.(Model).playing {
		t.Error("space should have resumed playback")
	}
}
