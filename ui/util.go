package ui

import (
	"fmt"
	"log"
	"time"

	lib "github.com/charles-m-knox/finance-planner-lib"
	"github.com/charles-m-knox/gtk-finance-planner/constants"
	"github.com/charles-m-knox/gtk-finance-planner/state"

	"github.com/gotk3/gotk3/gtk"
)

// Add a column to the tree view (during the initialization of the tree view)
func createColumn(title string, id int) (tvc *gtk.TreeViewColumn, err error) {
	cellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tvc, fmt.Errorf("unable to create text cell renderer: %v", err.Error())
	}

	column, err := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "markup", id)
	if err != nil {
		return tvc, fmt.Errorf("unable to create cell column: %v", err.Error())
	}
	column.SetResizable(true)
	column.SetClickable(true)

	return column, nil
}

// SetSpacerMarginsGtkEntry sets the standard spacer for all 4 different margin
// dimensions. It is a simple helper function that reduces code duplication.
// Might be nice to implement this with Go generics, but for now, one has to
// exist for each gtk widget type.
func SetSpacerMarginsGtkEntry(entry *gtk.Entry) {
	entry.SetMarginTop(constants.UISpacer)
	entry.SetMarginBottom(constants.UISpacer)
	entry.SetMarginStart(constants.UISpacer)
	entry.SetMarginEnd(constants.UISpacer)
}

// SetSpacerMarginsGtkCheckBtn sets the standard spacer for all 4 different
// margin dimensions.
func SetSpacerMarginsGtkCheckBtn(chk *gtk.CheckButton) {
	chk.SetMarginTop(constants.UISpacer)
	chk.SetMarginBottom(constants.UISpacer)
	chk.SetMarginStart(constants.UISpacer)
	chk.SetMarginEnd(constants.UISpacer)
}

// SetSpacerMarginsGtkCheckBtn sets the standard spacer for all 4 different
// margin dimensions.
func SetSpacerMarginsGtkBtn(btn *gtk.Button) {
	btn.SetMarginTop(constants.UISpacer)
	btn.SetMarginBottom(constants.UISpacer)
	btn.SetMarginStart(constants.UISpacer)
	btn.SetMarginEnd(constants.UISpacer)
}

// UpdateResults gets called whenever a change is made in the config and it
// needs to be reflected in the results page. switchTo will change the currently
// shown tab.
func UpdateResults(ws *state.WinState, switchTo bool) {
	var err error

	now := time.Now()

	*ws.Results, err = lib.GetResults(
		*ws.TX,
		lib.GetDateFromStrSafe(ws.StartDate, now),
		lib.GetDateFromStrSafe(ws.EndDate, now),
		ws.StartingBalance,
		func(_ string) {},
	)
	if err != nil {
		log.Fatal("failed to generate results from date strings", err.Error())
	}

	ws.Header.SetSubtitle(fmt.Sprintf("%v*", ws.OpenFileName))

	if ws.ResultsListStore != nil {
		err = SyncResultsListStore(ws.Results, ws.ResultsListStore)
		if err != nil {
			log.Print("failed to sync results list store:", err.Error())
		}
		ws.Win.ShowAll()
		if switchTo {
			ws.Notebook.SetCurrentPage(constants.TAB_RESULTS)
		}
	}
}
