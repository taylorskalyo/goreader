package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/gdamore/tcell/v2"
	"golang.org/x/net/html/atom"
	"gopkg.in/yaml.v3"
)

var (
	configDir    = xdg.ConfigHome
	appConfigDir = filepath.Join(configDir, "goreader")
	ConfigFile   = filepath.Join(appConfigDir, "config.yml")
)

type Theme map[string]Style

// Config stores configuration options.
type Config struct {
	Keybindings Keybindings `yaml:"keybindings"`
	Theme       Theme       `yaml:"theme"`
}

// Style controls an individual element's visual appearance when rendered.
type Style struct {
	Bold          *bool `yaml:"bold,omitempty"`
	Italic        *bool `yaml:"italic,omitempty"`
	StrikeThrough *bool `yaml:"strikethrough,omitempty"`
	Underline     *bool `yaml:"underline,omitempty"`

	Foreground *string `yaml:"foreground,omitempty"`
	Background *string `yaml:"background,omitempty"`
}

// Merge returns a new Style with attributes from other applied if present.
func (s Style) Merge(other Style) Style {
	if out, err := yaml.Marshal(other); err == nil {
		_ = yaml.Unmarshal(out, &s)
	}

	return s
}

func (s Style) String() string {
	var b strings.Builder
	b.WriteString("[")

	if s.Foreground != nil {
		b.WriteString(*s.Foreground)
	}

	b.WriteString(":")

	if s.Background != nil {
		b.WriteString(*s.Background)
	}

	b.WriteString(":")

	for flag, val := range map[string]*bool{
		"b": s.Bold,
		"i": s.Italic,
		"s": s.StrikeThrough,
		"u": s.Underline,
	} {
		if val != nil {
			if *val {
				b.WriteString(strings.ToLower(flag))
			} else {
				b.WriteString(strings.ToUpper(flag))
			}
		}
	}

	b.WriteString("]")

	return b.String()
}

// Keybindings map key presses to actions.
type Keybindings map[KeyChord]Action

// Default is the default configuration.
func Default() Config {
	return Config{
		Keybindings: DefaultKeybindings(),
		Theme:       DefaultTheme(),
	}
}

// DefaultKeybindings is the default keybinding.
func DefaultKeybindings() Keybindings {
	return Keybindings{
		KeyChord{Key: tcell.KeyDown}:  ActionDown,
		KeyChord{Key: tcell.KeyUp}:    ActionUp,
		KeyChord{Key: tcell.KeyLeft}:  ActionLeft,
		KeyChord{Key: tcell.KeyRight}: ActionRight,
		KeyChord{Key: tcell.KeyHome}:  ActionTop,
		KeyChord{Key: tcell.KeyEnd}:   ActionBottom,
		KeyChord{Key: tcell.KeyEsc}:   ActionExit,
		KeyChord{Key: tcell.KeyPgDn}:  ActionForward,
		KeyChord{Key: tcell.KeyPgUp}:  ActionBackward,

		KeyChord{Key: tcell.KeyRune, Rune: 'j'}: ActionDown,
		KeyChord{Key: tcell.KeyRune, Rune: 'k'}: ActionUp,
		KeyChord{Key: tcell.KeyRune, Rune: 'h'}: ActionLeft,
		KeyChord{Key: tcell.KeyRune, Rune: 'l'}: ActionRight,
		KeyChord{Key: tcell.KeyRune, Rune: 'g'}: ActionTop,
		KeyChord{Key: tcell.KeyRune, Rune: 'G'}: ActionBottom,
		KeyChord{Key: tcell.KeyRune, Rune: 'q'}: ActionExit,
		KeyChord{Key: tcell.KeyRune, Rune: 'f'}: ActionForward,
		KeyChord{Key: tcell.KeyRune, Rune: 'b'}: ActionBackward,
		KeyChord{Key: tcell.KeyRune, Rune: 'L'}: ActionChapterNext,
		KeyChord{Key: tcell.KeyRune, Rune: 'H'}: ActionChapterPrevious,
	}
}

// DefaultTheme is the default theme.
func DefaultTheme() Theme {
	bold := Style{Bold: pBool(true)}
	headingGeneric := Style{Foreground: pString(tcell.ColorTeal.Name())}

	return Theme{
		atom.Strong.String(): bold,
		atom.Em.String():     bold,
		atom.B.String():      bold,
		atom.I.String(): Style{
			Italic:     pBool(true),
			Foreground: pString(tcell.ColorOlive.Name()),
		},
		atom.Title.String(): Style{
			Foreground: pString(tcell.ColorMaroon.Name()),
		},
		atom.H1.String(): Style{
			Foreground: pString(tcell.ColorPurple.Name()),
		},
		atom.H2.String(): Style{
			Foreground: pString(tcell.ColorNavy.Name()),
		},
		atom.H3.String(): headingGeneric,
		atom.H4.String(): headingGeneric,
		atom.H5.String(): headingGeneric,
		atom.H6.String(): headingGeneric,
	}
}

func DefaultStyle() Style {
	return Style{
		Foreground: pString("-"),
		Background: pString("-"),

		Bold:          pBool(false),
		Italic:        pBool(false),
		StrikeThrough: pBool(false),
		Underline:     pBool(false),
	}
}

// Load reads configuration options from a file. Default values will be used
// for any options not specified in the file.
func Load() (*Config, error) {
	config := Default()
	data, err := os.ReadFile(ConfigFile)
	if os.IsNotExist(err) {
		return &config, nil
	}

	if err != nil {
		return &config, err
	}

	err = yaml.Unmarshal(data, &config)

	return &config, err
}

func pBool(b bool) *bool {
	return &b
}

func pString(s string) *string {
	return &s
}
