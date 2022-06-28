package ui

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	c "finance-planner/constants"
	"finance-planner/lib"
	"finance-planner/state"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func DelConfItem(ws *state.WinState) {
	if len(ws.SelectedConfigItems) > 0 && len(*ws.TX) > 1 {
		// remove the conf item from the UserTX config
		// TODO: remove this spammy logging once things are more efficient
		log.Println(ws.SelectedConfigItems)
		// newUserTX := []lib.TX{}
		// for _, i := range ws.SelectedConfigItems {
		for i := len(ws.SelectedConfigItems) - 1; i >= 0; i-- {
			if ws.SelectedConfigItems[i] > len(*ws.TX) {
				return
			}
			// TODO: remove this logging, but only once you've gotten rid
			// of weird edge cases - for example, clearing the tree view's
			// list store with any selected items causes a large number of
			// selection changes (this is probably very inefficient). To
			// address this, you need to figure out how to un-select values
			log.Println(i, ws.SelectedConfigItems, ws.SelectedConfigItems[i])
			*ws.TX = lib.RemoveTXAtIndex(*ws.TX, ws.SelectedConfigItems[i])
		}

		UpdateResults(ws, false)
		SyncConfigListStore(ws)

		// TODO: old code - jittery and inefficient
		// nb.RemovePage(c.TAB_CONFIG)
		// newConfigSw, newLabel := genConfigView()
		// nb.InsertPage(newConfigSw, newLabel, c.TAB_CONFIG)
		// ui.UpdateResults(ws, false)
		// win.ShowAll()
		// nb.SetCurrentPage(c.TAB_CONFIG)

		ws.SelectedConfigItems = []int{}
	}
}

func AddConfItem(ws *state.WinState) {
	// TODO: create/use helper function that generates new TX instances
	// TODO: refactor
	*ws.TX = append(*ws.TX, lib.TX{
		Order:     len(*ws.TX) + 1,
		Amount:    -500,
		Active:    true,
		Name:      "New",
		Frequency: "WEEKLY",
		Interval:  1,
	})

	UpdateResults(ws, false)
	SyncConfigListStore(ws)
}

func CloneConfItem(ws *state.WinState) {
	if len(ws.SelectedConfigItems) > 0 {
		log.Println(ws.SelectedConfigItems)
		// for _, i := range ws.SelectedConfigItems {
		for i := len(ws.SelectedConfigItems) - 1; i >= 0; i-- {
			if ws.SelectedConfigItems[i] > len(*ws.TX) {
				return
			}
			log.Println(i, ws.SelectedConfigItems, ws.SelectedConfigItems[i])
			*ws.TX = append(
				*ws.TX,
				(*ws.TX)[ws.SelectedConfigItems[i]])
			(*ws.TX)[len(*ws.TX)-1].Order = len(*ws.TX)
		}

		UpdateResults(ws, false)
		SyncConfigListStore(ws)
	}
}

