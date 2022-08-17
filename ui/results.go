package ui

import (
	"fmt"
	"log"

	c "finance-planner/constants"
	"finance-planner/lib"
	"finance-planner/state"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// GetNewResultsListStore creates a list store. This is what holds the data
// that will be shown on our results tree view
func GetNewResultsListStore() (ls *gtk.ListStore, err error) {
	ls, err = gtk.ListStoreNew(
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
		return ls, fmt.Errorf("unable to create results list store: %v", err.Error())
	}

	return
}

// Creates a tree view and the list store that holds its data
func setupTreeView(ls *gtk.ListStore) (tv *gtk.TreeView, err error) {
	treeView, err := gtk.TreeViewNew()
	if err != nil {
		return tv, fmt.Errorf("unable to create tree view: %v", err.Error())
	}

	tvs, _ := treeView.GetSelection()
	tvs.SetMode(gtk.SELECTION_MULTIPLE)

	for i := range c.ResultsColumns {
		tvc, err := (createColumn(c.ResultsColumns[i], c.ResultsColumnsIndexes[i]))
		if err != nil {
			return tv, fmt.Errorf(
				"failed to create column %v with id %v: %v",
				c.ResultsColumns[i],
				i,
				err.Error(),
			)
		}
		treeView.AppendColumn(tvc)
	}

	treeView.SetModel(ls)

	return treeView, nil
}

// Append a row to the list store for the tree view
func addRow(listStore *gtk.ListStore, result *lib.Result) error {
	// get an iterator for a new row at the end of the list store
	iter := listStore.Append()

	rowData := []interface{}{
		lib.FormatAsDate(result.Date),
		lib.FormatAsCurrency(result.Balance),              // lib.CurrencyMarkup(result.Balance),
		lib.FormatAsCurrency(result.CumulativeIncome),     // lib.CurrencyMarkup(result.CumulativeIncome),
		lib.FormatAsCurrency(result.CumulativeExpenses),   // lib.CurrencyMarkup(result.CumulativeExpenses),
		lib.FormatAsCurrency(result.DayExpenses),          // lib.CurrencyMarkup(result.DayExpenses),
		lib.FormatAsCurrency(result.DayIncome),            // lib.CurrencyMarkup(result.DayIncome),
		lib.FormatAsCurrency(result.DayNet),               // lib.CurrencyMarkup(result.DayNet),
		lib.FormatAsCurrency(result.DiffFromStart),        // lib.CurrencyMarkup(result.DiffFromStart),
		lib.GetCSVString(result.DayTransactionNamesSlice), // lib.MarkupColorSequence(result.DayTransactionNamesSlice),
	}

	// Set the contents of the list store row that the iterator represents
	err := listStore.Set(iter, c.ResultsColumnsIndexes, rowData)
	if err != nil {
		return fmt.Errorf("unable to add row: %v", err.Error())
	}

	return nil
}

// SyncResultsListStore clears and populates the target list store with all
// of the provided results
func SyncResultsListStore(results *[]lib.Result, ls *gtk.ListStore) error {
	if ls == nil {
		return fmt.Errorf("results list store cannot sync; is nil")
	}
	ls.Clear()

	// add rows to the tree's list store
	for _, result := range *results {
		err := addRow(ls, &result)
		if err != nil {
			return fmt.Errorf("failed to add row: %v", err.Error())
		}
	}

	return nil
}

// https://github.com/gotk3/gotk3-examples/blob/master/gtk-examples/treeview/treeview.go
func GetResultsAsTreeView(results *[]lib.Result, ls *gtk.ListStore) (tv *gtk.TreeView, err error) {
	tv, err = setupTreeView(ls)
	if err != nil {
		return tv, fmt.Errorf("failed to set up tree view: %v", err.Error())
	}

	err = SyncResultsListStore(results, ls)
	if err != nil {
		return tv, fmt.Errorf("failed to set up results tree view: %v", err.Error())
	}

	tv.SetRubberBanding(true)

	return
}

func GenerateResultsTab(txs *[]lib.TX, results []lib.Result, ls *gtk.ListStore) (grid *gtk.Grid, tabLabel *gtk.Label, err error) {
	// build the results tab page
	resultsTreeView, err := GetResultsAsTreeView(&results, ls)
	if err != nil {
		return grid, tabLabel, fmt.Errorf("failed to get results as tree view: %v", err.Error())
	}
	resultsTabLabel, err := gtk.LabelNew("Results")
	if err != nil {
		return grid, tabLabel, fmt.Errorf("failed to set tab label: %v", err.Error())
	}
	// TODO: support select capabilities later
	// resultsTreeSelection, err := resultsTreeView.GetSelection()
	// if err != nil {
	// 	return sw, tabLabel, fmt.Errorf("failed to get results tree vie sel: %v", err.Error())
	// }
	// resultsTreeSelection.Connect(c.GtkSignalChanged, SelectionChanged)
	resultsSw, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return grid, tabLabel, fmt.Errorf("Unable to create scrolled window: %v", err.Error())
	}
	resultsSw.Add(resultsTreeView)
	resultsSw.SetHExpand(true)
	resultsSw.SetVExpand(true)

	resultsGrid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("failed to create results grid", err)
	}

	resultsGrid.SetOrientation(gtk.ORIENTATION_VERTICAL)
	resultsGrid.Attach(resultsSw, 0, 0, c.FullGridWidth, 2)

	// TODO: mess with these more; it's preferable to have the tree view
	// a little more tight against the margins, but may be preferable in
	// some dimensions
	// resultsSw.SetMarginTop(c.UISpacer)
	// resultsSw.SetMarginBottom(c.UISpacer)
	// resultsSw.SetMarginStart(c.UISpacer)
	// resultsSw.SetMarginEnd(c.UISpacer)

	return resultsGrid, resultsTabLabel, nil
}

