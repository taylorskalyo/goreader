package config

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

var (
	// ModifierNames holds the written names of modifier keys. Useful to echo
	// back a key name, or to look up a key from a string value.
	ModifierNames = map[tcell.ModMask]string{
		tcell.ModShift: "Shift",
		tcell.ModCtrl:  "Ctrl",
		tcell.ModAlt:   "Alt",
		tcell.ModMeta:  "Meta",
	}

	namedKeys      = map[string]tcell.Key{}
	namedModifiers = map[string]tcell.ModMask{}
)

func init() {
	// Get mapping of string -> Key.
	namedKeys = make(map[string]tcell.Key, len(tcell.KeyNames))
	for k, v := range tcell.KeyNames {
		namedKeys[strings.ToLower(v)] = k
	}

	// Get mapping of string -> ModMask.
	namedModifiers = make(map[string]tcell.ModMask, len(ModifierNames))
	for k, v := range ModifierNames {
		namedModifiers[strings.ToLower(v)] = k
	}
}

// KeyChord represents a sequence of simultaneous key presses.
type KeyChord struct {
	ModMask tcell.ModMask
	Key     tcell.Key
	Rune    rune
}

// KeyChordFromEvent creates a new KeyChord from a tcell EventKey.
func KeyChordFromEvent(ev tcell.EventKey) KeyChord {
	return KeyChord{
		ModMask: ev.Modifiers(),
		Key:     ev.Key(),
		Rune:    ev.Rune(),
	}
}

// Event creates a new tcell.EventKey from a KeyChord.
func (kc KeyChord) Event() *tcell.EventKey {
	return tcell.NewEventKey(kc.Key, kc.Rune, kc.ModMask)
}

// UnmarshalText creates a new KeyChord from text. In it's textual form, a
// KeyChord uses the following format:
//
// [modifiers... +] [key | rune]
//
// Modifier and key names are case-insensitive. There can be any number of
// supported modifiers, and there must be exactly one key or rune. A "+"
// character must separate each element in the sequence.
//
// For example:
//
// CTRL + ALT + SPACE
// Meta + x
// esc
func (kc *KeyChord) UnmarshalText(text []byte) error {
	*kc = KeyChord{}
	keys := strings.Split(string(text), "+")

	strModifiers := keys[:len(keys)-1]
	for _, strMod := range strModifiers {
		if mod, ok := namedModifiers[strings.ToLower(strMod)]; ok {
			(*kc).ModMask |= mod
		} else {
			return fmt.Errorf("config: unrecognized keybind modifier \"%s\"", strMod)
		}
	}

	strKey := keys[len(keys)-1]
	ctrlKey := fmt.Sprintf("Ctrl-%s", strKey)
	isCtrl := (*kc).ModMask&tcell.ModCtrl != 0
	if key, ok := namedKeys[strings.ToLower(ctrlKey)]; ok && isCtrl {
		(*kc).Key = key
		(*kc).Rune = rune(key)
	} else if key, ok := namedKeys[strings.ToLower(strKey)]; ok {
		(*kc).Key = key
	} else if runes := []rune(strKey); len(runes) == 1 {
		(*kc).Key = tcell.KeyRune
		(*kc).Rune = runes[0]
	} else {
		return fmt.Errorf("config: unrecognized keybind key \"%s\"", strKey)
	}

	return nil
}

// MarshalText renders a KeyChord as text.
func (kc KeyChord) MarshalText() ([]byte, error) {
	return []byte(kc.String()), nil
}

// String renders a KeyChord as text.
func (kc KeyChord) String() string {
	keys := kc.modNames()
	if kc.Key == tcell.KeyRune {
		keys = append(keys, string(kc.Rune))
	} else {
		name := tcell.KeyNames[kc.Key]
		if kc.ModMask&tcell.ModCtrl != 0 && strings.HasPrefix(name, "Ctrl-") {
			name = strings.ToLower(name[5:])
		}
		keys = append(keys, name)
	}

	return strings.Join(keys, "+")
}

func (kc KeyChord) modNames() []string {
	modNames := []string{}
	for name, mod := range namedModifiers {
		if (kc.ModMask & mod) != 0 {
			modNames = append(modNames, name)
		}
	}

	return modNames
}