func addConfigTreeRow(listStore *gtk.ListStore, tx *lib.TX) error {
	// gets an iterator for a new row at the end of the list store
	iter := listStore.Append()

	rowData := []interface{}{
		tx.Order,
		tx.MarkupCurrency(lib.CurrencyMarkup(tx.Amount)),
		tx.Active,
		tx.MarkupText(tx.Name),
		tx.MarkupText(tx.Frequency),
		tx.MarkupText(fmt.Sprint(tx.Interval)),
		tx.DoesTXHaveWeekday(c.WeekdayMondayInt),
		tx.DoesTXHaveWeekday(c.WeekdayTuesdayInt),
		tx.DoesTXHaveWeekday(c.WeekdayWednesdayInt),
		tx.DoesTXHaveWeekday(c.WeekdayThursdayInt),
		tx.DoesTXHaveWeekday(c.WeekdayFridayInt),
		tx.DoesTXHaveWeekday(c.WeekdaySaturdayInt),
		tx.DoesTXHaveWeekday(c.WeekdaySundayInt),
		tx.MarkupText(fmt.Sprintf("%v-%v-%v", tx.StartsYear, tx.StartsMonth, tx.StartsDay)),
		tx.MarkupText(fmt.Sprintf("%v-%v-%v", tx.EndsYear, tx.EndsMonth, tx.EndsDay)),
		tx.MarkupText(tx.Note),
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

// ConfigChange is triggered when the user makes a change in the config tree
// view. This function is responsible for finding the underlying TX definition
// that corresponds to the tree view UI item that was changed, by keying off of
// the Order column (which is basically a unique ID). The `path` parameter is
// the value provided from the cell edit event, which is typically something
// like "1:2:5" or simply "1", depending on how the tree is constructed.
func ConfigChange(ws *state.WinState, path string, column int, newValue interface{}) {
	// start by iterating through the list store

	iterFn := func(model *gtk.TreeModel, searchPath *gtk.TreePath, iter *gtk.TreeIter) bool {
		// log.Printf("order changed, searchPath=%v, path=%v", searchPath.String(), path)

		if searchPath.String() == path {
			// key off of the Order column, which contains the unique identifier
			// TODO: elsewhere you need to add a function to assert that Order
			// is unique
			gv, err := ws.ConfigListStore.GetValue(iter, c.COLUMN_ORDER)
			if err != nil {
				log.Printf("failed to get value from config list store: %v", err.Error())
			}

			// convert the value to an int
			val, err := gv.GoValue()
			if err != nil {
				log.Printf("failed to get val string: %v", err.Error())
			}

			valInt := val.(int)
			// TODO: clean this up
			log.Printf("order from list store=%v", valInt)

			// now that we've found the unique ID for this TX definition, we can
			// proceed to find the actual TX definition
			for i := range *ws.TX {
				if (*ws.TX)[i].Order == valInt {
					// now we've found the actual TX definition and we can
					// make changes to the TX and propagate it to the ListStore
					if column == c.COLUMN_ORDER {
						// TODO: add function to enforce uniqueness for all
						// elements and prevent the change if non-unique
						// TODO: validate that the newValue is of int type
						// before unsafe type assertion
						nv, err := strconv.ParseInt(newValue.(string), 10, 64)
						if err != nil {
							log.Printf(
								"failed to convert interval %v to int: %v",
								newValue,
								err.Error(),
							)
						}
						(*ws.TX)[i].Order = int(nv)
						ws.ConfigListStore.Set(iter, []int{c.COLUMN_ORDER}, []interface{}{nv})

					} else if column == c.COLUMN_AMOUNT {
						nv := int(lib.ParseDollarAmount(newValue.(string), false))
						(*ws.TX)[i].Amount = nv
						formatted := (*ws.TX)[i].MarkupText(lib.CurrencyMarkup(nv))
						ws.ConfigListStore.Set(iter, []int{c.COLUMN_AMOUNT}, []interface{}{formatted})

					} else if column == c.COLUMN_ACTIVE {
						nv := !(newValue.(bool))
						(*ws.TX)[i].Active = nv
						ws.ConfigListStore.Set(iter, []int{c.COLUMN_ACTIVE}, []interface{}{nv})

					} else if column == c.COLUMN_NAME {
						nv := newValue.(string)
						(*ws.TX)[i].Name = nv
						mnv := (*ws.TX)[i].MarkupText(nv)
						ws.ConfigListStore.Set(iter, []int{c.COLUMN_NAME}, []interface{}{mnv})

					} else if column == c.COLUMN_FREQUENCY {
						// TODO: refactor for unit testability
						nv := strings.ToUpper(strings.TrimSpace(newValue.(string)))
						if nv == "Y" {
							nv = "YEARLY"
						}
						if nv == "W" {
							nv = "WEEKLY"
						}
						if nv == "M" {
							nv = "MONTHLY"
						}
						if nv != "WEEKLY" && nv != "MONTHLY" && nv != "YEARLY" {
							// TODO: invalid input message dialog
							log.Print("unacceptable value provided; please enter y/m/w/monthly/weekly/yearly")
							break
						}

						(*ws.TX)[i].Frequency = nv
						mnv := (*ws.TX)[i].MarkupText(nv)
						ws.ConfigListStore.Set(iter, []int{c.COLUMN_FREQUENCY}, []interface{}{mnv})

					} else if column == c.COLUMN_INTERVAL {
						nv, err := strconv.ParseInt(newValue.(string), 10, 64)
						if err != nil {
							log.Printf("failed to convert interval %v to int: %v", newValue, err.Error())
						}
						if nv <= 0 {
							nv = 1
						}
						(*ws.TX)[i].Interval = int(nv)
						mnv := (*ws.TX)[i].MarkupText(fmt.Sprint(nv))
						ws.ConfigListStore.Set(iter, []int{c.COLUMN_INTERVAL}, []interface{}{mnv})

					} else if column == c.COLUMN_STARTS {
						nvs := newValue.(string)
						yr, mo, day := lib.ParseYearMonthDateString(
							strings.TrimSpace(nvs),
						)
						(*ws.TX)[i].StartsYear = yr
						(*ws.TX)[i].StartsMonth = mo
						(*ws.TX)[i].StartsDay = day
						nv := fmt.Sprintf("%v-%v-%v", yr, mo, day)
						mnv := (*ws.TX)[i].MarkupText(nv)
						ws.ConfigListStore.Set(iter, []int{c.COLUMN_STARTS}, []interface{}{mnv})

					} else if column == c.COLUMN_ENDS {
						// TODO: refactor similar code from above case
						nvs := newValue.(string)
						yr, mo, day := lib.ParseYearMonthDateString(
							strings.TrimSpace(nvs),
						)
						(*ws.TX)[i].EndsYear = yr
						(*ws.TX)[i].EndsMonth = mo
						(*ws.TX)[i].EndsDay = day
						nv := fmt.Sprintf("%v-%v-%v", yr, mo, day)
						mnv := (*ws.TX)[i].MarkupText(nv)
						ws.ConfigListStore.Set(iter, []int{c.COLUMN_ENDS}, []interface{}{mnv})

					} else if lib.IsWeekday(c.ConfigColumns[column]) {
						// TODO: make this a little more polished; it probably
						// doesn't need to be this involved
						weekday := lib.WeekdayIndex[c.ConfigColumns[column]]
						(*ws.TX)[i].Weekdays = lib.ToggleDayFromWeekdays((*ws.TX)[i].Weekdays, weekday)
						nv := (*ws.TX)[i].DoesTXHaveWeekday(weekday)
						ws.ConfigListStore.Set(iter, []int{column}, []interface{}{nv})

					} else if column == c.COLUMN_NOTE {
						nv := newValue.(string)
						(*ws.TX)[i].Name = nv
						mnv := (*ws.TX)[i].MarkupText(nv)
						ws.ConfigListStore.Set(iter, []int{c.COLUMN_NOTE}, []interface{}{mnv})

					} else {
						log.Printf(
							"warning: column id %v was modified, but there is no case to handle it",
							column,
						)
					}

					break
				}
			}

			return true
		}

		return false
	}

	ws.ConfigListStore.ForEach(iterFn)
	UpdateResults(ws, false)
}

// getOrderColumn builds out an "Order" column, which is an integer column
// that allows the user to control an unsorted default order of recurring
// transactions in the config view
func getOrderColumn(ws *state.WinState) (tvc *gtk.TreeViewColumn, err error) {
	orderCellEditingStarted := func(a *gtk.CellRendererText, e *gtk.CellEditable, path string) {
		log.Println("editing-started", a, path)
	}
	orderCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
		ConfigChange(ws, path, c.COLUMN_ORDER, newText)
	}
	orderCellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tvc, fmt.Errorf("unable to create amount column renderer: %v", err.Error())
	}
	orderCellRenderer.SetProperty("editable", true)
	orderCellRenderer.SetVisible(true)
	orderCellRenderer.Connect("editing-started", orderCellEditingStarted)
	orderCellRenderer.Connect("edited", orderCellEditingFinished)
	orderColumn, err := gtk.TreeViewColumnNewWithAttribute(c.ColumnOrder, orderCellRenderer, "text", c.COLUMN_ORDER)
	if err != nil {
		return tvc, fmt.Errorf("unable to create amount cell column: %v", err.Error())
	}
	orderColumn.SetResizable(true)
	orderColumn.SetClickable(true)
	orderColumn.SetVisible(true)

	orderColumnBtn, err := orderColumn.GetButton()
	if err != nil {
		log.Printf("failed to get active column header button: %v", err.Error())
	}
	orderColumnBtn.ToWidget().Connect(c.GtkSignalClicked, func() {
		if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnOrder, c.Asc) {
			ws.ConfigColumnSort = fmt.Sprintf("%v%v", c.ColumnOrder, c.Desc)
		} else if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnOrder, c.Desc) {
			ws.ConfigColumnSort = c.None
		} else {
			ws.ConfigColumnSort = fmt.Sprintf("%v%v", c.ColumnOrder, c.Asc)
		}
		UpdateResults(ws, false)
		err := SyncConfigListStore(ws)
		if err != nil {
			// TODO: create a "show message" function pointer and call it here
			// if it's not nil?
			log.Printf("failed to sync list store: %v", err.Error())
		}
	})

	return orderColumn, nil
}

