package ui

import (
	"finance-planner/lib"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// TODO: refactor into consts
var configColumns = []string{
	"Amount",    // int in cents; 500 = $5.00
	"Active",    // bool true/false
	"Name",      // editable string
	"Frequency", // dropdown, monthly/daily/weekly/yearly
	"Interval",  // integer, occurs every x frequency
	"Monday",    // bool
	"Tuesday",   // bool
	"Wednesday", // bool
	"Thursday",  // bool
	"Friday",    // bool
	"Saturday",  // bool
	"Sunday",    // bool
	"Starts",    // string
	// "StartsDay",   // int
	// "StartsMonth", // int
	// "StartsYear",  // int
	"Ends", // string
	// "EndsDay",     // int
	// "EndsMonth",   // int
	// "EndsYear",    // int
	// "RRule",       // rendered rrule
	"Note", // editable string
}

const (
	COLUMN_AMOUNT    = iota // int in cents; 500 = $5.00
	COLUMN_ACTIVE           // bool true/false
	COLUMN_NAME             // editable string
	COLUMN_FREQUENCY        // dropdown, monthly/daily/weekly/yearly
	COLUMN_INTERVAL         // integer, occurs every x frequency
	COLUMN_MONDAY           // bool
	COLUMN_TUESDAY          // bool
	COLUMN_WEDNESDAY        // bool
	COLUMN_THURSDAY         // bool
	COLUMN_FRIDAY           // bool
	COLUMN_SATURDAY         // bool
	COLUMN_SUNDAY           // bool
	COLUMN_STARTS           // string
	COLUMN_ENDS             // string
	COLUMN_NOTE             // editable string
)

var configColumnTypes = []glib.Type{
	glib.TYPE_STRING,
	glib.TYPE_BOOLEAN,
	glib.TYPE_STRING,
	glib.TYPE_STRING,
	glib.TYPE_STRING,
	glib.TYPE_BOOLEAN,
	glib.TYPE_BOOLEAN,
	glib.TYPE_BOOLEAN,
	glib.TYPE_BOOLEAN,
	glib.TYPE_BOOLEAN,
	glib.TYPE_BOOLEAN,
	glib.TYPE_BOOLEAN,
	glib.TYPE_STRING,
	glib.TYPE_STRING,
	// glib.TYPE_INT,
	// glib.TYPE_INT,
	// glib.TYPE_INT,
	// glib.TYPE_INT,
	// glib.TYPE_INT,
	// glib.TYPE_INT,
	// glib.TYPE_STRING,
	glib.TYPE_STRING,
}

// ui functions to build out a config tree view

func doesTXHaveWeekday(tx lib.TX, weekday int) bool {
	for _, d := range tx.Weekdays {
		if weekday == d {
			return true
		}
	}
	return false
}

func addConfigTreeRow(listStore *gtk.ListStore, tx *lib.TX) error {
	// Get an iterator for a new row at the end of the list store
	iter := listStore.Append()

	rowData := []interface{}{
		lib.FormatAsCurrency(tx.Amount),
		tx.Active,
		tx.Name,
		tx.Frequency,
		fmt.Sprintf("%v", tx.Interval),
		doesTXHaveWeekday(*tx, 0),
		doesTXHaveWeekday(*tx, 1),
		doesTXHaveWeekday(*tx, 2),
		doesTXHaveWeekday(*tx, 3),
		doesTXHaveWeekday(*tx, 4),
		doesTXHaveWeekday(*tx, 5),
		doesTXHaveWeekday(*tx, 6),
		fmt.Sprintf("%v-%v-%v", tx.StartsYear, tx.StartsMonth, tx.StartsDay),
		fmt.Sprintf("%v-%v-%v", tx.EndsYear, tx.EndsMonth, tx.EndsDay),
		// tx.StartsDay,
		// tx.StartsMonth,
		// tx.StartsYear,
		// tx.EndsDay,
		// tx.EndsMonth,
		// tx.EndsYear,
		// tx.RRule,
		tx.Note,
	}

	columnsIndexes := []int{}
	for i := range rowData {
		columnsIndexes = append(columnsIndexes, i)
	}

	// Set the contents of the list store row that the iterator represents
	err := listStore.Set(iter, columnsIndexes, rowData)

	if err != nil {
		return fmt.Errorf("unable to add config tree row: %v", err.Error())
	}

	return nil
}

// Creates a tree view and the list store that holds its data
func setupConfigTreeView(txs *[]lib.TX, updateResults func()) (tv *gtk.TreeView, ls *gtk.ListStore, err error) {
	treeView, err := gtk.TreeViewNew()
	if err != nil {
		return tv, ls, fmt.Errorf("unable to create config tree view: %v", err.Error())
	}

	// TODO: spread this?
	// func spread(a ...interface{}) []interface{} {
	// 	return a
	// }
	// https://stackoverflow.com/a/57310617

	// create a list store. This is what holds the data that will be
	// shown on our tree view
	listStore, err := gtk.ListStoreNew(
		glib.TYPE_STRING,
		glib.TYPE_BOOLEAN,
		glib.TYPE_STRING,
		glib.TYPE_STRING,
		glib.TYPE_STRING,
		glib.TYPE_BOOLEAN,
		glib.TYPE_BOOLEAN,
		glib.TYPE_BOOLEAN,
		glib.TYPE_BOOLEAN,
		glib.TYPE_BOOLEAN,
		glib.TYPE_BOOLEAN,
		glib.TYPE_BOOLEAN,
		glib.TYPE_STRING,
		glib.TYPE_STRING,
		// glib.TYPE_INT,
		// glib.TYPE_INT,
		// glib.TYPE_INT,
		// glib.TYPE_INT,
		// glib.TYPE_INT,
		// glib.TYPE_INT,
		// glib.TYPE_STRING,
		glib.TYPE_STRING,
	)
	if err != nil {
		return tv, ls, fmt.Errorf("unable to create config list store: %v", err.Error())
	}

	// amtCellRenderer, err := gtk.CellRendererSpinNew()
	// if err != nil {
	// 	return tv, ls, fmt.Errorf("unable to create amount column renderer: %v", err.Error())
	// }
	// https://docs.gtk.org/gtk3/class.CellRendererSpin.html#properties
	// amtSpinnerRange, err := gtk.AdjustmentNew(500, -99999999999999, -99999999999999, 100, 10000, 10000)
	// if err != nil {
	// 	return tv, ls, fmt.Errorf("unable to generate amtSpinnerRange: %v", err.Error())
	// }
	// amtCellRenderer.SetProperty("adjustment", *amtSpinnerRange)
	// amtCellRenderer.SetProperty("digits", 0)

	// amount column
	amtCellEditingStarted := func(a *gtk.CellRendererText, e *gtk.CellEditable, path string) {
		log.Println("editing-started", a, path)
	}
	amtCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
		i, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			log.Printf("failed to parse path \"%v\" as an int: %v", path, err.Error())
		}
		log.Println("edited", a, path, newText)
		newValue := int(lib.ParseDollarAmount(newText, false))
		log.Println(newText, newValue)
		(*txs)[i].Amount = newValue
		formattedNewValue := lib.FormatAsCurrency(newValue)
		// push the value to the tree view's list store as well as updating the TX definition
		listStore.ForEach(func(model *gtk.TreeModel, searchPath *gtk.TreePath, iter *gtk.TreeIter) bool {
			if searchPath.String() == path {
				listStore.Set(iter, []int{COLUMN_AMOUNT}, []interface{}{formattedNewValue})
				return true
			}
			return false
		})
		updateResults()
	}
	amtCellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tv, ls, fmt.Errorf("unable to create amount column renderer: %v", err.Error())
	}
	amtCellRenderer.SetProperty("editable", true)
	amtCellRenderer.SetVisible(true)
	amtCellRenderer.Connect("editing-started", amtCellEditingStarted)
	amtCellRenderer.Connect("edited", amtCellEditingFinished)
	amtColumn, err := gtk.TreeViewColumnNewWithAttribute("Amount", amtCellRenderer, "text", 0)
	if err != nil {
		return tv, ls, fmt.Errorf("unable to create amount cell column: %v", err.Error())
	}
	amtColumn.SetResizable(true)
	amtColumn.SetClickable(true)
	amtColumn.SetVisible(true)
	treeView.AppendColumn(amtColumn)

	// active column
	activeColumn, err := createCheckboxColumn("Active", 1, false, listStore, txs, updateResults)
	if err != nil {
		return tv, ls, fmt.Errorf(
			"failed to create checkbox config column 'Active': %v", err.Error(),
		)
	}
	treeView.AppendColumn(activeColumn)

	// name column
	nameCellEditingStarted := func(a *gtk.CellRendererText, e *gtk.CellEditable, path string) {
		log.Println("editing-started", a, path)
	}
	nameCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
		i, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			log.Printf("failed to parse path \"%v\" as an int: %v", path, err.Error())
		}
		log.Println("edited", a, path, newText)
		newValue := strings.TrimSpace(newText)
		(*txs)[i].Name = newValue
		// push the value to the tree view's list store as well as updating the TX definition
		listStore.ForEach(func(model *gtk.TreeModel, searchPath *gtk.TreePath, iter *gtk.TreeIter) bool {
			if searchPath.String() == path {
				listStore.Set(iter, []int{COLUMN_NAME}, []interface{}{newValue})
				return true
			}
			return false
		})
		updateResults()
	}
	nameCellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tv, ls, fmt.Errorf("unable to create Name column renderer: %v", err.Error())
	}
	nameCellRenderer.SetProperty("editable", true)
	nameCellRenderer.SetVisible(true)
	nameCellRenderer.Connect("editing-started", nameCellEditingStarted)
	nameCellRenderer.Connect("edited", nameCellEditingFinished)
	nameColumn, err := gtk.TreeViewColumnNewWithAttribute("Name", nameCellRenderer, "text", 2)
	if err != nil {
		return tv, ls, fmt.Errorf("unable to create Name cell column: %v", err.Error())
	}
	nameColumn.SetResizable(true)
	nameColumn.SetClickable(true)
	nameColumn.SetVisible(true)
	treeView.AppendColumn(nameColumn)

	// freq column
	freqCellEditingStarted := func(a *gtk.CellRendererText, e *gtk.CellEditable, path string) {
		log.Println("editing-started", a, path)
	}
	freqCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
		i, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			log.Printf("failed to parse path \"%v\" as an int: %v", path, err.Error())
		}
		log.Println("edited", a, path, newText)
		newValue := strings.ToUpper(strings.TrimSpace(newText))
		if newValue == "Y" {
			newValue = "YEARLY"
		}
		if newValue == "W" {
			newValue = "WEEKLY"
		}
		if newValue == "M" {
			newValue = "MONTHLY"
		}
		if newValue != "WEEKLY" && newValue != "MONTHLY" && newValue != "YEARLY" {
			return
		}
		(*txs)[i].Frequency = newValue
		// push the value to the tree view's list store as well as updating the TX definition
		listStore.ForEach(func(model *gtk.TreeModel, searchPath *gtk.TreePath, iter *gtk.TreeIter) bool {
			if searchPath.String() == path {
				listStore.Set(iter, []int{COLUMN_FREQUENCY}, []interface{}{newValue})
				return true
			}
			return false
		})
		updateResults()
	}
	freqCellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tv, ls, fmt.Errorf("unable to create Frequency column renderer: %v", err.Error())
	}
	freqCellRenderer.SetProperty("editable", true)
	freqCellRenderer.SetVisible(true)
	freqCellRenderer.Connect("editing-started", freqCellEditingStarted)
	freqCellRenderer.Connect("edited", freqCellEditingFinished)
	freqColumn, err := gtk.TreeViewColumnNewWithAttribute("Frequency", freqCellRenderer, "text", 3)
	if err != nil {
		return tv, ls, fmt.Errorf("unable to create Frequency cell column: %v", err.Error())
	}
	freqColumn.SetResizable(true)
	freqColumn.SetClickable(true)
	freqColumn.SetVisible(true)
	treeView.AppendColumn(freqColumn)

	// interval column
	intervalCellEditingStarted := func(a *gtk.CellRendererText, e *gtk.CellEditable, path string) {
		log.Println("editing-started", a, path)
	}
	intervalCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
		log.Println("edited", a, path, newText)
		i, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			log.Printf("failed to parse path \"%v\" as an int: %v", path, err.Error())
		}
		log.Println("edited", a, path, newText)
		newValue, err := strconv.ParseInt(newText, 10, 64)
		if err != nil {
			log.Println("failed to parse interval integer", err.Error())
			return
		}
		log.Println("edited", a, path, newText)
		if newValue < 0 {
			return
		}
		(*txs)[i].Interval = int(newValue)
		// push the value to the tree view's list store as well as updating the TX definition
		listStore.ForEach(func(model *gtk.TreeModel, searchPath *gtk.TreePath, iter *gtk.TreeIter) bool {
			if searchPath.String() == path {
				listStore.Set(iter, []int{COLUMN_INTERVAL}, []interface{}{newValue})
				return true
			}
			return false
		})
		updateResults()
	}
	intervalCellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tv, ls, fmt.Errorf("unable to create Interval column renderer: %v", err.Error())
	}
	intervalCellRenderer.SetProperty("editable", true)
	intervalCellRenderer.SetVisible(true)
	intervalCellRenderer.Connect("editing-started", intervalCellEditingStarted)
	intervalCellRenderer.Connect("edited", intervalCellEditingFinished)
	intervalColumn, err := gtk.TreeViewColumnNewWithAttribute("Interval", intervalCellRenderer, "text", 4)
	if err != nil {
		return tv, ls, fmt.Errorf("unable to create Interval cell column: %v", err.Error())
	}
	intervalColumn.SetResizable(true)
	intervalColumn.SetClickable(true)
	intervalColumn.SetVisible(true)
	treeView.AppendColumn(intervalColumn)

	// weekday columns
	for i, weekday := range []string{
		"Monday",
		"Tuesday",
		"Wednesday",
		"Thursday",
		"Friday",
		"Saturday",
		"Sunday",
	} {
		// offset by 4 previous columns
		activeColumn, err := createCheckboxColumn(weekday, i+5, false, listStore, txs, updateResults)
		if err != nil {
			return tv, ls, fmt.Errorf(
				"failed to create checkbox config column '%v': %v", weekday, err.Error(),
			)
		}
		treeView.AppendColumn(activeColumn)
	}

	// starts column
	startsCellEditingStarted := func(a *gtk.CellRendererText, e *gtk.CellEditable, path string) {
		log.Println("editing-started", a, path)
	}
	startsCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
		i, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			log.Printf("failed to parse path \"%v\" as an int: %v", path, err.Error())
		}
		log.Println("edited", a, path, newText)
		yr, mo, day := lib.ParseYearMonthDateString(strings.TrimSpace(newText))
		if err != nil {
			log.Println("failed to parse interval integer", err.Error())
			return
		}
		(*txs)[i].StartsYear = yr
		(*txs)[i].StartsMonth = mo
		(*txs)[i].StartsDay = day
		// push the value to the tree view's list store as well as updating the TX definition
		listStore.ForEach(func(model *gtk.TreeModel, searchPath *gtk.TreePath, iter *gtk.TreeIter) bool {
			if searchPath.String() == path {
				listStore.Set(iter, []int{COLUMN_STARTS}, []interface{}{fmt.Sprintf("%v-%v-%v", yr, mo, day)})
				return true
			}
			return false
		})
		updateResults()
	}
	startsCellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tv, ls, fmt.Errorf("unable to create Starts column renderer: %v", err.Error())
	}
	startsCellRenderer.SetProperty("editable", true)
	startsCellRenderer.SetVisible(true)
	startsCellRenderer.Connect("editing-started", startsCellEditingStarted)
	startsCellRenderer.Connect("edited", startsCellEditingFinished)
	startsColumn, err := gtk.TreeViewColumnNewWithAttribute("Starts", startsCellRenderer, "text", 12)
	if err != nil {
		return tv, ls, fmt.Errorf("unable to create Starts cell column: %v", err.Error())
	}
	startsColumn.SetResizable(true)
	startsColumn.SetClickable(true)
	startsColumn.SetVisible(true)
	treeView.AppendColumn(startsColumn)

	// ends column
	endsCellEditingStarted := func(a *gtk.CellRendererText, e *gtk.CellEditable, path string) {
		log.Println("editing-started", a, path)
	}
	endsCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
		log.Println("edited", a, path, newText)
		i, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			log.Printf("failed to parse path \"%v\" as an int: %v", path, err.Error())
		}
		yr, mo, day := lib.ParseYearMonthDateString(strings.TrimSpace(newText))
		if err != nil {
			log.Println("failed to parse interval integer", err.Error())
			return
		}
		(*txs)[i].EndsYear = yr
		(*txs)[i].EndsMonth = mo
		(*txs)[i].EndsDay = day
		// push the value to the tree view's list store as well as updating the TX definition
		listStore.ForEach(func(model *gtk.TreeModel, searchPath *gtk.TreePath, iter *gtk.TreeIter) bool {
			if searchPath.String() == path {
				listStore.Set(iter, []int{COLUMN_ENDS}, []interface{}{fmt.Sprintf("%v-%v-%v", yr, mo, day)})
				return true
			}
			return false
		})
		updateResults()
	}
	endsCellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tv, ls, fmt.Errorf("unable to create Ends column renderer: %v", err.Error())
	}
	endsCellRenderer.SetProperty("editable", true)
	endsCellRenderer.SetVisible(true)
	endsCellRenderer.Connect("editing-started", endsCellEditingStarted)
	endsCellRenderer.Connect("edited", endsCellEditingFinished)
	endsColumn, err := gtk.TreeViewColumnNewWithAttribute("Ends", endsCellRenderer, "text", 13)
	if err != nil {
		return tv, ls, fmt.Errorf("unable to create Ends cell column: %v", err.Error())
	}
	endsColumn.SetResizable(true)
	endsColumn.SetClickable(true)
	endsColumn.SetVisible(true)
	treeView.AppendColumn(endsColumn)

	// notes column
	notesCellEditingStarted := func(a *gtk.CellRendererText, e *gtk.CellEditable, path string) {
		log.Println("editing-started", a, path)
	}
	notesCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
		log.Println("edited", a, path, newText)
		i, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			log.Printf("failed to parse path \"%v\" as an int: %v", path, err.Error())
		}
		(*txs)[i].Note = newText
		// push the value to the tree view's list store as well as updating the TX definition
		listStore.ForEach(func(model *gtk.TreeModel, searchPath *gtk.TreePath, iter *gtk.TreeIter) bool {
			if searchPath.String() == path {
				listStore.Set(iter, []int{COLUMN_NOTE}, []interface{}{newText})
				return true
			}
			return false
		})
		updateResults()
	}
	notesCellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tv, ls, fmt.Errorf("unable to create Notes column renderer: %v", err.Error())
	}
	notesCellRenderer.SetProperty("editable", true)
	notesCellRenderer.SetVisible(true)
	notesCellRenderer.Connect("editing-started", notesCellEditingStarted)
	notesCellRenderer.Connect("edited", notesCellEditingFinished)
	notesColumn, err := gtk.TreeViewColumnNewWithAttribute("Notes", notesCellRenderer, "text", 14)
	if err != nil {
		return tv, ls, fmt.Errorf("unable to create Notes cell column: %v", err.Error())
	}
	notesColumn.SetResizable(true)
	notesColumn.SetClickable(true)
	notesColumn.SetVisible(true)
	treeView.AppendColumn(notesColumn)

	// frequency column combo box attempt
	// freqCellRenderer, err := gtk.CellRendererComboNew()
	// if err != nil {
	// 	return tv, ls, fmt.Errorf("unable to create Frequency column renderer: %v", err.Error())
	// }
	// freqComboChoicesListStore, err := gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING)
	// if err != nil {
	// 	return tv, ls, fmt.Errorf("unable to create freq cell combo choice tree model: %v", err.Error())
	// }

	// the below section was me trying to get the editable combo box to work
	// by using one row and three columns

	// // Get an iterator for a new row at the end of the list store
	// freqComboModelIter := freqComboChoicesListStore.Append()
	// // Set the contents of the list store row that the iterator represents
	// err = freqComboChoicesListStore.Set(freqComboModelIter, []int{0, 1, 2}, []interface{}{"WEEKLY", "MONTHLY", "DAILY"})
	// if err != nil {
	// 	return tv, ls, fmt.Errorf("unable to iter freq cell combo choice tree model: %v", err.Error())
	// }
	// Get an iterator for a new row at the end of the list store

	// the below section was me trying to get the editable combo box to work
	// by using three rows

	// freqComboModelIter := freqComboChoicesListStore.Append()
	// // Set the contents of the list store row that the iterator represents
	// err = freqComboChoicesListStore.Set(freqComboModelIter, []int{1}, []interface{}{"MONTHLY"})
	// if err != nil {
	// 	return tv, ls, fmt.Errorf("unable to iter 1 freq cell combo choice tree model: %v", err.Error())
	// }
	// // Get an iterator for a new row at the end of the list store
	// freqComboModelIter = freqComboChoicesListStore.Append()
	// // Set the contents of the list store row that the iterator represents
	// err = freqComboChoicesListStore.Set(freqComboModelIter, []int{1}, []interface{}{"WEEKLY"})
	// if err != nil {
	// 	return tv, ls, fmt.Errorf("unable to iter 2 freq cell combo choice tree model: %v", err.Error())
	// }
	// // Get an iterator for a new row at the end of the list store
	// freqComboModelIter = freqComboChoicesListStore.Append()
	// // Set the contents of the list store row that the iterator represents
	// err = freqComboChoicesListStore.Set(freqComboModelIter, []int{1}, []interface{}{"DAILY"})
	// if err != nil {
	// 	return tv, ls, fmt.Errorf("unable to iter 3 freq cell combo choice tree model: %v", err.Error())
	// }
	// freqCellRenderer.SetProperty("model", freqComboChoicesListStore.TreeModel) // https://docs.gtk.org/gtk3/property.CellRendererCombo.model.html
	// freqCellRenderer.SetProperty("text-column", 1)                             // https://docs.gtk.org/gtk3/property.CellRendererCombo.text-column.html
	// freqCellRenderer.SetProperty("editable", true)
	// freqCellRenderer.SetProperty("has-entry", false)
	// freqCellRenderer.SetVisible(true)
	// freqCellRenderer.Connect("editing-started", func(a *gtk.CellRendererCombo, e *gtk.CellEditable, path string) {
	// 	log.Println("editing-started", a, path)
	// })
	// freqCellRenderer.Connect("edited", func(a *gtk.CellRendererCombo, path string, newText string) {
	// 	log.Println("edited", a, path, newText)
	// })
	// freqColumn, err := gtk.TreeViewColumnNewWithAttribute("Frequency", freqCellRenderer, "text", 3)
	// if err != nil {
	// 	return tv, ls, fmt.Errorf("unable to create Name cell column: %v", err.Error())
	// }
	// freqColumn.SetResizable(true)
	// freqColumn.SetClickable(true)
	// freqColumn.SetVisible(true)
	// treeView.AppendColumn(freqColumn)

	// end my attempt for the frequency column

	// for i := range configColumns {
	// 	if configColumnTypes[i] == glib.TYPE_STRING {
	// 		tvc, err := createColumn(configColumns[i], i)
	// 		if err != nil {
	// 			return tv, ls, fmt.Errorf(
	// 				"failed to create text config column %v with id %v: %v",
	// 				configColumns[i],
	// 				i,
	// 				err.Error(),
	// 			)
	// 		}
	// 		treeView.AppendColumn(tvc)
	// 		continue
	// 	}
	// 	if configColumnTypes[i] == glib.TYPE_INT {
	// 		tvc, err := createColumn(configColumns[i], i)
	// 		if err != nil {
	// 			return tv, ls, fmt.Errorf(
	// 				"failed to create text config column %v with id %v: %v",
	// 				configColumns[i],
	// 				i,
	// 				err.Error(),
	// 			)
	// 		}
	// 		treeView.AppendColumn(tvc)
	// 		continue
	// 	}
	// 	if configColumnTypes[i] == glib.TYPE_BOOLEAN {
	// 		tvc, err := createCheckboxColumn(configColumns[i], i, false, listStore, txs)
	// 		if err != nil {
	// 			return tv, ls, fmt.Errorf(
	// 				"failed to create checkbox config column %v with id %v: %v",
	// 				configColumns[i],
	// 				i,
	// 				err.Error(),
	// 			)
	// 		}
	// 		treeView.AppendColumn(tvc)
	// 		continue
	// 	}
	// }

	treeView.SetModel(listStore)

	return treeView, listStore, nil
}

func GetConfigAsTreeView(txs *[]lib.TX, updateResults func()) (tv *gtk.TreeView, ls *gtk.ListStore, err error) {
	treeView, listStore, err := setupConfigTreeView(txs, updateResults)
	if err != nil {
		return tv, ls, fmt.Errorf("failed to set up config tree view: %v", err.Error())
	}

	// add rows to the tree's list store
	for _, tx := range *txs {
		err := addConfigTreeRow(listStore, &tx)
		if err != nil {
			return tv, ls, fmt.Errorf("failed to add config row: %v", err.Error())
		}
	}

	treeView.SetRubberBanding(true)

	return treeView, listStore, nil
}
