package ui

import (
	"fmt"

	c "finance-planner/constants"

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