// getAmountColumn builds out an "Amount" column, which is an integer column
// that allows the user to input a positive or negative cash amount and it will
// be parsed into currency and displayed as currency as well
func getAmountColumn(ws *state.WinState) (tvc *gtk.TreeViewColumn, err error) {
	amtCellEditingStarted := func(a *gtk.CellRendererText, e *gtk.CellEditable, path string) {
		log.Println("editing-started", a, path)
	}
	amtCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
		ConfigChange(ws, path, c.COLUMN_AMOUNT, newText)
	}
	amtCellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tvc, fmt.Errorf("unable to create amount column renderer: %v", err.Error())
	}
	amtCellRenderer.SetProperty("editable", true)
	amtCellRenderer.SetVisible(true)
	amtCellRenderer.Connect("editing-started", amtCellEditingStarted)
	amtCellRenderer.Connect("edited", amtCellEditingFinished)
	amtColumn, err := gtk.TreeViewColumnNewWithAttribute(c.ColumnAmount, amtCellRenderer, "markup", c.COLUMN_AMOUNT)
	if err != nil {
		return tvc, fmt.Errorf("unable to create amount cell column: %v", err.Error())
	}
	amtColumn.SetResizable(true)
	amtColumn.SetClickable(true)
	amtColumn.SetVisible(true)

	amtColumnBtn, err := amtColumn.GetButton()
	if err != nil {
		log.Printf("failed to get active column header button: %v", err.Error())
	}
	amtColumnBtn.ToWidget().Connect(c.GtkSignalClicked, func() {
		if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnAmount, c.Asc) {
			ws.ConfigColumnSort = fmt.Sprintf("%v%v", c.ColumnAmount, c.Desc)
		} else if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnAmount, c.Desc) {
			ws.ConfigColumnSort = c.None
		} else {
			ws.ConfigColumnSort = fmt.Sprintf("%v%v", c.ColumnAmount, c.Asc)
		}
		log.Printf("amount column clicked, sort column: %v", ws.ConfigColumnSort)
		UpdateResults(ws, false)
		err := SyncConfigListStore(ws)
		if err != nil {
			log.Printf("failed to sync list store: %v", err.Error())
		}
	})

	return amtColumn, nil
}

// getActiveColumn builds out an "Active" column, which is a boolean column
// that allows the user to enable/disable a recurring transaction
func getActiveColumn(ws *state.WinState) (tvc *gtk.TreeViewColumn, err error) {
	activeColumn, err := createCheckboxColumn(
		c.ColumnActive,
		c.COLUMN_ACTIVE,
		false,
		ws.ConfigListStore,
		ws,
	)
	if err != nil {
		return tvc, fmt.Errorf(
			"failed to create checkbox column '%v': %v",
			c.ColumnActive,
			err.Error(),
		)
	}
	activeColumnHeaderBtn, err := activeColumn.GetButton()
	if err != nil {
		log.Printf("failed to get active column header button: %v", err.Error())
	}
	activeColumnHeaderBtn.ToWidget().Connect(c.GtkSignalClicked, func() {
		if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnActive, c.Asc) {
			ws.ConfigColumnSort = fmt.Sprintf("%v%v", c.ColumnActive, c.Desc)
		} else if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnActive, c.Desc) {
			ws.ConfigColumnSort = c.None
		} else {
			ws.ConfigColumnSort = fmt.Sprintf("%v%v", c.ColumnActive, c.Asc)
		}
		log.Printf("active column clicked, sort column: %v", ws.ConfigColumnSort)
		UpdateResults(ws, false)
		err := SyncConfigListStore(ws)
		if err != nil {
			log.Printf("failed to sync list store: %v", err.Error())
		}
	})

	return activeColumn, nil
}

// getNameColumn builds out a "Name" column, which is a string column that
// allows users to set the name of a recurring transaction
func getNameColumn(ws *state.WinState) (tvc *gtk.TreeViewColumn, err error) {
	nameCellEditingStarted := func(a *gtk.CellRendererText, e *gtk.CellEditable, path string) {
		log.Println("editing-started", a, path)
	}
	nameCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
		ConfigChange(ws, path, c.COLUMN_NAME, newText)
	}
	nameCellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tvc, fmt.Errorf("unable to create Name column renderer: %v", err.Error())
	}
	nameCellRenderer.SetProperty("editable", true)
	nameCellRenderer.SetVisible(true)
	nameCellRenderer.Connect("editing-started", nameCellEditingStarted)
	nameCellRenderer.Connect("edited", nameCellEditingFinished)
	nameColumn, err := gtk.TreeViewColumnNewWithAttribute(c.ColumnName, nameCellRenderer, "markup", c.COLUMN_NAME)
	if err != nil {
		return tvc, fmt.Errorf("unable to create Name cell column: %v", err.Error())
	}
	nameColumn.SetResizable(true)
	nameColumn.SetClickable(true)
	nameColumn.SetVisible(true)
	nameColumnHeaderBtn, err := nameColumn.GetButton()
	if err != nil {
		log.Printf("failed to get name column header button: %v", err.Error())
	}
	nameColumnHeaderBtn.ToWidget().Connect(c.GtkSignalClicked, func() {
		if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnName, c.Asc) {
			ws.ConfigColumnSort = fmt.Sprintf("%v%v", c.ColumnName, c.Desc)
		} else if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnName, c.Desc) {
			ws.ConfigColumnSort = c.None
		} else {
			ws.ConfigColumnSort = fmt.Sprintf("%v%v", c.ColumnName, c.Asc)
		}
		log.Printf("name column clicked, sort column: %v", ws.ConfigColumnSort)
		UpdateResults(ws, false)
		err := SyncConfigListStore(ws)
		if err != nil {
			log.Printf("failed to sync list store: %v", err.Error())
		}
	})

	return nameColumn, nil
}

