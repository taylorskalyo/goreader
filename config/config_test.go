package config

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html/atom"
	"gopkg.in/yaml.v3"
)

func TestDefault(t *testing.T) {
	if data, err := yaml.Marshal(Default()); assert.NoError(t, err, "Default config should be marshal-able") {
		t.Logf("Default config:\n%s", data)
		var actual Config
		if err := yaml.Unmarshal(data, &actual); assert.NoError(t, err, "Default config should be unmarshal-able") {
			assert.Equal(t, Default(), actual)
		}
	}
}

func TestUnmarshal(t *testing.T) {
	testCases := []struct {
		name     string
		yaml     []byte
		expected Config
	}{
		{
			"KeyChordModiferVariations",
			[]byte(`keybindings:
  "Ctrl+ALT+PgUp": Up`),
			Config{
				Keybindings: Keybindings{
					KeyChord{
						Key:     tcell.KeyPgUp,
						ModMask: tcell.ModCtrl | tcell.ModAlt,
					}: ActionUp,
				},
			},
		},
		{
			"StyleColorVariations",
			[]byte(`theme:
  em:
      foreground: red
      background: "#111213"`),
			Config{
				Theme: Theme{
					atom.Em.String(): Style{
						Foreground: pString(tcell.ColorRed.String()),
						Background: pString(tcell.NewRGBColor(0x11, 0x12, 0x13).String()),
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			var actual Config
			if err := yaml.Unmarshal(tc.yaml, &actual); assert.NoError(tt, err) {
				assert.Equal(tt, tc.expected, actual)
			}
		})
	}
}

func TestUnmarshalError(t *testing.T) {
	testCases := []struct {
		name     string
		yaml     []byte
		expected string
	}{
		{
			"BadKey",
			[]byte(`keybindings:
  "Ctrl+Alt+ThisKeyDoesntExist": Up`),
			"unrecognized keybind key",
		},
		{
			"BadModifier",
			[]byte(`keybindings:
  "ThisModifierDoesntExist+j": Up`),
			"unrecognized keybind modifier",
		},
		{
			"BadAction",
			[]byte(`keybindings:
  "x": ThisActionDoesntExist`),
			"unrecognized event",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			var actual Config
			err := yaml.Unmarshal(tc.yaml, &actual)
			assert.ErrorContains(tt, err, tc.expected)
		})
	}
}

func TestStyleMerge(t *testing.T) {
	expected := Style{
		Bold:       pBool(true),
		Foreground: pString(tcell.ColorYellow.String()),
	}

	actual := Style{
		Bold:       pBool(false),
		Foreground: pString(tcell.ColorYellow.String()),
	}.Merge(Style{
		Bold: pBool(true),
	})

	assert.Equal(t, expected.String(), actual.String())
}
