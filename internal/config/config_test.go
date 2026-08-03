package config

import (
	"testing"

	"github.com/adrg/xdg"
)

// withTempConfig points XDG at a fresh temp dir. adrg/xdg caches paths at init,
// so we must Reload after changing the environment.
func withTempConfig(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	xdg.Reload()
}

func TestDefaultsWarmupOn(t *testing.T) {
	if !Defaults().Warmup {
		t.Error("warmup should default to on")
	}
}

func TestLoadMissingReturnsDefaults(t *testing.T) {
	withTempConfig(t)
	got, err := Load()
	if err != nil {
		t.Fatalf("Load on missing file: %v", err)
	}
	if got != Defaults() {
		t.Errorf("missing file should yield defaults, got %+v", got)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	withTempConfig(t)
	want := Config{WPM: 450, ChunkSize: 2, Theme: "solarized", Adaptive: false}
	if err := want.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != want {
		t.Errorf("round trip mismatch: got %+v, want %+v", got, want)
	}
}

func TestLoadClampsBadValues(t *testing.T) {
	withTempConfig(t)
	bad := Config{WPM: 99999, ChunkSize: 7, Theme: "", Adaptive: true}
	if err := bad.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.WPM != MaxWPM {
		t.Errorf("WPM should clamp to %d, got %d", MaxWPM, got.WPM)
	}
	if got.ChunkSize != MaxChunk {
		t.Errorf("ChunkSize should clamp to %d, got %d", MaxChunk, got.ChunkSize)
	}
	if got.Theme != DefaultTheme {
		t.Errorf("empty theme should default to %q, got %q", DefaultTheme, got.Theme)
	}
}