// getFrequencyColumn builds out a "Frequency" column, which is a string
// column that allows the user to specify monthly/weekly/yearly recurrences
// TODO: make this form of input more user friendly - currently it requires
// users to just simply type "MONTHLY"/"YEARLY"/"WEEKLY"
func getFrequencyColumn(ws *state.WinState) (tvc *gtk.TreeViewColumn, err error) {
	freqCellEditingStarted := func(a *gtk.CellRendererText, e *gtk.CellEditable, path string) {
		log.Println("editing-started", a, path)
	}
	freqCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
		ConfigChange(ws, path, c.COLUMN_FREQUENCY, newText)
	}
	freqCellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tvc, fmt.Errorf("unable to create Frequency column renderer: %v", err.Error())
	}
	freqCellRenderer.SetProperty("editable", true)
	freqCellRenderer.SetVisible(true)
	freqCellRenderer.Connect("editing-started", freqCellEditingStarted)
	freqCellRenderer.Connect("edited", freqCellEditingFinished)
	freqColumn, err := gtk.TreeViewColumnNewWithAttribute(c.ColumnFrequency, freqCellRenderer, "markup", c.COLUMN_FREQUENCY)
	if err != nil {
		return tvc, fmt.Errorf("unable to create Frequency cell column: %v", err.Error())
	}
	freqColumn.SetResizable(true)
	freqColumn.SetClickable(true)
	freqColumn.SetVisible(true)

	freqColumnBtn, err := freqColumn.GetButton()
	if err != nil {
		log.Printf("failed to get frequency column header button: %v", err.Error())
	}
	freqColumnBtn.ToWidget().Connect(c.GtkSignalClicked, func() {
		if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnFrequency, c.Asc) {
			ws.ConfigColumnSort = fmt.Sprintf("%v%v", c.ColumnFrequency, c.Desc)
		} else if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnFrequency, c.Desc) {
			ws.ConfigColumnSort = c.None
		} else {
			ws.ConfigColumnSort = fmt.Sprintf("%v%v", c.ColumnFrequency, c.Asc)
		}
		log.Printf("frequency column clicked, sort column: %v", ws.ConfigColumnSort)
		UpdateResults(ws, false)
		err := SyncConfigListStore(ws)
		if err != nil {
			log.Printf("failed to sync list store: %v", err.Error())
		}
	})

	return freqColumn, nil
}

// getIntervalColumn builds out an "Interval" column, which is an integer
// column that allows the user to specify the interval at which a recurring
// transaction occurs - for example, interval=2 means that every 2
// occurrences of the recurring transaction's calendar sequence, the transaction
// will occur, as opposed to interval=1, where it would occur every single
// calendar occurrence. e.g. bi-monthly on the 7th versus every month on the
// 7th.
// TODO: refactor this to use signal constants, and also refactor functions
// for more unit testability
func getIntervalColumn(ws *state.WinState) (tvc *gtk.TreeViewColumn, err error) {
	intervalCellEditingStarted := func(a *gtk.CellRendererText, e *gtk.CellEditable, path string) {
		log.Println("editing-started", a, path)
	}
	intervalCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
		ConfigChange(ws, path, c.COLUMN_INTERVAL, newText)
	}
	intervalCellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tvc, fmt.Errorf("unable to create Interval column renderer: %v", err.Error())
	}
	intervalCellRenderer.SetProperty("editable", true)
	intervalCellRenderer.SetVisible(true)
	intervalCellRenderer.Connect("editing-started", intervalCellEditingStarted)
	intervalCellRenderer.Connect("edited", intervalCellEditingFinished)
	intervalColumn, err := gtk.TreeViewColumnNewWithAttribute(c.ColumnInterval, intervalCellRenderer, "markup", c.COLUMN_INTERVAL)
	if err != nil {
		return tvc, fmt.Errorf("unable to create Interval cell column: %v", err.Error())
	}
	intervalColumn.SetResizable(true)
	intervalColumn.SetClickable(true)
	intervalColumn.SetVisible(true)
	intervalColumnBtn, err := intervalColumn.GetButton()
	if err != nil {
		log.Printf("failed to get interval column header button: %v", err.Error())
	}
	intervalColumnBtn.ToWidget().Connect(c.GtkSignalClicked, func() {
		if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnInterval, c.Asc) {
			ws.ConfigColumnSort = fmt.Sprintf("%v%v", c.ColumnInterval, c.Desc)
		} else if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnInterval, c.Desc) {
			ws.ConfigColumnSort = c.None
		} else {
			ws.ConfigColumnSort = fmt.Sprintf("%v%v", c.ColumnInterval, c.Asc)
		}
		log.Printf("interval column clicked, sort column: %v", ws.ConfigColumnSort)
		UpdateResults(ws, false)
		err := SyncConfigListStore(ws)
		if err != nil {
			log.Printf("failed to sync list store: %v", err.Error())
		}
	})

	return intervalColumn, nil
}

