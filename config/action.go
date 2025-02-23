package config

import "fmt"

const (
	ActionExit Action = iota
	ActionUp
	ActionDown
	ActionLeft
	ActionRight
	ActionTop
	ActionBottom
	ActionBackward
	ActionForward
	ActionChapterPrevious
	ActionChapterNext
)

var (
	// ActionNames holds the written names of callable events. Useful to echo back
	// an event name, or to look up an event from a string value.
	ActionNames = map[Action]string{
		ActionUp:              "Up",
		ActionDown:            "Down",
		ActionLeft:            "Left",
		ActionRight:           "Right",
		ActionTop:             "Top",
		ActionBottom:          "Botom",
		ActionBackward:        "Backward",
		ActionForward:         "Forward",
		ActionChapterPrevious: "ChapterPrevious",
		ActionChapterNext:     "ChapterNext",
		ActionExit:            "Exit",
	}

	namedActions = map[string]Action{}
)

func init() {
	// Get mapping of string -> Action.
	namedActions = make(map[string]Action, len(ActionNames))
	for k, v := range ActionNames {
		namedActions[v] = k
	}
}

// Action is an action that can be bound to a sequence of key presses.
type Action int16

// UnmarshalText creates a new Action from text.
func (e *Action) UnmarshalText(text []byte) error {
	var ok bool

	if *e, ok = namedActions[string(text)]; !ok {
		return fmt.Errorf("config: unrecognized event \"%s\"", text)
	}

	return nil
}

// MarshalText renders an Action as text.
func (e Action) MarshalText() ([]byte, error) {
	return []byte(ActionNames[e]), nil
}
