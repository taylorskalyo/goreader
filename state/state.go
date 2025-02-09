package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"golang.org/x/mod/semver"
)

const (
	ver1 = "1.0.0"

	verLatest = ver1
)

var (
	//
	ErrMigration = errors.New("reading: cannot migrate reading state")

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

type BaseState struct {
	// Version is the version of the state schema.
	Version string
}

type State struct {
	BaseState

	// Library is a collection of reading states.
	Library map[string]Progress
}

func newState() State {
	s := State{}
	s.Library = make(map[string]Progress)

	return s
}

// LoadProgress opens the statefile in $XDG_STATE_HOME and looks for the gievn
// book identifier. If not present, or if an error occurs, it returns an empty
// state.
func LoadProgress(id string) Progress {
	state, err := loadState()
	if err == nil {
		if rs, exists := state.Library[id]; exists {
			return rs
		}
	}

	if !os.IsNotExist(err) {
		// TODO: log warning
	}

	return Progress{}
}

// loadState will open a statefile. It will attempt to migrate old versions to
// the latest statefile schema.
func loadState() (State, error) {
	base := BaseState{}
	state := newState()
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return state, err
	}

	if err = json.Unmarshal(data, &base); err != nil {
		return state, err
	}

	if semver.Major(base.Version) == semver.Major(verLatest) {
		err = json.Unmarshal(data, &state)
	} else {
		err = ErrMigration
	}

	return state, err
}

// StoreProgress saves the identifier of the current book (as a key) and the
// current progress (as a value) to a .json file in $XDG_STATE_HOME, either
// creating or updating it.
func StoreProgress(id string, rs Progress) error {
	if err := os.MkdirAll(appStateDir, 0700); err != nil {
		return err
	}

	state, err := loadState()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	state.Version = verLatest
	state.Library[id] = rs

	data, err := json.MarshalIndent(state, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(stateFile, data, 0644)
}