// getStartsColumn builds out a "Starts" column, which is a string column that
// allows the user to type in a starting date, such as 2020-02-01.
// TODO: refactoring and cleanup
func getStartsColumn(ws *state.WinState) (tvc *gtk.TreeViewColumn, err error) {
	startsCellEditingStarted := func(a *gtk.CellRendererText, e *gtk.CellEditable, path string) {
		log.Println("editing-started", a, path)
	}
	startsCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
		ConfigChange(ws, path, c.COLUMN_STARTS, newText)
	}
	startsCellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tvc, fmt.Errorf("unable to create Starts column renderer: %v", err.Error())
	}
	startsCellRenderer.SetProperty("editable", true)
	startsCellRenderer.SetVisible(true)
	startsCellRenderer.Connect("editing-started", startsCellEditingStarted)
	startsCellRenderer.Connect("edited", startsCellEditingFinished)
	startsColumn, err := gtk.TreeViewColumnNewWithAttribute(c.ColumnStarts, startsCellRenderer, "markup", c.COLUMN_STARTS)
	if err != nil {
		return tvc, fmt.Errorf("unable to create Starts cell column: %v", err.Error())
	}
	startsColumn.SetResizable(true)
	startsColumn.SetClickable(true)
	startsColumn.SetVisible(true)
	startsColumnHeaderBtn, err := startsColumn.GetButton()
	if err != nil {
		log.Printf("failed to get starts column header button: %v", err.Error())
	}
	startsColumnHeaderBtn.ToWidget().Connect(c.GtkSignalClicked, func() {
		if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnStarts, c.Asc) {
			ws.ConfigColumnSort = fmt.Sprintf("%v%v", c.ColumnStarts, c.Desc)
		} else if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnStarts, c.Desc) {
			ws.ConfigColumnSort = c.None
		} else {
			ws.ConfigColumnSort = fmt.Sprintf("%v%v", c.ColumnStarts, c.Asc)
		}
		log.Printf("starts column clicked, sort column: %v", ws.ConfigColumnSort)
		UpdateResults(ws, false)
		err := SyncConfigListStore(ws)
		if err != nil {
			log.Printf("failed to sync list store: %v", err.Error())
		}
	})

	return startsColumn, nil
}

// getEndsColumn builds out an "Ends" column, which is a string column that
// allows the user to type in an ending date, such as 2020-02-01.
// TODO: refactoring and cleanup
func getEndsColumn(ws *state.WinState) (tvc *gtk.TreeViewColumn, err error) {
	endsCellEditingStarted := func(a *gtk.CellRendererText, e *gtk.CellEditable, path string) {
		log.Println("editing-started", a, path)
	}
	endsCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
		ConfigChange(ws, path, c.COLUMN_ENDS, newText)
	}
	endsCellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tvc, fmt.Errorf("unable to create Ends column renderer: %v", err.Error())
	}
	endsCellRenderer.SetProperty("editable", true)
	endsCellRenderer.SetVisible(true)
	endsCellRenderer.Connect("editing-started", endsCellEditingStarted)
	endsCellRenderer.Connect("edited", endsCellEditingFinished)
	endsColumn, err := gtk.TreeViewColumnNewWithAttribute(c.ColumnEnds, endsCellRenderer, "markup", c.COLUMN_ENDS)
	if err != nil {
		return tvc, fmt.Errorf("unable to create Ends cell column: %v", err.Error())
	}
	endsColumn.SetResizable(true)
	endsColumn.SetClickable(true)
	endsColumn.SetVisible(true)
	endsColumnHeaderBtn, err := endsColumn.GetButton()
	if err != nil {
		log.Printf("failed to get ends column header button: %v", err.Error())
	}
	endsColumnHeaderBtn.ToWidget().Connect(c.GtkSignalClicked, func() {
		if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnEnds, c.Asc) {
			ws.ConfigColumnSort = fmt.Sprintf("%v%v", c.ColumnEnds, c.Desc)
		} else if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnEnds, c.Desc) {
			ws.ConfigColumnSort = c.None
		} else {
			ws.ConfigColumnSort = fmt.Sprintf("%v%v", c.ColumnEnds, c.Asc)
		}
		log.Printf("ends column clicked, sort column: %v", ws.ConfigColumnSort)
		UpdateResults(ws, false)
		err := SyncConfigListStore(ws)
		if err != nil {
			log.Printf("failed to sync list store: %v", err.Error())
		}
	})

	return endsColumn, nil
}

// getNotesColumn builds out a "Note" column, which is a string column that
// allows the user to type in whatever notes they want for a recurring
// transaction.
// TODO: refactoring and cleanup
func getNotesColumn(ws *state.WinState) (tvc *gtk.TreeViewColumn, err error) {
	// notes column
	notesCellEditingStarted := func(a *gtk.CellRendererText, e *gtk.CellEditable, path string) {
		log.Println("editing-started", a, path)
	}
	notesCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
		ConfigChange(ws, path, c.COLUMN_NOTE, newText)
	}
	notesCellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tvc, fmt.Errorf("unable to create Notes column renderer: %v", err.Error())
	}
	notesCellRenderer.SetProperty("editable", true)
	notesCellRenderer.SetVisible(true)
	notesCellRenderer.Connect("editing-started", notesCellEditingStarted)
	notesCellRenderer.Connect("edited", notesCellEditingFinished)
	notesColumn, err := gtk.TreeViewColumnNewWithAttribute(c.ColumnNote, notesCellRenderer, "markup", c.COLUMN_NOTE)
	if err != nil {
		return tvc, fmt.Errorf("unable to create Notes cell column: %v", err.Error())
	}
	notesColumn.SetResizable(true)
	notesColumn.SetClickable(true)
	notesColumn.SetVisible(true)
	notesColumnHeaderBtn, err := notesColumn.GetButton()
	if err != nil {
		log.Printf("failed to get notes column header button: %v", err.Error())
	}
	notesColumnHeaderBtn.ToWidget().Connect(c.GtkSignalClicked, func() {
		if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnNote, c.Asc) {
			ws.ConfigColumnSort = fmt.Sprintf("%v%v", c.ColumnNote, c.Desc)
		} else if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnNote, c.Desc) {
			ws.ConfigColumnSort = c.None
		} else {
			ws.ConfigColumnSort = fmt.Sprintf("%v%v", c.ColumnNote, c.Asc)
		}
		log.Printf("notes column clicked, sort column: %v", ws.ConfigColumnSort)
		UpdateResults(ws, false)
		err := SyncConfigListStore(ws)
		if err != nil {
			log.Printf("failed to sync list store: %v", err.Error())
		}
	})

	return notesColumn, nil
}

