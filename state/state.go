package state

import (
	"time"

	"finance-planner/lib"

	"github.com/gotk3/gotk3/gtk"
)

// contains the user's configuration on a per-window basis - for example,
// HideInactive is a value that differs on a per-window basis. Other state
// values, such as the currently sorted column, should also be stored here.
// Additionally, utility functions such as ones that allow an error dialog
// to be shown from anywhere, should be stored here as pointers.
type WinState struct {
	HideInactive        bool
	CurrentColumnSort   string
	OpenFileName        string
	StartDate           string
	StartDateTime       time.Time
	EndDate             string
	EndDateTime         time.Time
	SelectedConfigItems []int
	ShowMessageDialog   *func(message string)
	ConfigListStore     *gtk.ListStore
	ResultsListStore    *gtk.ListStore
	UpdateResults       *func()   // function that gets called upon editing things
	TX                  *[]lib.TX // transaction definitions for the current window
	Results             *[]lib.Result
	App                 *gtk.Application
	Win                 *gtk.ApplicationWindow
}