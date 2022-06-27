package ui

import (
	"fmt"
	"log"

	c "finance-planner/constants"
	"finance-planner/lib"
	"finance-planner/state"

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
	entry.SetMarginTop(c.UISpacer)
	entry.SetMarginBottom(c.UISpacer)
	entry.SetMarginStart(c.UISpacer)
	entry.SetMarginEnd(c.UISpacer)
}

// SetSpacerMarginsGtkCheckBtn sets the standard spacer for all 4 different
// margin dimensions.
func SetSpacerMarginsGtkCheckBtn(entry *gtk.CheckButton) {
	entry.SetMarginTop(c.UISpacer)
	entry.SetMarginBottom(c.UISpacer)
	entry.SetMarginStart(c.UISpacer)
	entry.SetMarginEnd(c.UISpacer)
}

// SetSpacerMarginsGtkCheckBtn sets the standard spacer for all 4 different
// margin dimensions.
func SetSpacerMarginsGtkBtn(entry *gtk.Button) {
	entry.SetMarginTop(c.UISpacer)
	entry.SetMarginBottom(c.UISpacer)
	entry.SetMarginStart(c.UISpacer)
	entry.SetMarginEnd(c.UISpacer)
}

// UpdateResults gets called whenever a change is made in the config and it
// needs to be reflected in the results page. switchTo will change the currently
// shown tab.
func UpdateResults(ws *state.WinState, switchTo bool) {
	var err error
	*ws.Results, err = lib.GenerateResultsFromDateStrings(
		ws.TX,
		ws.StartingBalance,
		ws.StartDate,
		ws.EndDate,
	)
	if err != nil {
		log.Fatal("failed to generate results from date strings", err.Error())
	}

	ws.Header.SetSubtitle(fmt.Sprintf("%v*", ws.OpenFileName))

	// TODO: clear and re-populate the list store instead of removing the
	// entire tab page

	// nb.RemovePage(c.TAB_RESULTS)
	// newSw, newLabel, ws.ResultsListStore, err := ui.GenerateResultsTab(&userTX, *ws.Results)
	// if err != nil {
	// 	log.Fatalf("failed to generate results tab: %v", err.Error())
	// }
	// nb.InsertPage(newSw, newLabel, c.TAB_RESULTS)

	if ws.ResultsListStore != nil {
		err = SyncResultsListStore(ws.Results, ws.ResultsListStore)
		if err != nil {
			log.Print("failed to sync results list store:", err.Error())
		}
		ws.Win.ShowAll()
		if switchTo {
			ws.Notebook.SetCurrentPage(c.TAB_RESULTS)
		}
	}
}