// Creates a tree view and the list store that holds its data
func setupConfigTreeView(ws *state.WinState) (tv *gtk.TreeView, err error) {
	treeView, err := gtk.TreeViewNew()
	if err != nil {
		return tv, fmt.Errorf("unable to create config tree view: %v", err.Error())
	}

	orderColumn, err := getOrderColumn(ws)
	if err != nil {
		return tv, fmt.Errorf("failed to create config order column: %v", err.Error())
	}
	treeView.AppendColumn(orderColumn)

	amtColumn, err := getAmountColumn(ws)
	if err != nil {
		return tv, fmt.Errorf("failed to create config amount column: %v", err.Error())
	}
	treeView.AppendColumn(amtColumn)

	activeColumn, err := getActiveColumn(ws)
	if err != nil {
		return tv, fmt.Errorf("failed to create config active column: %v", err.Error())
	}
	treeView.AppendColumn(activeColumn)

	nameColumn, err := getNameColumn(ws)
	if err != nil {
		return tv, fmt.Errorf("failed to create config name column: %v", err.Error())
	}
	treeView.AppendColumn(nameColumn)

	freqColumn, err := getFrequencyColumn(ws)
	if err != nil {
		return tv, fmt.Errorf("failed to create config frequency column: %v", err.Error())
	}
	treeView.AppendColumn(freqColumn)

	intervalColumn, err := getIntervalColumn(ws)
	if err != nil {
		return tv, fmt.Errorf("failed to create config interval column: %v", err.Error())
	}
	treeView.AppendColumn(intervalColumn)

	// weekday columns
	for i := range c.Weekdays {
		// prevents pointers from changing which column is referred to
		weekdayIndex := int(i)
		weekday := string(c.Weekdays[weekdayIndex])

		// offset by 4 previous columns
		weekdayColumn, err := createCheckboxColumn(
			weekday,
			c.COLUMN_MONDAY+weekdayIndex,
			false,
			ws.ConfigListStore,
			ws,
		)
		if err != nil {
			return tv, fmt.Errorf(
				"failed to create checkbox config column '%v': %v", weekday, err.Error(),
			)
		}
		weekdayColumnBtn, err := weekdayColumn.GetButton()
		if err != nil {
			log.Printf("failed to get frequency column header button: %v", err.Error())
		}
		weekdayColumnBtn.ToWidget().Connect(c.GtkSignalClicked, func() {
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", weekday, c.Asc) {
				ws.ConfigColumnSort = fmt.Sprintf("%v%v", weekday, c.Desc)
			} else if ws.ConfigColumnSort == fmt.Sprintf("%v%v", weekday, c.Desc) {
				ws.ConfigColumnSort = c.None
			} else {
				ws.ConfigColumnSort = fmt.Sprintf("%v%v", weekday, c.Asc)
			}
			log.Printf("%v column clicked, sort column: %v, %v", weekday, ws.ConfigColumnSort, weekdayIndex)
			UpdateResults(ws, false)
			err := SyncConfigListStore(ws)
			if err != nil {
				log.Printf("failed to sync list store: %v", err.Error())
			}
		})
		treeView.AppendColumn(weekdayColumn)
	}

	startsColumn, err := getStartsColumn(ws)
	if err != nil {
		return tv, fmt.Errorf("failed to create config starts column: %v", err.Error())
	}
	treeView.AppendColumn(startsColumn)

	endsColumn, err := getEndsColumn(ws)
	if err != nil {
		return tv, fmt.Errorf("failed to create config ends column: %v", err.Error())
	}
	treeView.AppendColumn(endsColumn)

	notesColumn, err := getNotesColumn(ws)
	if err != nil {
		return tv, fmt.Errorf("failed to create config notes column: %v", err.Error())
	}
	treeView.AppendColumn(notesColumn)

	treeView.SetModel(ws.ConfigListStore)

	return treeView, nil
}

// GetNewConfigListStore creates a list store. This is what holds the data
// that will be shown on our results tree view
func GetNewConfigListStore() (ls *gtk.ListStore, err error) {
	ls, err = gtk.ListStoreNew(
		glib.TYPE_INT,
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
		glib.TYPE_STRING,
	)
	if err != nil {
		return ls, fmt.Errorf("unable to create config list store: %v", err.Error())
	}

	return
}

