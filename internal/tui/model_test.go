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

// TestRestartKeys confirms both home and r jump playback back to the start.
func TestRestartKeys(t *testing.T) {
	tokens := []parser.Token{{Text: "one"}, {Text: "two"}, {Text: "three"}, {Text: "four"}}
	for _, key := range []tea.KeyPressMsg{
		{Code: tea.KeyHome},
		{Code: 'r', Text: "r"},
	} {
		m := New(tokens, config.Defaults())
		m.idx = 3 // pretend we've read to the end
		restarted, _ := m.Update(key)
		if got := restarted.(Model).idx; got != 0 {
			t.Errorf("key %q: expected idx 0 after restart, got %d", key.String(), got)
		}
	}
}

// TestEmptyStartNoPlayback covers launching rat with no document: it must not
// auto-play, schedule no ticks, and ignore playback keys.
func TestEmptyStartNoPlayback(t *testing.T) {
	m := New(nil, config.Defaults())
	if m.hasDoc {
		t.Fatal("no document expected when started empty")
	}
	if m.playing {
		t.Fatal("should not auto-play with no document")
	}
	if cmd := m.Init(); cmd != nil {
		t.Error("empty reader should schedule no tick")
	}
	updated, _ := m.Update(spaceKey)
	if updated.(Model).playing {
		t.Error("space must not start playback when there is no document")
	}
}

// TestBrowseKeyOpensPicker confirms f opens the file picker from the empty state.
func TestBrowseKeyOpensPicker(t *testing.T) {
	m := New(nil, config.Defaults())
	fKey := tea.KeyPressMsg{Code: 'f', Text: "f"}
	updated, cmd := m.Update(fKey)
	m2 := updated.(Model)
	if !m2.browsing {
		t.Error("f should open the file picker (browsing=true)")
	}
	if cmd == nil {
		t.Error("opening the picker should return an init command to read the directory")
	}
}

// TestQuitKeys confirms esc quits the reader just like q and ctrl+c: each should
// yield a command that produces a tea.QuitMsg.
func TestQuitKeys(t *testing.T) {
	tokens := []parser.Token{{Text: "one"}, {Text: "two"}}
	cases := map[string]tea.KeyPressMsg{
		"q":      {Code: 'q', Text: "q"},
		"ctrl+c": {Code: 'c', Mod: tea.ModCtrl},
		"esc":    {Code: tea.KeyEscape},
	}
	for name, key := range cases {
		if got := key.String(); got != name {
			t.Fatalf("expected key %q to stringify to itself, got %q", name, got)
		}
		m := New(tokens, config.Defaults())
		_, cmd := m.Update(key)
		if cmd == nil {
			t.Fatalf("%s: expected a quit command, got nil", name)
		}
		if _, ok := cmd().(tea.QuitMsg); !ok {
			t.Errorf("%s: command did not produce a tea.QuitMsg", name)
		}
	}
}
