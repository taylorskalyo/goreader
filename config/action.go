package config

import "fmt"

const (
	ActionExit Action = iota
	ActionDown
	ActionUp
	ActionLeft
	ActionRight
	ActionTop
	ActionBottom
	ActionForward
	ActionBackward
	ActionChapterNext
	ActionChapterPrevious
)

var (
	// ActionNames holds the written names of callable events. Useful to echo back
	// an event name, or to look up an event from a string value.
	ActionNames = map[Action]string{
		ActionExit:            "Exit",
		ActionDown:            "Down",
		ActionUp:              "Up",
		ActionLeft:            "Left",
		ActionRight:           "Right",
		ActionTop:             "Top",
		ActionBottom:          "Botom",
		ActionForward:         "Forward",
		ActionBackward:        "Backward",
		ActionChapterNext:     "ChapterNext",
		ActionChapterPrevious: "ChapterPrevious",
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
