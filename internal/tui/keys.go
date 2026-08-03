package tui

// Key strings as reported by tea.KeyPressMsg.String() in Bubble Tea v2. Keeping
// them here as named constants gives the Update switch and the help footer a
// single source of truth.
const (
	keyPlayPause    = "space" // Bubble Tea v2 stringifies the space key as "space", not " "
	keyQuit         = "q"
	keyQuitCtrl     = "ctrl+c"
	keyQuitEsc      = "esc"
	keySeekBack     = "left"
	keySeekFwd      = "right"
	keySentenceBack = "shift+left"
	keySentenceFwd  = "shift+right"
	keyRestart      = "home"
	keyFaster       = "up"
	keySlower       = "down"
	keyChunkDown    = "["
	keyChunkUp      = "]"
	keyTheme        = "t"
	keyAdaptive     = "a"
	keyBrowse       = "f"
	keyHelp         = "?"
)

// helpEntry pairs a key label with what it does, for the toggled footer.
type helpEntry struct {
	keys string
	desc string
}

// helpEntries drives the expanded help footer in reading order.
var helpEntries = []helpEntry{
	{"space", "play/pause"},
	{"←/→", "seek ±1"},
	{"⇧←/→", "±1 sentence"},
	{"home", "restart"},
	{"↑/↓", "speed ±25"},
	{"[ ]", "chunk 1–3"},
	{"t", "theme"},
	{"a", "adaptive"},
	{"f", "open file"},
	{"?", "help"},
	{"q/esc", "quit"},
}
