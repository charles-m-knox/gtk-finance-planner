package ui

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"
)

// Add a column to the tree view (during the initialization of the tree view)
func createColumn(title string, id int) (tvc *gtk.TreeViewColumn, err error) {
	cellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tvc, fmt.Errorf("unable to create text cell renderer: %v", err.Error())
	}

	column, err := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "markup", id)
	// column, err := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "text", id)
	if err != nil {
		return tvc, fmt.Errorf("unable to create cell column: %v", err.Error())
	}
	column.SetResizable(true)
	column.SetClickable(true)

	return column, nil
}
