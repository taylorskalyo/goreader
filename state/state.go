package state

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

var (
	stateDir    = xdg.StateHome
	appStateDir = filepath.Join(stateDir, "goreader")
	stateFile   = filepath.Join(appStateDir, "progress.json")
)

// Progress stores information about a book being read.
type Progress struct {
	// Title is a human-readable identifier for a book. Useful in the case
	// someone tries to manually modify their state file.
	Title string

	// Chapter represents the current chapter being read.
	Chapter int

	// Position represents the current position within a chapter.
	Position float64
}

// State represents the entire state file.
type State struct {
	// Library is a collection of reading progress states.
	Library map[string]Progress
}

func newState() State {
	return State{
		Library: map[string]Progress{},
	}
}

// LoadProgress opens the state file in $XDG_STATE_HOME and looks for the given
// book identifier. If not present, or if an error occurs, it returns an empty
// state.
func LoadProgress(id string) (Progress, error) {
	state, err := loadState()
	if err == nil {
		if rs, exists := state.Library[id]; exists {
			return rs, nil
		}
	}

	return Progress{}, err
}

// loadState will open a state file.
func loadState() (State, error) {
	state := newState()
	data, err := os.ReadFile(stateFile)
	if err == nil {
		err = json.Unmarshal(data, &state)
	}

	return state, err
}

// StoreProgress saves the identifier of the current book (as a key) and the
// current progress (as a value) to a state file in $XDG_STATE_HOME, either
// creating or updating it.
func StoreProgress(id string, rs Progress) error {
	if err := os.MkdirAll(appStateDir, 0700); err != nil {
		return err
	}

	state, err := loadState()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	state.Library[id] = rs

	data, err := json.MarshalIndent(state, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(stateFile, data, 0644)
}