func SyncConfigListStore(ws *state.WinState) error {
	// sort first
	sort.SliceStable(
		*ws.TX,
		func(i, j int) bool {
			// invisible order column (default when no sort is set)
			if ws.ConfigColumnSort == c.None {
				return (*ws.TX)[j].Order > (*ws.TX)[i].Order
			}

			// Order
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnOrder, c.Asc) {
				return (*ws.TX)[j].Order > (*ws.TX)[i].Order
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnOrder, c.Desc) {
				return (*ws.TX)[i].Order > (*ws.TX)[j].Order
			}

			// active
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnActive, c.Asc) {
				return (*ws.TX)[j].Active
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnActive, c.Desc) {
				return (*ws.TX)[i].Active
			}

			// weekdays
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.WeekdayMonday, c.Asc) {
				return (*ws.TX)[j].DoesTXHaveWeekday(c.WeekdayMondayInt)
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.WeekdayMonday, c.Desc) {
				return (*ws.TX)[i].DoesTXHaveWeekday(c.WeekdayMondayInt)
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.WeekdayTuesday, c.Asc) {
				return (*ws.TX)[j].DoesTXHaveWeekday(c.WeekdayTuesdayInt)
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.WeekdayTuesday, c.Desc) {
				return (*ws.TX)[i].DoesTXHaveWeekday(c.WeekdayTuesdayInt)
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.WeekdayWednesday, c.Asc) {
				return (*ws.TX)[j].DoesTXHaveWeekday(c.WeekdayWednesdayInt)
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.WeekdayWednesday, c.Desc) {
				return (*ws.TX)[i].DoesTXHaveWeekday(c.WeekdayWednesdayInt)
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.WeekdayThursday, c.Asc) {
				return (*ws.TX)[j].DoesTXHaveWeekday(c.WeekdayThursdayInt)
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.WeekdayThursday, c.Desc) {
				return (*ws.TX)[i].DoesTXHaveWeekday(c.WeekdayThursdayInt)
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.WeekdayFriday, c.Asc) {
				return (*ws.TX)[j].DoesTXHaveWeekday(c.WeekdayFridayInt)
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.WeekdayFriday, c.Desc) {
				return (*ws.TX)[i].DoesTXHaveWeekday(c.WeekdayFridayInt)
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.WeekdaySaturday, c.Asc) {
				return (*ws.TX)[j].DoesTXHaveWeekday(c.WeekdaySaturdayInt)
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.WeekdaySaturday, c.Desc) {
				return (*ws.TX)[i].DoesTXHaveWeekday(c.WeekdaySaturdayInt)
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.WeekdaySunday, c.Asc) {
				return (*ws.TX)[j].DoesTXHaveWeekday(c.WeekdaySundayInt)
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.WeekdaySunday, c.Desc) {
				return (*ws.TX)[i].DoesTXHaveWeekday(c.WeekdaySundayInt)
			}

			// other numeric columns
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnAmount, c.Asc) {
				return (*ws.TX)[j].Amount > (*ws.TX)[i].Amount
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnAmount, c.Desc) {
				return (*ws.TX)[i].Amount > (*ws.TX)[j].Amount
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnFrequency, c.Asc) {
				return (*ws.TX)[j].Frequency > (*ws.TX)[i].Frequency
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnFrequency, c.Desc) {
				return (*ws.TX)[i].Frequency > (*ws.TX)[j].Frequency
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnInterval, c.Asc) {
				return (*ws.TX)[j].Interval > (*ws.TX)[i].Interval
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnInterval, c.Desc) {
				return (*ws.TX)[i].Interval > (*ws.TX)[j].Interval
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnNote, c.Asc) {
				return strings.ToLower((*ws.TX)[j].Note) > strings.ToLower((*ws.TX)[i].Note)
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnNote, c.Desc) {
				return strings.ToLower((*ws.TX)[i].Note) > strings.ToLower((*ws.TX)[j].Note)
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnName, c.Asc) {
				return strings.ToLower((*ws.TX)[j].Name) > strings.ToLower((*ws.TX)[i].Name)
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnName, c.Desc) {
				return strings.ToLower((*ws.TX)[i].Name) > strings.ToLower((*ws.TX)[j].Name)
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnStarts, c.Asc) {
				jstarts := fmt.Sprintf("%v-%v-%v", (*ws.TX)[j].StartsYear, (*ws.TX)[j].StartsMonth, (*ws.TX)[j].StartsDay)
				istarts := fmt.Sprintf("%v-%v-%v", (*ws.TX)[i].StartsYear, (*ws.TX)[i].StartsMonth, (*ws.TX)[i].StartsDay)
				return jstarts > istarts
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnStarts, c.Desc) {
				jstarts := fmt.Sprintf("%v-%v-%v", (*ws.TX)[j].StartsYear, (*ws.TX)[j].StartsMonth, (*ws.TX)[j].StartsDay)
				istarts := fmt.Sprintf("%v-%v-%v", (*ws.TX)[i].StartsYear, (*ws.TX)[i].StartsMonth, (*ws.TX)[i].StartsDay)
				return istarts > jstarts
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnEnds, c.Asc) {
				jends := fmt.Sprintf("%v-%v-%v", (*ws.TX)[j].EndsYear, (*ws.TX)[j].EndsMonth, (*ws.TX)[j].EndsDay)
				iends := fmt.Sprintf("%v-%v-%v", (*ws.TX)[i].EndsYear, (*ws.TX)[i].EndsMonth, (*ws.TX)[i].EndsDay)
				return jends > iends
			}
			if ws.ConfigColumnSort == fmt.Sprintf("%v%v", c.ColumnEnds, c.Desc) {
				jEnds := fmt.Sprintf("%v-%v-%v", (*ws.TX)[j].EndsYear, (*ws.TX)[j].EndsMonth, (*ws.TX)[j].EndsDay)
				iends := fmt.Sprintf("%v-%v-%v", (*ws.TX)[i].EndsYear, (*ws.TX)[i].EndsMonth, (*ws.TX)[i].EndsDay)
				return iends > jEnds
			}
			return false
			// return txs[j].Date.After(txs[i].Date)
		},
	)

	ws.ConfigListStore.Clear()

	// add rows to the tree's list store
	for _, tx := range *ws.TX {
		// skip adding this row to the list store if this record is inactive,
		// and HideInactive is true
		if !tx.Active && ws.HideInactive == true {
			continue
		}
		err := addConfigTreeRow(ws.ConfigListStore, &tx)
		if err != nil {
			return fmt.Errorf("failed to sync list store: %v", err.Error())
		}
	}

	return nil
}

func GetConfigAsTreeView(ws *state.WinState) (tv *gtk.TreeView, err error) {
	treeView, err := setupConfigTreeView(ws)
	if err != nil {
		return tv, fmt.Errorf("failed to set up config tree view: %v", err.Error())
	}

	err = SyncConfigListStore(ws)
	if err != nil {
		return tv, fmt.Errorf("failed to add config row: %v", err.Error())
	}

	treeView.SetRubberBanding(true)

	return treeView, nil
}

func GetHideInactiveCheckbox(ws *state.WinState) *gtk.CheckButton {
	hideInactiveCheckBoxClickedHandler := func(chkBtn *gtk.CheckButton) {
		ws.HideInactive = !ws.HideInactive
		chkBtn.SetActive(ws.HideInactive)
		SyncConfigListStore(ws)
	}

	hideInactiveCheckbox, err := gtk.CheckButtonNewWithMnemonic(c.HideInactiveBtnLabel)
	if err != nil {
		log.Fatal("failed to create 'hide inactive' checkbox:", err)
	}

	SetSpacerMarginsGtkCheckBtn(hideInactiveCheckbox)

	hideInactiveCheckbox.SetActive(ws.HideInactive)

	hideInactiveCheckbox.Connect(c.GtkSignalClicked, hideInactiveCheckBoxClickedHandler)

	return hideInactiveCheckbox
}

func GetConfEditButtons(ws *state.WinState) (*gtk.Button, *gtk.Button, *gtk.Button) {
	addConfItemBtn, err := gtk.ButtonNewWithMnemonic(c.AddBtnLabel)
	if err != nil {
		log.Fatal("failed to create add conf item button:", err)
	}
	SetSpacerMarginsGtkBtn(addConfItemBtn)

	delConfItemBtn, err := gtk.ButtonNewWithMnemonic(c.DelBtnLabel)
	if err != nil {
		log.Fatal("failed to create add conf item button:", err)
	}
	SetSpacerMarginsGtkBtn(delConfItemBtn)

	cloneConfItemBtn, err := gtk.ButtonNewWithMnemonic(c.CloneBtnLabel)
	if err != nil {
		log.Fatal("failed to create clone conf item button:", err)
	}
	SetSpacerMarginsGtkBtn(cloneConfItemBtn)

	return addConfItemBtn, delConfItemBtn, cloneConfItemBtn
}

