package ui

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"finance-planner/lib"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const (
	Spacer = 10
)

const (
	ColumnDate                = "Date"
	ColumnBalance             = "Balance"
	ColumnCumulativeIncome    = "CumulativeIncome"
	ColumnCumulativeExpenses  = "CumulativeExpenses"
	ColumnDayExpenses         = "DayExpenses"
	ColumnDayIncome           = "DayIncome"
	ColumnDayNet              = "DayNet"
	ColumnDiffFromStart       = "DiffFromStart"
	ColumnDayTransactionNames = "DayTransactionNames"
)

const (
	ColumnDateIndex = iota
	ColumnBalanceIndex
	ColumnCumulativeIncomeIndex
	ColumnCumulativeExpensesIndex
	ColumnDayExpensesIndex
	ColumnDayIncomeIndex
	ColumnDayNetIndex
	ColumnDiffFromStartIndex
	ColumnDayTransactionNamesIndex
)

var columns = []string{
	ColumnDate,
	ColumnBalance,
	ColumnCumulativeIncome,
	ColumnCumulativeExpenses,
	ColumnDayExpenses,
	ColumnDayIncome,
	ColumnDayNet,
	ColumnDiffFromStart,
	ColumnDayTransactionNames,
}

// make columnsIndexes the same length as the "columns" variable
var columnsIndexes = []int{
	ColumnDateIndex,
	ColumnBalanceIndex,
	ColumnCumulativeIncomeIndex,
	ColumnCumulativeExpensesIndex,
	ColumnDayExpensesIndex,
	ColumnDayIncomeIndex,
	ColumnDayNetIndex,
	ColumnDiffFromStartIndex,
	ColumnDayTransactionNamesIndex,
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

// tree view code

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

func createCheckboxColumn(title string, id int, radio bool, listStore *gtk.ListStore, txs *[]lib.TX, updateResults func()) (tvc *gtk.TreeViewColumn, err error) {
	cellRenderer, err := gtk.CellRendererToggleNew()
	if err != nil {
		return tvc, fmt.Errorf("unable to create checkbox column renderer: %v", err.Error())
	}
	// err = cellRenderer.Set("xalign", 0.0)
	// if err != nil {
	// 	return tvc, fmt.Errorf("failed to set xalign prop for checkbox column: %v", err.Error())
	// }
	cellRenderer.SetActive(true)
	cellRenderer.SetRadio(radio)
	cellRenderer.SetActivatable(true)
	cellRenderer.SetVisible(true)
	// cellRenderer.SetSentitive(true)
	cellRenderer.Connect("toggled", func(a *gtk.CellRendererToggle, path string) {
		// TODO: using nested structures results in a path that looks like 1:2:5, parse accordingly
		i, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			log.Printf("failed to parse path \"%v\" as an int: %v", path, err.Error())
		}

		// TODO: refactor into consts
		if lib.IsWeekday(configColumns[id]) {
			weekday := lib.WeekdayIndex[configColumns[id]]
			(*txs)[i].Weekdays = lib.ToggleDayFromWeekdays((*txs)[i].Weekdays, weekday)

			listStore.ForEach(func(model *gtk.TreeModel, searchPath *gtk.TreePath, iter *gtk.TreeIter) bool {
				if searchPath.String() == path {
					listStore.Set(
						iter,
						[]int{id},
						[]interface{}{
							doesTXHaveWeekday((*txs)[i], weekday),
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
		} else if configColumns[id] == ColumnActive {
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

	for i := range columns {
		tvc, err := (createColumn(columns[i], columnsIndexes[i]))
		if err != nil {
			return tv, ls, fmt.Errorf(
				"failed to create column %v with id %v: %v",
				columns[i],
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

// TODO: move to lib
func currencyMarkup(input int) string {
	currency := lib.FormatAsCurrency(input)
	if input == 0 {
		return fmt.Sprintf(`<i><span foreground="#CCCCCC">%v</span></i>`, currency)
	}
	if input > 0 {
		return fmt.Sprintf(`<span foreground="#c2e1b5">%v</span>`, currency)
	}
	if input < 0 {
		return fmt.Sprintf(`<span foreground="#dda49e">%v</span>`, currency)
	}

	return currency
}

var colorSequences = []string{
	"#d9e7fd",
	"#b4cffb",
	"#8eb7f9",
	"#699ff7",
	"#4387f5",
}

func markupDayTransactionNames(input []string) string {
	result := new(strings.Builder)
	if len(input) > 0 {
		result.WriteString(fmt.Sprintf("(%v) ", len(input)))
	}
	for i, name := range input {
		colorSequenceIndex := i % len(colorSequences)
		result.WriteString(fmt.Sprintf(`<u><span foreground="%v">%v</span></u>; `, colorSequences[colorSequenceIndex], name))
	}
	return result.String()
}

// Append a row to the list store for the tree view
func addRow(listStore *gtk.ListStore, result *lib.Result) error {
	// Get an iterator for a new row at the end of the list store
	iter := listStore.Append()

	rowData := []interface{}{
		lib.FormatAsDate(result.Date),
		currencyMarkup(result.Balance),
		currencyMarkup(result.CumulativeIncome),
		currencyMarkup(result.CumulativeExpenses),
		currencyMarkup(result.DayExpenses),
		currencyMarkup(result.DayIncome),
		currencyMarkup(result.DayNet),
		currencyMarkup(result.DiffFromStart),
		markupDayTransactionNames(result.DayTransactionNamesSlice),
		// result.DayTransactionNames,
	}

	// Set the contents of the list store row that the iterator represents
	err := listStore.Set(iter, columnsIndexes, rowData)
	if err != nil {
		return fmt.Errorf("unable to add row: %v", err.Error())
	}

	return nil
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
	resultsSw.SetMarginTop(Spacer)
	resultsSw.SetMarginBottom(Spacer)
	resultsSw.SetMarginStart(Spacer)
	resultsSw.SetMarginEnd(Spacer)
	return resultsSw, resultsTabLabel, nil
}