func GetResultsInputs(ws *state.WinState) (*gtk.Entry, *gtk.Entry, *gtk.Entry) {
	startingBalanceInput, err := gtk.EntryNew()
	if err != nil {
		log.Fatal("failed to create starting balance input entry:", err)
	}

	stDateInput, err := gtk.EntryNew()
	if err != nil {
		log.Fatal("failed to create start date input entry:", err)
	}

	endDateInput, err := gtk.EntryNew()
	if err != nil {
		log.Fatal("failed to create end date input entry:", err)
	}

	stDateInputUpdate := func(e *gtk.Entry) {
		s, _ := e.GetText()
		y, m, d := lib.ParseYearMonthDateString(s)
		if s == "" || (y == 0 && m == 0 && d == 0) {
			e.SetText("")
			(*ws.ShowMessageDialog)(c.MsgInvalidDateInput, gtk.MESSAGE_ERROR)
			return
		}
		ws.StartDate = fmt.Sprintf("%v-%v-%v", y, m, d)
		e.SetText(ws.StartDate)
		UpdateResults(ws, true)
	}

	endDateInputUpdate := func(e *gtk.Entry) {
		s, _ := e.GetText()
		y, m, d := lib.ParseYearMonthDateString(s)
		if s == "" || (y == 0 && m == 0 && d == 0) {
			e.SetText("")
			(*ws.ShowMessageDialog)(c.MsgInvalidDateInput, gtk.MESSAGE_ERROR)
			return
		}
		ws.EndDate = fmt.Sprintf("%v-%v-%v", y, m, d)
		e.SetText(ws.EndDate)
		UpdateResults(ws, true)
	}

	updateStartingBalance := func(e *gtk.Entry) {
		s, _ := e.GetText()
		if s == "" {
			(*ws.ShowMessageDialog)(c.MsgStartBalanceCannotBeEmpty, gtk.MESSAGE_ERROR)
			return
		}

		ws.StartingBalance = int(lib.ParseDollarAmount(s, true))

		e.SetText(
			lib.FormatAsCurrency(
				ws.StartingBalance,
			),
		)
		UpdateResults(ws, true)
	}

	startingBalanceInput.SetPlaceholderText(c.BalanceInputPlaceholderText)
	stDateInput.SetPlaceholderText(ws.StartDate)
	endDateInput.SetPlaceholderText(ws.EndDate)

	startingBalanceInput.Connect(c.GtkSignalActivate, updateStartingBalance)
	stDateInput.Connect(c.GtkSignalActivate, stDateInputUpdate)
	stDateInput.Connect(c.GtkSignalFocusOut, stDateInputUpdate)
	endDateInput.Connect(c.GtkSignalActivate, endDateInputUpdate)
	endDateInput.Connect(c.GtkSignalFocusOut, endDateInputUpdate)

	SetSpacerMarginsGtkEntry(startingBalanceInput)
	SetSpacerMarginsGtkEntry(stDateInput)
	SetSpacerMarginsGtkEntry(endDateInput)

	return startingBalanceInput, stDateInput, endDateInput
}
