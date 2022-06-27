package ui

import (
	"fmt"
	"log"
	"strconv"

	c "finance-planner/constants"
	"finance-planner/lib"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func createCheckboxColumn(title string, id int, radio bool, listStore *gtk.ListStore, txs *[]lib.TX, updateResults func()) (tvc *gtk.TreeViewColumn, err error) {
	cellRenderer, err := gtk.CellRendererToggleNew()
	if err != nil {
		return tvc, fmt.Errorf("unable to create checkbox column renderer: %v", err.Error())
	}
	cellRenderer.SetActive(true)
	cellRenderer.SetRadio(radio)
	cellRenderer.SetActivatable(true)
	cellRenderer.SetVisible(true)
	cellRenderer.Connect("toggled", func(a *gtk.CellRendererToggle, path string) {
		// TODO: using nested structures results in a path that looks
		// like 1:2:5 -  parse accordingly
		i, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			log.Printf("failed to parse path \"%v\" as an int: %v", path, err.Error())
		}

		// TODO: refactor into consts
		if lib.IsWeekday(c.ConfigColumns[id]) {
			weekday := lib.WeekdayIndex[c.ConfigColumns[id]]
			(*txs)[i].Weekdays = lib.ToggleDayFromWeekdays((*txs)[i].Weekdays, weekday)

			listStore.ForEach(func(model *gtk.TreeModel, searchPath *gtk.TreePath, iter *gtk.TreeIter) bool {
				if searchPath.String() == path {
					listStore.Set(
						iter,
						[]int{id},
						[]interface{}{
							(*txs)[i].DoesTXHaveWeekday(weekday),
						})
					return true
				}
				return false
			})
			updateResults()
			err := SyncListStore(txs, listStore)
			if err != nil {
				log.Printf("failed to sync list store: %v", err.Error())
			}
		} else if c.ConfigColumns[id] == c.ColumnActive {
			(*txs)[i].Active = !(*txs)[i].Active
			listStore.ForEach(func(model *gtk.TreeModel, searchPath *gtk.TreePath, iter *gtk.TreeIter) bool {
				if searchPath.String() == path {
					listStore.Set(
						iter,
						[]int{id},
						[]interface{}{(*txs)[i].Active})
					return true
				}
				return false
			})
			updateResults()
			err := SyncListStore(txs, listStore)
			if err != nil {
				log.Printf("failed to sync list store: %v", err.Error())
			}
		}
	})

	column, err := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "active", id)
	// column, err := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "active", id)
	if err != nil {
		return tvc, fmt.Errorf("unable to create checkbox cell column: %v", err.Error())
	}
	column.SetResizable(true)
	column.SetClickable(true)
	column.SetVisible(true)

	return column, nil
}

// Creates a tree view and the list store that holds its data
func setupTreeView() (tv *gtk.TreeView, ls *gtk.ListStore, err error) {
	treeView, err := gtk.TreeViewNew()
	if err != nil {
		return tv, ls, fmt.Errorf("unable to create tree view: %v", err.Error())
	}

	tvs, _ := treeView.GetSelection()
	tvs.SetMode(gtk.SELECTION_MULTIPLE)

	for i := range c.ResultsColumns {
		tvc, err := (createColumn(c.ResultsColumns[i], c.ResultsColumnsIndexes[i]))
		if err != nil {
			return tv, ls, fmt.Errorf(
				"failed to create column %v with id %v: %v",
				c.ResultsColumns[i],
				i,
				err.Error(),
			)
		}
		treeView.AppendColumn(tvc)
	}

	// create a list store. This is what holds the data that will be
	// shown on our tree view
	listStore, err := gtk.ListStoreNew(
		glib.TYPE_STRING,
		glib.TYPE_STRING,
		glib.TYPE_STRING,
		glib.TYPE_STRING,
		glib.TYPE_STRING,
		glib.TYPE_STRING,
		glib.TYPE_STRING,
		glib.TYPE_STRING,
		glib.TYPE_STRING,
	)
	if err != nil {
		return tv, ls, fmt.Errorf("unable to create list store: %v", err.Error())
	}
	treeView.SetModel(listStore)

	return treeView, listStore, nil
}

// Append a row to the list store for the tree view
func addRow(listStore *gtk.ListStore, result *lib.Result) error {
	// get an iterator for a new row at the end of the list store
	iter := listStore.Append()

	rowData := []interface{}{
		lib.FormatAsDate(result.Date),
		lib.CurrencyMarkup(result.Balance),
		lib.CurrencyMarkup(result.CumulativeIncome),
		lib.CurrencyMarkup(result.CumulativeExpenses),
		lib.CurrencyMarkup(result.DayExpenses),
		lib.CurrencyMarkup(result.DayIncome),
		lib.CurrencyMarkup(result.DayNet),
		lib.CurrencyMarkup(result.DiffFromStart),
		lib.MarkupColorSequence(result.DayTransactionNamesSlice),
	}

	// Set the contents of the list store row that the iterator represents
	err := listStore.Set(iter, c.ResultsColumnsIndexes, rowData)
	if err != nil {
		return fmt.Errorf("unable to add row: %v", err.Error())
	}

	return nil
}

// https://github.com/gotk3/gotk3-examples/blob/master/gtk-examples/treeview/treeview.go
func GetResultsAsTreeView(results *[]lib.Result) (tv *gtk.TreeView, ls *gtk.ListStore, err error) {
	treeView, listStore, err := setupTreeView()
	if err != nil {
		return tv, ls, fmt.Errorf("failed to set up tree view: %v", err.Error())
	}

	// add rows to the tree's list store
	for _, result := range *results {
		err := addRow(listStore, &result)
		if err != nil {
			return tv, ls, fmt.Errorf("failed to add row: %v", err.Error())
		}
	}

	treeView.SetRubberBanding(true)

	return treeView, listStore, nil
}

func GenerateResultsTab(txs *[]lib.TX, results []lib.Result) (sw *gtk.ScrolledWindow, tabLabel *gtk.Label, err error) {
	// build the results tab page
	resultsTreeView, _, err := GetResultsAsTreeView(&results)
	if err != nil {
		return sw, tabLabel, fmt.Errorf("failed to get results as tree view: %v", err.Error())
	}
	resultsTabLabel, err := gtk.LabelNew("Results")
	if err != nil {
		return sw, tabLabel, fmt.Errorf("failed to set tab label: %v", err.Error())
	}
	// TODO: support select capabilities later
	// resultsTreeSelection, err := resultsTreeView.GetSelection()
	// if err != nil {
	// 	return sw, tabLabel, fmt.Errorf("failed to get results tree vie sel: %v", err.Error())
	// }
	// resultsTreeSelection.Connect("changed", SelectionChanged)
	resultsSw, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return sw, tabLabel, fmt.Errorf("Unable to create scrolled window: %v", err.Error())
	}
	resultsSw.Add(resultsTreeView)
	resultsSw.SetHExpand(true)
	resultsSw.SetVExpand(true)
	resultsSw.SetMarginTop(c.UISpacer)
	resultsSw.SetMarginBottom(c.UISpacer)
	resultsSw.SetMarginStart(c.UISpacer)
	resultsSw.SetMarginEnd(c.UISpacer)
	return resultsSw, resultsTabLabel, nil
}
