package styles

import (
	"os"
	"strings"
)

// Icon represents an icon with Nerd Font and Unicode fallback
type Icon struct {
	NerdFont string
	Fallback string
}

// Resource icons
var (
	IconEvents   = Icon{NerdFont: "󰡸", Fallback: "📡"}
	IconPersons  = Icon{NerdFont: "󰀄", Fallback: "👤"}
	IconFlags    = Icon{NerdFont: "", Fallback: "🚩"}
	IconSettings = Icon{NerdFont: "", Fallback: "⚙️"}
	IconHogQL    = Icon{NerdFont: "", Fallback: "🗄️"}

	// Navigation icons
	IconCollapse = Icon{NerdFont: "", Fallback: "◀"}
	IconExpand   = Icon{NerdFont: "", Fallback: "▶"}
	IconUp       = Icon{NerdFont: "", Fallback: "↑"}
	IconDown     = Icon{NerdFont: "", Fallback: "↓"}

	// Status icons
	IconEnabled  = Icon{NerdFont: "", Fallback: "✓"}
	IconDisabled = Icon{NerdFont: "", Fallback: "✗"}
	IconWarning  = Icon{NerdFont: "", Fallback: "⚠"}
	IconInfo     = Icon{NerdFont: "", Fallback: "ℹ"}

	// Environment icons
	IconDev  = Icon{NerdFont: "", Fallback: "DEV"}
	IconProd = Icon{NerdFont: "", Fallback: "PROD"}
)

// String returns the appropriate icon representation based on Nerd Font support
func (i Icon) String() string {
	if UseNerdFonts() {
		return i.NerdFont
	}
	return i.Fallback
}

// UseNerdFonts detects if the terminal supports Nerd Fonts
// This is a heuristic based on common terminal emulators and TERM variable
func UseNerdFonts() bool {
	// Check for explicit environment variable
	if nf := os.Getenv("NERD_FONTS"); nf != "" {
		return nf == "1" || strings.ToLower(nf) == "true"
	}

	// Detect common terminals that support Nerd Fonts
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")

	// Known Nerd Font-friendly terminals
	nerdFontTerminals := []string{
		"iTerm.app",
		"Alacritty",
		"kitty",
		"WezTerm",
	}

	for _, t := range nerdFontTerminals {
		if strings.Contains(termProgram, t) {
			return true
		}
	}

	// If TERM includes "256color" or "truecolor", assume modern terminal
	if strings.Contains(term, "256color") || strings.Contains(term, "truecolor") {
		return true
	}

	// Default to fallback for safety
	return false
}
