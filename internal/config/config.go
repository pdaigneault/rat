// Package config loads and persists the reader's preferences as XDG TOML at
// ~/.config/rat/config.toml. Preferences that change during a session (WPM,
// theme, chunk size, adaptive) are saved immediately so choices survive a crash.
package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/adrg/xdg"
)

// Bounds and defaults, exported so flags and the TUI clamp to the same range.
const (
	MinWPM     = 100
	MaxWPM     = 1000
	DefaultWPM = 300

	MinChunk     = 1
	MaxChunk     = 3
	DefaultChunk = 1

	DefaultTheme    = "dark"
	DefaultAdaptive = true
)

// Config is the persisted preference set. TOML keys are lower-case field names.
type Config struct {
	WPM       int    `toml:"wpm"`
	ChunkSize int    `toml:"chunk_size"`
	Theme     string `toml:"theme"`
	Adaptive  bool   `toml:"adaptive"`
}

// Defaults returns a Config populated with the built-in defaults.
func Defaults() Config {
	return Config{
		WPM:       DefaultWPM,
		ChunkSize: DefaultChunk,
		Theme:     DefaultTheme,
		Adaptive:  DefaultAdaptive,
	}
}

// Path returns the config file location, honouring XDG_CONFIG_HOME.
func Path() string {
	return filepath.Join(xdg.ConfigHome, "rat", "config.toml")
}

// Load returns defaults merged with the config file if it exists. A missing file
// is not an error: the defaults are returned and will be written lazily on the
// first Save. Values are clamped so a hand-edited file cannot push the reader
// out of its valid range.
func Load() (Config, error) {
	cfg := Defaults()
	_, err := toml.DecodeFile(Path(), &cfg)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}
	cfg.clamp()
	return cfg, nil
}

// Save writes the config to disk, creating the directory tree if needed.
func (c Config) Save() error {
	path := Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(c)
}

// clamp forces every field into its valid range, repairing bad file values.
func (c *Config) clamp() {
	if c.WPM < MinWPM {
		c.WPM = MinWPM
	}
	if c.WPM > MaxWPM {
		c.WPM = MaxWPM
	}
	if c.ChunkSize < MinChunk {
		c.ChunkSize = MinChunk
	}
	if c.ChunkSize > MaxChunk {
		c.ChunkSize = MaxChunk
	}
	if c.Theme == "" {
		c.Theme = DefaultTheme
	}
}