func GetConfigTab(ws *state.WinState) (*gtk.ScrolledWindow, *gtk.Label) {
	// TODO: refactor the config tree view list store using the same
	// approach that was used for the results list store
	configTreeView, err := GetConfigAsTreeView(ws)
	if err != nil {
		log.Fatalf("failed to get config as tree view: %v", err.Error())
	}

	configTreeView.SetEnableSearch(false)
	configTab, err := gtk.LabelNew(c.ConfigTabLabel)
	if err != nil {
		log.Fatalf("failed to set config tab label: %v", err.Error())
	}
	configTreeSelection, err := configTreeView.GetSelection()
	if err != nil {
		log.Fatalf("failed to get results tree vie sel: %v", err.Error())
	}
	configTreeSelection.SetMode(gtk.SELECTION_MULTIPLE)

	selectionChanged := func(s *gtk.TreeSelection) {
		rows := s.GetSelectedRows(ws.ConfigListStore)
		// items := make([]string, 0, rows.Length())
		ws.SelectedConfigItems = []int{}

		for l := rows; l != nil; l = l.Next() {
			path := l.Data().(*gtk.TreePath)
			i, _ := strconv.ParseInt(path.String(), 10, 64)
			ws.SelectedConfigItems = append(ws.SelectedConfigItems, int(i))
			// iter, _ := configListStore.GetIter(path)
			// value, _ := configListStore.GetValue(iter, 0)
			// log.Println(path, iter, value)
			// str, _ := value.GetString()
			// items = append(items, str)
		}
		log.Printf("selection changed: %v, %v", s, ws.SelectedConfigItems)
	}

	configTreeSelection.Connect(c.GtkSignalChanged, selectionChanged)
	configSw, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		log.Fatal("failed to create scrolled window:", err)
	}
	configSw.Add(configTreeView)
	configSw.SetHExpand(true)
	configSw.SetVExpand(true)

	// TODO: mess with these more; it's preferable to have the tree view
	// a little more tight against the margins, but may be preferable in
	// some dimensions
	// configSw.SetMarginTop(c.UISpacer)
	// configSw.SetMarginBottom(c.UISpacer)
	// configSw.SetMarginStart(c.UISpacer)
	// configSw.SetMarginEnd(c.UISpacer)

	return configSw, configTab
}

// TODO: refactor and explain this more
// LoadConfig launches the GTK file chooser dialog and opens a config file for
// usage in the application. Depending on the value of openInNewWindow, it will
// either load the config in the existing window, or launch a completely new
// window with its own independent WinState (this function doesn't touch that
// new & independent WinState, though; it's abstracted inside the `primary`
// function call)
func LoadConfig(
	ws *state.WinState,
	primary func(application *gtk.Application, filename string) *state.WinState,
	openInNewWindow bool,
) {
	p, err := gtk.FileChooserDialogNewWith2Buttons(
		"Load config",
		ws.Win,
		gtk.FILE_CHOOSER_ACTION_OPEN,
		"_Load",
		gtk.RESPONSE_OK,
		"_Cancel",
		gtk.RESPONSE_CANCEL,
	)
	if err != nil {
		log.Fatal("failed to create load config file picker", err.Error())
	}
	p.Connect(c.ActionClose, func() { p.Close() })
	p.Connect("response", func(dialog *gtk.FileChooserDialog, resp int) {
		if resp == int(gtk.RESPONSE_OK) {
			// folder, _ := dialog.FileChooser.GetCurrentFolder()
			// GetFilename includes the full path and file name
			ws.OpenFileName = dialog.FileChooser.GetFilename()
			if openInNewWindow {
				p.Close()
				p.Destroy()
				newWinState := primary(ws.App, ws.OpenFileName)
				newWinState.Win.ShowAll()
				return
			} else {
				*ws.TX, err = lib.LoadConfig(ws.OpenFileName)
				if err != nil {
					m := fmt.Sprintf("Failed to load config \"%v\": %v", ws.OpenFileName, err.Error())
					d := gtk.MessageDialogNew(ws.Win, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "%s", m)
					log.Println(m)
					d.Run()
					d.Destroy()
					p.Close()
					p.Destroy()
					return
				}
				extraDialogMessageText := ""
				if len(*ws.TX) == 0 {
					extraDialogMessageText = " The configuration was empty, so a sample recurring transaction has been added."
					*ws.TX = []lib.TX{{
						Order:     1,
						Amount:    -500,
						Active:    true,
						Name:      "New",
						Frequency: "WEEKLY",
						Interval:  1,
					}}
				}
				ws.Notebook.RemovePage(c.TAB_CONFIG)
				newConfigSw, newLabel := GetConfigTab(ws)
				ws.Notebook.InsertPage(newConfigSw, newLabel, c.TAB_CONFIG)
				UpdateResults(ws, false)
				ws.Win.ShowAll()
				ws.Notebook.SetCurrentPage(c.TAB_CONFIG)
				ws.Header.SetSubtitle(ws.OpenFileName)
				m := fmt.Sprintf("Success! Loaded file \"%v\" successfully.%v", ws.OpenFileName, extraDialogMessageText)
				d := gtk.MessageDialogNew(ws.Win, gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_OK, "%s", m)
				log.Println(m)
				p.Close()
				p.Destroy()
				d.Run()
				d.Destroy()
				return
			}
		}
		p.Close()
		p.Destroy()
	})
	p.Dialog.ShowAll()
}

// GetConfigBaseComponents returns the base components that are required in
// order to view the transaction tree view presented inside of a GTK notebook
// tab.
func GetConfigBaseComponents(ws *state.WinState) (*gtk.Grid, *gtk.Label) {
	configGrid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("failed to create grid:", err)
	}

	configGrid.SetOrientation(gtk.ORIENTATION_VERTICAL)
	configSw, configTab := GetConfigTab(ws)
	configGrid.Attach(configSw, 0, 0, 1, 1)

	return configGrid, configTab
}
