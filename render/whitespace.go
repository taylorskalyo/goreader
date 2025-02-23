package render

import (
	"regexp"
	"strings"
	"unicode"
)

type whitespaceFn func(string) string

// processWhitespace collapses whitepsace within text.
//
// https://www.w3.org/TR/CSS22/text.html#white-space-model
func processWhitespace(text string) string {
	for _, fn := range []whitespaceFn{
		wsRemoveSurroundLF,
		wsTransformLF,
		wsTransformTab,
		wsTransformSpace,
	} {
		text = fn(text)
	}

	return text
}

var (
	reRemoveSurroundLF = regexp.MustCompile("(?m)[\t\r ]*\n[\t\r ]*")
	reTransformSpace   = regexp.MustCompile(" +")
)

// wsRemoveSurroundLF collapses whitepsace within text near a linefeed (LF).
//
// Each tab (U+0009), carriage return (U+000D), or space (U+0020) character
// surrounding a linefeed (U+000A) character is removed if 'white-space' is set
// to 'normal', 'nowrap', or 'pre-line'.
func wsRemoveSurroundLF(text string) string {
	return reRemoveSurroundLF.ReplaceAllString(text, "\n")
}

// wsTransformLF collapses linefeed (LF) characters within text.
//
// If 'white-space' is set to 'normal' or 'nowrap', linefeed characters are
// transformed for rendering purpose into one of the following characters: a
// space character, a zero width space character (U+200B), or no character
// (i.e., not rendered), according to UA-specific algorithms based on the
// content script.
func wsTransformLF(text string) string {
	return strings.ReplaceAll(text, "\n", " ")
}

// wsTransformTab collapses tab characters within text.
//
// Every tab (U+0009) is converted to a space (U+0020)
func wsTransformTab(text string) string {
	return strings.ReplaceAll(text, "\t", " ")
}

// wsTransformSpace collapses space characters within text.
//
// Any space (U+0020) following another space (U+0020) — even a space before
// the inline, if that space also has 'white-space' set to 'normal', 'nowrap'
// or 'pre-line' — is removed.
func wsTransformSpace(text string) string {
	return reTransformSpace.ReplaceAllString(text, " ")
}

func hasText(text string) bool {
	if len(text) > 0 {
		for _, r := range text {
			if !unicode.IsSpace(r) {
				return true
			}
		}
	}

	return false
}
