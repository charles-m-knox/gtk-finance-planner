package ui

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"git.cmcode.dev/cmcode/gtk-finance-planner/constants"
	"git.cmcode.dev/cmcode/gtk-finance-planner/oldutil"
	"git.cmcode.dev/cmcode/gtk-finance-planner/state"

	lib "git.cmcode.dev/cmcode/finance-planner-lib"
	"git.cmcode.dev/cmcode/uuid"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// ClearAllSelections clears out the selected items in the config view
func ClearAllSelections(ws *state.WinState) {
	ws.SelectedConfIDs = make(map[string]bool)
}

func DelConfItem(ws *state.WinState) {
	if len(*ws.TX) <= 1 {
		return
	}

	for id, sel := range ws.SelectedConfIDs {
		if !sel {
			continue
		}

		lib.RemoveTXByID(ws.TX, id)
	}

	ClearAllSelections(ws)

	UpdateResults(ws, false)
	SyncConfigListStore(ws)
}

func AddConfItem(ws *state.WinState) {
	now := time.Now()
	n := lib.GetNewTX(now)

	*ws.TX = append(*ws.TX, n)

	// scroll to end of vertical view, since new conf items are added at
	// the bottom
	SetConfigScrollPosition(ws, 65535, -1)

	ClearAllSelections(ws)

	UpdateResults(ws, false)
	SyncConfigListStore(ws)

	RestoreConfigScrollPosition(ws)
}

func CloneConfItem(ws *state.WinState) {
	if len(ws.SelectedConfIDs) <= 0 {
		return
	}

	now := time.Now()

	for id, sel := range ws.SelectedConfIDs {
		if !sel {
			continue
		}

		i, err := lib.GetTXByID(ws.TX, id)
		if err != nil {
			log.Printf("config item with id=%v was missing", id)
			continue
		}

		*ws.TX = append(
			*ws.TX,
			(*ws.TX)[i],
		)

		// j is the index of the item we just added
		j := len(*ws.TX) - 1

		// update the latest item in the list
		(*ws.TX)[j].UpdatedAt = now
		(*ws.TX)[j].CreatedAt = now
		(*ws.TX)[j].ID = uuid.New()
	}

	ClearAllSelections(ws)

	UpdateResults(ws, false)
	SyncConfigListStore(ws)
}

// GetTXAsRow builds a GTK treeview-compatible set of fields & columns for a
// provided TX definition.
// TODO: refactor this to be more flexible. For example, it would be nice to
// be able to hide/show some columns. This could maybe be done with a map.
func GetTXAsRow(tx *lib.TX) (cells []interface{}, columns []int) {
	cells = []interface{}{
		lib.FormatAsCurrency(tx.Amount), // tx.MarkupCurrency(lib.CurrencyMarkup(tx.Amount)),
		tx.Active,
		tx.Name,                 // tx.MarkupText(tx.Name),
		tx.Frequency,            // tx.MarkupText(tx.Frequency),
		fmt.Sprint(tx.Interval), // tx.MarkupText(fmt.Sprint(tx.Interval)),
		tx.Weekdays[constants.WeekdayMondayInt],
		tx.Weekdays[constants.WeekdayTuesdayInt],
		tx.Weekdays[constants.WeekdayWednesdayInt],
		tx.Weekdays[constants.WeekdayThursdayInt],
		tx.Weekdays[constants.WeekdayFridayInt],
		tx.Weekdays[constants.WeekdaySaturdayInt],
		tx.Weekdays[constants.WeekdaySundayInt],
		fmt.Sprintf("%v-%v-%v", tx.StartsYear, tx.StartsMonth, tx.StartsDay), // tx.MarkupText(fmt.Sprintf("%v-%v-%v", tx.StartsYear, tx.StartsMonth, tx.StartsDay)),
		fmt.Sprintf("%v-%v-%v", tx.EndsYear, tx.EndsMonth, tx.EndsDay),       // tx.MarkupText(fmt.Sprintf("%v-%v-%v", tx.EndsYear, tx.EndsMonth, tx.EndsDay)),
		tx.Note, // tx.MarkupText(tx.Note),
		tx.ID,
		tx.CreatedAt.Format(time.RFC3339),
		tx.UpdatedAt.Format(time.RFC3339),
	}

	columns = []int{}
	for i := range cells {
		columns = append(columns, i)
	}

	return cells, columns
}

func addConfigTreeRow(ls *gtk.ListStore, tx *lib.TX) error {
	// gets an iterator for a new row at the end of the list store
	iter := ls.Append()

	cells, columns := GetTXAsRow(tx)

	// Set the contents of the list store row that the iterator represents
	err := ls.Set(iter, columns, cells)
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
		if searchPath.String() != path {
			return false
		}

		val, err := oldutil.GetListStoreValue(ws.ConfigListStore, iter, constants.COLUMN_ID)
		if err != nil {
			log.Printf("config change error (list store value): %v", err.Error())
		}

		id := val.(string)

		// now that we've found the unique ID for this TX definition, we can
		// proceed to find the actual TX definition
		for i := range *ws.TX {
			if (*ws.TX)[i].ID != id && id != "" {
				continue
			}

			// now we've found the actual TX definition and we can
			// make changes to the TX and propagate it to the ListStore
			switch column {
			// case constants.COLUMN_ORDER:
			// TODO: validate that the newValue is of int type
			// before unsafe type assertion
			// nv, err := strconv.ParseInt(newValue.(string), 10, 64)
			// if err != nil {
			// 	log.Printf(
			// 		"failed to convert interval %v to int: %v",
			// 		newValue,
			// 		err.Error(),
			// 	)
			// }
			// nvi := int(nv)
			// if nvi <= 0 {
			// 	nvi = 1
			// }
			// (*ws.TX)[i].Order = nvi
			case constants.COLUMN_AMOUNT:
				nv := int(lib.ParseDollarAmount(newValue.(string), false))
				(*ws.TX)[i].Amount = nv
			case constants.COLUMN_ACTIVE:
				nv := !(newValue.(bool))
				(*ws.TX)[i].Active = nv
			case constants.COLUMN_NAME:
				nv := newValue.(string)
				(*ws.TX)[i].Name = nv
			case constants.COLUMN_FREQUENCY:
				// TODO: refactor for unit testability
				nv := strings.ToUpper(strings.TrimSpace(newValue.(string)))
				if nv == constants.Y {
					nv = constants.YEARLY
				}
				if nv == constants.W {
					nv = constants.WEEKLY
				}
				if nv == constants.M {
					nv = constants.MONTHLY
				}
				if nv != constants.WEEKLY && nv != constants.MONTHLY && nv != constants.YEARLY {
					(*ws.ShowMessageDialog)(constants.MsgInvalidRecurrence, gtk.MESSAGE_ERROR)
					break
				}

				(*ws.TX)[i].Frequency = nv
			case constants.COLUMN_INTERVAL:
				nv, err := strconv.ParseInt(newValue.(string), 10, 64)
				if err != nil {
					(*ws.ShowMessageDialog)(fmt.Sprintf(
						"failed to convert interval %v to int: %v",
						newValue,
						err.Error(),
					), gtk.MESSAGE_ERROR)
				}
				if nv <= 0 {
					nv = 1
				}
				(*ws.TX)[i].Interval = int(nv)
			case constants.COLUMN_STARTS:
				nvs := newValue.(string)
				yr, mo, day := lib.ParseYearMonthDateString(
					strings.TrimSpace(nvs),
				)
				(*ws.TX)[i].StartsYear = yr
				(*ws.TX)[i].StartsMonth = mo
				(*ws.TX)[i].StartsDay = day
			case constants.COLUMN_ENDS:
				// TODO: refactor similar code from above case
				nvs := newValue.(string)
				yr, mo, day := lib.ParseYearMonthDateString(
					strings.TrimSpace(nvs),
				)
				(*ws.TX)[i].EndsYear = yr
				(*ws.TX)[i].EndsMonth = mo
				(*ws.TX)[i].EndsDay = day
			case constants.COLUMN_NOTE:
				nv := newValue.(string)
				(*ws.TX)[i].Note = nv
			default:
				if oldutil.IsWeekday(constants.ConfigColumns[column]) {
					weekday := oldutil.WeekdayIndex[constants.ConfigColumns[column]]
					(*ws.TX)[i].Weekdays[weekday] = !(*ws.TX)[i].Weekdays[weekday]
					break
				}

				log.Printf(
					"warning: column id %v was modified, but there is no case to handle it",
					column,
				)

			}

			(*ws.TX)[i].UpdatedAt = time.Now()

			cells, columns := GetTXAsRow(&(*ws.TX)[i])
			ws.ConfigListStore.Set(iter, columns, cells)
		}

		return true
	}

	ws.ConfigListStore.ForEach(iterFn)

	// err := lib.ValidateTransactions(ws.TX)
	// if err != nil {
	// 	log.Printf("validation warning: %v", err.Error())
	// 	// since there was a validation error, we have to update every single
	// 	// item in the store
	// 	err = SyncConfigListStore(ws)
	// 	if err != nil {
	// 		// an error occurred during the list store config after the user
	// 		// made a change.
	// 		(*ws.ShowMessageDialog)(fmt.Sprintf(
	// 			"Error code %v - failed to sync after a config change: %v",
	// 			constants.ErrorCodeSyncConfigListStore,
	// 			err.Error(),
	// 		), gtk.MESSAGE_ERROR)
	// 	}
	// }

	UpdateResults(ws, false)
}

// SetConfigScrollPosition saves the current config scrolled window's vertical
// and horizontal scrollbar positions, so that they can be recalled later. This
// is typically followed by RestoreConfigScrollPosition. To leave a value
// unchanged, provide a value of -1.
func SetConfigScrollPosition(ws *state.WinState, vert float64, horiz float64) {
	if ws.ConfigScrolledWindow == nil {
		return
	}

	if vert == -1 {
		ws.ConfigVScroll = ws.ConfigScrolledWindow.GetVScrollbar().GetValue()
	} else {
		ws.ConfigVScroll = vert
	}

	if vert == -1 {
		ws.ConfigHScroll = ws.ConfigScrolledWindow.GetHScrollbar().GetValue()
	} else {
		ws.ConfigVScroll = horiz
	}
}

// RestoreConfigScrollPosition is a workaround.
// This allows the scrollbar to be manipulated properly; without it,
// some sort of rendering race condition occurs, and we end up
// either not setting the scroll at all, or if we set it for > 30ms,
// there is a perceivable jump from the top of the ScrolledWindow to
// the previous ScrolledWindow scroll position.
//
// Note that setting the wait time to be too short may nullify the effect
// because on low-power devices, it may take longer than e.g. 30ms to
// re-render.
//
// Additionally, it is important to point out that we are only doing
// this if a pointer to the scrolled window exists, otherwise calls
// to ws.Win.ShowAll() will cause the window to be resized according to
// the size of the window when there isn't a gtk.ScrolledWindow yet, in
// addition to nil pointer references.
func RestoreConfigScrollPosition(ws *state.WinState) {
	if ws.ConfigScrolledWindow == nil {
		return
	}

	// recall the scrollbar position
	go func() {
		time.Sleep(200 * time.Millisecond)

		if ws.ConfigScrolledWindow == nil {
			return
		}

		// The following is a safe way of doing this one-liner:
		// ws.ConfigScrolledWindow.GetVScrollbar().GetAdjustment().SetValue(ws.ConfigVScroll)

		v := ws.ConfigScrolledWindow.GetVScrollbar()
		if v != nil {
			adj := v.GetAdjustment()

			if adj != nil {
				adj.SetValue(ws.ConfigVScroll)
			}
		}

		// Again, the following is a safe way of doing this one-liner:
		// ws.ConfigScrolledWindow.GetHScrollbar().GetAdjustment().SetValue(ws.ConfigHScroll)

		h := ws.ConfigScrolledWindow.GetHScrollbar()
		if h != nil {
			adj := h.GetAdjustment()

			if adj != nil {
				adj.SetValue(ws.ConfigHScroll)
			}
		}

		ws.Win.ShowAll()

		// reset the values
		ws.ConfigVScroll = float64(0)
		ws.ConfigHScroll = float64(0)
	}()
}

func SetConfigSortColumn(ws *state.WinState, column int) {
	columnStr := constants.ConfigColumns[column]
	next := lib.GetNextSort(ws.ConfigColumnSort, columnStr)

	ws.ConfigColumnSort = next
	err := SyncConfigListStore(ws)
	if err != nil {
		(*ws.ShowMessageDialog)(fmt.Sprintf(
			"Error code %v - failed to sync after a column sort order change: %v",
			constants.ErrorCodeSyncConfigListStoreAfterColumnSortChange,
			err.Error(),
		), gtk.MESSAGE_ERROR)
	}

	ClearAllSelections(ws)

	UpdateResults(ws, false)
}

// getOrderColumn builds out an "Order" column, which is an integer column
// that allows the user to control an unsorted default order of recurring
// transactions in the config view
// func getOrderColumn(ws *state.WinState) (tvc *gtk.TreeViewColumn, err error) {
// 	orderCellRenderer, err := gtk.CellRendererTextNew()
// 	if err != nil {
// 		return tvc, fmt.Errorf("unable to create amount column renderer: %v", err.Error())
// 	}

// 	orderColumn, err := gtk.TreeViewColumnNewWithAttribute(
// 		constants.ColumnOrder,
// 		orderCellRenderer,
// 		"text",
// 		constants.COLUMN_ORDER,
// 	)
// 	if err != nil {
// 		return tvc, fmt.Errorf(
// 			"unable to create amount cell column: %v",
// 			err.Error(),
// 		)
// 	}

// 	orderColumnBtn, err := orderColumn.GetButton()
// 	if err != nil {
// 		log.Printf(
// 			"failed to get active column header button: %v",
// 			err.Error(),
// 		)
// 	}

// 	orderCellEditingStarted := func(
// 		a *gtk.CellRendererText,
// 		e *gtk.CellEditable,
// 		path string,
// 	) {
// 		log.Println(constants.GtkSignalEditingStart, a, path)
// 	}

// 	orderCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
// 		ConfigChange(ws, path, constants.COLUMN_ORDER, newText)
// 	}
// 	orderCellRenderer.SetProperty("editable", true)
// 	orderCellRenderer.SetVisible(true)

// 	orderCellRenderer.Connect(constants.GtkSignalEditingStart, orderCellEditingStarted)
// 	orderCellRenderer.Connect(constants.GtkSignalEdited, orderCellEditingFinished)

// 	orderColumn.SetResizable(true)
// 	orderColumn.SetClickable(true)
// 	orderColumn.SetVisible(true)

// 	orderColumnBtn.ToWidget().Connect(constants.GtkSignalClicked, func() {
// 		SetConfigSortColumn(ws, constants.COLUMN_ORDER)
// 	})

// 	return orderColumn, nil
// }

// getAmountColumn builds out an "Amount" column, which is an integer column
// that allows the user to input a positive or negative cash amount and it will
// be parsed into currency and displayed as currency as well
func getAmountColumn(ws *state.WinState) (tvc *gtk.TreeViewColumn, err error) {
	amtCellEditingStarted := func(a *gtk.CellRendererText, e *gtk.CellEditable, path string) {
		// log.Println(constants.GtkSignalEditingStart, a, path)
	}
	amtCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
		ConfigChange(ws, path, constants.COLUMN_AMOUNT, newText)
	}
	amtCellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tvc, fmt.Errorf("unable to create amount column renderer: %v", err.Error())
	}
	amtCellRenderer.SetProperty("editable", true)
	amtCellRenderer.SetVisible(true)
	amtCellRenderer.Connect(constants.GtkSignalEditingStart, amtCellEditingStarted)
	amtCellRenderer.Connect(constants.GtkSignalEdited, amtCellEditingFinished)
	amtColumn, err := gtk.TreeViewColumnNewWithAttribute(constants.ColumnAmount, amtCellRenderer, "text", constants.COLUMN_AMOUNT)
	// amtColumn, err := gtk.TreeViewColumnNewWithAttribute(c.ColumnAmount, amtCellRenderer, "markup", c.COLUMN_AMOUNT)
	if err != nil {
		return tvc, fmt.Errorf(
			"unable to create amount cell column: %v",
			err.Error(),
		)
	}
	amtColumn.SetResizable(true)
	amtColumn.SetClickable(true)
	amtColumn.SetVisible(true)

	amtColumnBtn, err := amtColumn.GetButton()
	if err != nil {
		log.Printf("failed to get active column header button: %v", err.Error())
	}
	amtColumnBtn.ToWidget().Connect(constants.GtkSignalClicked, func() {
		SetConfigSortColumn(ws, constants.COLUMN_AMOUNT)
	})

	return amtColumn, nil
}

// getActiveColumn builds out an "Active" column, which is a boolean column
// that allows the user to enable/disable a recurring transaction
func getActiveColumn(ws *state.WinState) (tvc *gtk.TreeViewColumn, err error) {
	activeColumn, err := createCheckboxColumn(
		constants.ColumnActive,
		constants.COLUMN_ACTIVE,
		false,
		ws.ConfigListStore,
		ws,
	)
	if err != nil {
		return tvc, fmt.Errorf(
			"failed to create checkbox column '%v': %v",
			constants.ColumnActive,
			err.Error(),
		)
	}
	activeColumnHeaderBtn, err := activeColumn.GetButton()
	if err != nil {
		log.Printf("failed to get active column header button: %v", err.Error())
	}
	activeColumnHeaderBtn.ToWidget().Connect(constants.GtkSignalClicked, func() {
		SetConfigSortColumn(ws, constants.COLUMN_ACTIVE)
	})

	return activeColumn, nil
}

// getNameColumn builds out a "Name" column, which is a string column that
// allows users to set the name of a recurring transaction
func getNameColumn(ws *state.WinState) (tvc *gtk.TreeViewColumn, err error) {
	nameCellEditingStarted := func(a *gtk.CellRendererText, e *gtk.CellEditable, path string) {
		// log.Println(constants.GtkSignalEditingStart, a, path)
	}
	nameCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
		ConfigChange(ws, path, constants.COLUMN_NAME, newText)
	}
	nameCellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tvc, fmt.Errorf("unable to create Name column renderer: %v", err.Error())
	}
	nameCellRenderer.SetProperty("editable", true)
	nameCellRenderer.SetVisible(true)
	nameCellRenderer.Connect(constants.GtkSignalEditingStart, nameCellEditingStarted)
	nameCellRenderer.Connect(constants.GtkSignalEdited, nameCellEditingFinished)
	nameColumn, err := gtk.TreeViewColumnNewWithAttribute(constants.ColumnName, nameCellRenderer, "text", constants.COLUMN_NAME)
	// nameColumn, err := gtk.TreeViewColumnNewWithAttribute(c.ColumnName, nameCellRenderer, "markup", c.COLUMN_NAME)
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
	nameColumnHeaderBtn.ToWidget().Connect(constants.GtkSignalClicked, func() {
		SetConfigSortColumn(ws, constants.COLUMN_NAME)
	})

	return nameColumn, nil
}

// getFrequencyColumn builds out a "Frequency" column, which is a string
// column that allows the user to specify monthly/weekly/yearly recurrences
// TODO: make this form of input more user friendly - currently it requires
// users to just simply type "MONTHLY"/"YEARLY"/"WEEKLY"
func getFrequencyColumn(ws *state.WinState) (tvc *gtk.TreeViewColumn, err error) {
	freqCellEditingStarted := func(a *gtk.CellRendererText, e *gtk.CellEditable, path string) {
		// log.Println(constants.GtkSignalEditingStart, a, path)
	}
	freqCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
		ConfigChange(ws, path, constants.COLUMN_FREQUENCY, newText)
	}
	freqCellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tvc, fmt.Errorf("unable to create Frequency column renderer: %v", err.Error())
	}
	freqCellRenderer.SetProperty("editable", true)
	freqCellRenderer.SetVisible(true)
	freqCellRenderer.Connect(constants.GtkSignalEditingStart, freqCellEditingStarted)
	freqCellRenderer.Connect(constants.GtkSignalEdited, freqCellEditingFinished)
	freqColumn, err := gtk.TreeViewColumnNewWithAttribute(constants.ColumnFrequency, freqCellRenderer, "text", constants.COLUMN_FREQUENCY)
	// freqColumn, err := gtk.TreeViewColumnNewWithAttribute(c.ColumnFrequency, freqCellRenderer, "markup", c.COLUMN_FREQUENCY)
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
	freqColumnBtn.ToWidget().Connect(constants.GtkSignalClicked, func() {
		SetConfigSortColumn(ws, constants.COLUMN_FREQUENCY)
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
		// log.Println(constants.GtkSignalEditingStart, a, path)
	}
	intervalCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
		ConfigChange(ws, path, constants.COLUMN_INTERVAL, newText)
	}
	intervalCellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tvc, fmt.Errorf("unable to create Interval column renderer: %v", err.Error())
	}
	intervalCellRenderer.SetProperty("editable", true)
	intervalCellRenderer.SetVisible(true)
	intervalCellRenderer.Connect(constants.GtkSignalEditingStart, intervalCellEditingStarted)
	intervalCellRenderer.Connect(constants.GtkSignalEdited, intervalCellEditingFinished)
	intervalColumn, err := gtk.TreeViewColumnNewWithAttribute(constants.ColumnInterval, intervalCellRenderer, "text", constants.COLUMN_INTERVAL)
	// intervalColumn, err := gtk.TreeViewColumnNewWithAttribute(c.ColumnInterval, intervalCellRenderer, "markup", c.COLUMN_INTERVAL)
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
	intervalColumnBtn.ToWidget().Connect(constants.GtkSignalClicked, func() {
		SetConfigSortColumn(ws, constants.COLUMN_INTERVAL)
	})

	return intervalColumn, nil
}

// getStartsColumn builds out a "Starts" column, which is a string column that
// allows the user to type in a starting date, such as 2020-02-01.
// TODO: refactoring and cleanup
func getStartsColumn(ws *state.WinState) (tvc *gtk.TreeViewColumn, err error) {
	startsCellEditingStarted := func(a *gtk.CellRendererText, e *gtk.CellEditable, path string) {
		// log.Println(constants.GtkSignalEditingStart, a, path)
	}
	startsCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
		ConfigChange(ws, path, constants.COLUMN_STARTS, newText)
	}
	startsCellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tvc, fmt.Errorf("unable to create Starts column renderer: %v", err.Error())
	}
	startsCellRenderer.SetProperty("editable", true)
	startsCellRenderer.SetVisible(true)
	startsCellRenderer.Connect(constants.GtkSignalEditingStart, startsCellEditingStarted)
	startsCellRenderer.Connect(constants.GtkSignalEdited, startsCellEditingFinished)
	startsColumn, err := gtk.TreeViewColumnNewWithAttribute(constants.ColumnStarts, startsCellRenderer, "text", constants.COLUMN_STARTS)
	// startsColumn, err := gtk.TreeViewColumnNewWithAttribute(c.ColumnStarts, startsCellRenderer, "markup", c.COLUMN_STARTS)
	if err != nil {
		return tvc, fmt.Errorf("unable to create Starts cell column: %v", err.Error())
	}
	startsColumn.SetResizable(true)
	startsColumn.SetClickable(true)
	startsColumn.SetVisible(true)
	startsColumnHeaderBtn, err := startsColumn.GetButton()
	if err != nil {
		log.Printf("failed to get Starts column header button: %v", err.Error())
	}
	startsColumnHeaderBtn.ToWidget().Connect(constants.GtkSignalClicked, func() {
		SetConfigSortColumn(ws, constants.COLUMN_STARTS)
	})

	return startsColumn, nil
}

// getEndsColumn builds out an "Ends" column, which is a string column that
// allows the user to type in an ending date, such as 2020-02-01.
// TODO: refactoring and cleanup
func getEndsColumn(ws *state.WinState) (tvc *gtk.TreeViewColumn, err error) {
	endsCellEditingStarted := func(a *gtk.CellRendererText, e *gtk.CellEditable, path string) {
		// log.Println(constants.GtkSignalEditingStart, a, path)
	}
	endsCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
		ConfigChange(ws, path, constants.COLUMN_ENDS, newText)
	}
	endsCellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tvc, fmt.Errorf("unable to create Ends column renderer: %v", err.Error())
	}
	endsCellRenderer.SetProperty("editable", true)
	endsCellRenderer.SetVisible(true)
	endsCellRenderer.Connect(constants.GtkSignalEditingStart, endsCellEditingStarted)
	endsCellRenderer.Connect(constants.GtkSignalEdited, endsCellEditingFinished)
	endsColumn, err := gtk.TreeViewColumnNewWithAttribute(constants.ColumnEnds, endsCellRenderer, "text", constants.COLUMN_ENDS)
	// endsColumn, err := gtk.TreeViewColumnNewWithAttribute(c.ColumnEnds, endsCellRenderer, "markup", c.COLUMN_ENDS)
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
	endsColumnHeaderBtn.ToWidget().Connect(constants.GtkSignalClicked, func() {
		SetConfigSortColumn(ws, constants.COLUMN_ENDS)
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
		// log.Println(constants.GtkSignalEditingStart, a, path)
	}
	notesCellEditingFinished := func(a *gtk.CellRendererText, path string, newText string) {
		ConfigChange(ws, path, constants.COLUMN_NOTE, newText)
	}
	notesCellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return tvc, fmt.Errorf("unable to create Notes column renderer: %v", err.Error())
	}
	notesCellRenderer.SetProperty("editable", true)
	notesCellRenderer.SetVisible(true)
	notesCellRenderer.Connect(constants.GtkSignalEditingStart, notesCellEditingStarted)
	notesCellRenderer.Connect(constants.GtkSignalEdited, notesCellEditingFinished)
	notesColumn, err := gtk.TreeViewColumnNewWithAttribute(constants.ColumnNote, notesCellRenderer, "text", constants.COLUMN_NOTE)
	// notesColumn, err := gtk.TreeViewColumnNewWithAttribute(c.ColumnNote, notesCellRenderer, "markup", c.COLUMN_NOTE)
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
	notesColumnHeaderBtn.ToWidget().Connect(constants.GtkSignalClicked, func() {
		SetConfigSortColumn(ws, constants.COLUMN_NOTE)
	})

	return notesColumn, nil
}

func getReadOnlyColumn(ws *state.WinState, name string, id int) (tvc *gtk.TreeViewColumn, err error) {
	rend, err := gtk.CellRendererTextNew()
	if err != nil {
		return tvc, fmt.Errorf(
			"unable to create %v column renderer: %v",
			name,
			err.Error(),
		)
	}
	rend.SetProperty("editable", false)
	rend.SetVisible(true)
	col, err := gtk.TreeViewColumnNewWithAttribute(name, rend, "text", id)
	if err != nil {
		return tvc, fmt.Errorf(
			"unable to create %v cell column: %v",
			name,
			err.Error(),
		)
	}
	col.SetResizable(true)
	col.SetClickable(true)
	col.SetVisible(true)
	header, err := col.GetButton()
	if err != nil {
		log.Printf("failed to get %v column header button: %v", name, err.Error())
	}

	header.ToWidget().Connect(constants.GtkSignalClicked, func() { SetConfigSortColumn(ws, id) })

	return col, nil
}

// Creates a tree view and the list store that holds its data
func setupConfigTreeView(ws *state.WinState) (tv *gtk.TreeView, err error) {
	treeView, err := gtk.TreeViewNew()
	if err != nil {
		return tv, fmt.Errorf("unable to create config tree view: %v", err.Error())
	}

	// orderColumn, err := getOrderColumn(ws)
	// if err != nil {
	// 	return tv, fmt.Errorf("failed to create config order column: %v", err.Error())
	// }
	// treeView.AppendColumn(orderColumn)

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
	for i := range constants.Weekdays {
		// prevents pointers from changing which column is referred to
		weekdayIndex := int(i)
		weekday := string(constants.Weekdays[weekdayIndex])

		// offset by 4 previous columns
		weekdayColumn, err := createCheckboxColumn(
			weekday,
			constants.COLUMN_MONDAY+weekdayIndex,
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
		weekdayColumnBtn.ToWidget().Connect(constants.GtkSignalClicked, func(g *gtk.Button) {
			SetConfigSortColumn(ws, constants.COLUMN_MONDAY+weekdayIndex)
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

	idColumn, err := getReadOnlyColumn(ws, constants.ColumnID, constants.COLUMN_ID)
	if err != nil {
		return tv, fmt.Errorf("failed to create config notes column: %v", err.Error())
	}
	treeView.AppendColumn(idColumn)

	createdColumn, err := getReadOnlyColumn(ws, constants.ColumnCreatedAt, constants.COLUMN_CREATEDAT)
	if err != nil {
		return tv, fmt.Errorf("failed to create config notes column: %v", err.Error())
	}
	treeView.AppendColumn(createdColumn)

	updatedColumn, err := getReadOnlyColumn(ws, constants.ColumnUpdatedAt, constants.COLUMN_UPDATEDAT)
	if err != nil {
		return tv, fmt.Errorf("failed to create config notes column: %v", err.Error())
	}
	treeView.AppendColumn(updatedColumn)

	treeView.SetModel(ws.ConfigListStore)

	return treeView, nil
}

// GetNewConfigListStore creates a list store. This is what holds the data
// that will be shown on our config tree view
func GetNewConfigListStore() (ls *gtk.ListStore, err error) {
	ls, err = gtk.ListStoreNew(
		// glib.TYPE_INT,
		glib.TYPE_STRING,  // COLUMN_AMOUNT
		glib.TYPE_BOOLEAN, // COLUMN_ACTIVE
		glib.TYPE_STRING,  // COLUMN_NAME
		glib.TYPE_STRING,  // COLUMN_FREQUENCY
		glib.TYPE_STRING,  // COLUMN_INTERVAL
		glib.TYPE_BOOLEAN, // COLUMN_MONDAY
		glib.TYPE_BOOLEAN, // COLUMN_TUESDAY
		glib.TYPE_BOOLEAN, // COLUMN_WEDNESDAY
		glib.TYPE_BOOLEAN, // COLUMN_THURSDAY
		glib.TYPE_BOOLEAN, // COLUMN_FRIDAY
		glib.TYPE_BOOLEAN, // COLUMN_SATURDAY
		glib.TYPE_BOOLEAN, // COLUMN_SUNDAY
		glib.TYPE_STRING,  // COLUMN_STARTS
		glib.TYPE_STRING,  // COLUMN_ENDS
		glib.TYPE_STRING,  // COLUMN_NOTE
		glib.TYPE_STRING,  // COLUMN_ID
		glib.TYPE_STRING,  // COLUMN_CREATEDAT
		glib.TYPE_STRING,  // COLUMN_UPDATEDAT
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
			tj := (*ws.TX)[j]
			ti := (*ws.TX)[i]

			switch ws.ConfigColumnSort {

			// invisible order column (default when no sort is set)
			case constants.None:
				// return tj.Order > ti.Order
				return tj.CreatedAt.After(ti.CreatedAt)

			// Order
			// case fmt.Sprintf("%v%v", constants.ColumnOrder, constants.Asc):
			// 	return tj.Order > ti.Order
			// case fmt.Sprintf("%v%v", constants.ColumnOrder, constants.Desc):
			// 	return ti.Order > tj.Order

			// active
			case fmt.Sprintf("%v%v", constants.ColumnActive, constants.Asc):
				return tj.Active
			case fmt.Sprintf("%v%v", constants.ColumnActive, constants.Desc):
				return ti.Active

			// weekdays
			case fmt.Sprintf("%v%v", constants.WeekdayMonday, constants.Asc):
				return tj.Weekdays[constants.WeekdayMondayInt]
			case fmt.Sprintf("%v%v", constants.WeekdayMonday, constants.Desc):
				return ti.Weekdays[constants.WeekdayMondayInt]
			case fmt.Sprintf("%v%v", constants.WeekdayTuesday, constants.Asc):
				return tj.Weekdays[constants.WeekdayTuesdayInt]
			case fmt.Sprintf("%v%v", constants.WeekdayTuesday, constants.Desc):
				return ti.Weekdays[constants.WeekdayTuesdayInt]
			case fmt.Sprintf("%v%v", constants.WeekdayWednesday, constants.Asc):
				return tj.Weekdays[constants.WeekdayWednesdayInt]
			case fmt.Sprintf("%v%v", constants.WeekdayWednesday, constants.Desc):
				return ti.Weekdays[constants.WeekdayWednesdayInt]
			case fmt.Sprintf("%v%v", constants.WeekdayThursday, constants.Asc):
				return tj.Weekdays[constants.WeekdayThursdayInt]
			case fmt.Sprintf("%v%v", constants.WeekdayThursday, constants.Desc):
				return ti.Weekdays[constants.WeekdayThursdayInt]
			case fmt.Sprintf("%v%v", constants.WeekdayFriday, constants.Asc):
				return tj.Weekdays[constants.WeekdayFridayInt]
			case fmt.Sprintf("%v%v", constants.WeekdayFriday, constants.Desc):
				return ti.Weekdays[constants.WeekdayFridayInt]
			case fmt.Sprintf("%v%v", constants.WeekdaySaturday, constants.Asc):
				return tj.Weekdays[constants.WeekdaySaturdayInt]
			case fmt.Sprintf("%v%v", constants.WeekdaySaturday, constants.Desc):
				return ti.Weekdays[constants.WeekdaySaturdayInt]
			case fmt.Sprintf("%v%v", constants.WeekdaySunday, constants.Asc):
				return tj.Weekdays[constants.WeekdaySundayInt]
			case fmt.Sprintf("%v%v", constants.WeekdaySunday, constants.Desc):
				return ti.Weekdays[constants.WeekdaySundayInt]

			// other columns
			case fmt.Sprintf("%v%v", constants.ColumnAmount, constants.Asc):
				return tj.Amount > ti.Amount
			case fmt.Sprintf("%v%v", constants.ColumnAmount, constants.Desc):
				return ti.Amount > tj.Amount

			case fmt.Sprintf("%v%v", constants.ColumnFrequency, constants.Asc):
				return tj.Frequency > ti.Frequency
			case fmt.Sprintf("%v%v", constants.ColumnFrequency, constants.Desc):
				return ti.Frequency > tj.Frequency

			case fmt.Sprintf("%v%v", constants.ColumnInterval, constants.Asc):
				return tj.Interval > ti.Interval
			case fmt.Sprintf("%v%v", constants.ColumnInterval, constants.Desc):
				return ti.Interval > tj.Interval
			case fmt.Sprintf("%v%v", constants.ColumnNote, constants.Asc):

				return strings.ToLower(tj.Note) > strings.ToLower(ti.Note)
			case fmt.Sprintf("%v%v", constants.ColumnNote, constants.Desc):
				return strings.ToLower(ti.Note) > strings.ToLower(tj.Note)

			case fmt.Sprintf("%v%v", constants.ColumnName, constants.Asc):
				return strings.ToLower(tj.Name) > strings.ToLower(ti.Name)
			case fmt.Sprintf("%v%v", constants.ColumnName, constants.Desc):
				return strings.ToLower(ti.Name) > strings.ToLower(tj.Name)

			case fmt.Sprintf("%v%v", constants.ColumnID, constants.Asc):
				return strings.ToLower(tj.ID) > strings.ToLower(ti.ID)
			case fmt.Sprintf("%v%v", constants.ColumnID, constants.Desc):
				return strings.ToLower(ti.ID) > strings.ToLower(tj.ID)

			case fmt.Sprintf("%v%v", constants.ColumnCreatedAt, constants.Asc):
				return tj.CreatedAt.After(ti.CreatedAt)
			case fmt.Sprintf("%v%v", constants.ColumnCreatedAt, constants.Desc):
				return ti.CreatedAt.Before(tj.CreatedAt)

			case fmt.Sprintf("%v%v", constants.ColumnUpdatedAt, constants.Asc):
				return tj.UpdatedAt.After(ti.UpdatedAt)
			case fmt.Sprintf("%v%v", constants.ColumnUpdatedAt, constants.Desc):
				return ti.UpdatedAt.Before(tj.UpdatedAt)

			case fmt.Sprintf("%v%v", constants.ColumnStarts, constants.Asc):
				ist := fmt.Sprintf("%v-%v-%v", tj.StartsYear, tj.StartsMonth, tj.StartsDay)
				jst := fmt.Sprintf("%v-%v-%v", ti.StartsYear, ti.StartsMonth, ti.StartsDay)
				return ist > jst
			case fmt.Sprintf("%v%v", constants.ColumnStarts, constants.Desc):
				ist := fmt.Sprintf("%v-%v-%v", tj.StartsYear, tj.StartsMonth, tj.StartsDay)
				jst := fmt.Sprintf("%v-%v-%v", ti.StartsYear, ti.StartsMonth, ti.StartsDay)
				return jst > ist

			case fmt.Sprintf("%v%v", constants.ColumnEnds, constants.Asc):
				jend := fmt.Sprintf("%v-%v-%v", tj.EndsYear, tj.EndsMonth, tj.EndsDay)
				iend := fmt.Sprintf("%v-%v-%v", ti.EndsYear, ti.EndsMonth, ti.EndsDay)
				return jend > iend
			case fmt.Sprintf("%v%v", constants.ColumnEnds, constants.Desc):
				jend := fmt.Sprintf("%v-%v-%v", tj.EndsYear, tj.EndsMonth, tj.EndsDay)
				iend := fmt.Sprintf("%v-%v-%v", ti.EndsYear, ti.EndsMonth, ti.EndsDay)
				return iend > jend

			default:
				return false
				// return txs[j].Date.After(txs[i].Date)
			}
		},
	)

	SetConfigScrollPosition(ws, -1, -1)

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

	RestoreConfigScrollPosition(ws)

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

	hideInactiveCheckbox, err := gtk.CheckButtonNewWithMnemonic(constants.HideInactiveBtnLabel)
	if err != nil {
		log.Fatal("failed to create 'hide inactive' checkbox:", err)
	}

	SetSpacerMarginsGtkCheckBtn(hideInactiveCheckbox)

	hideInactiveCheckbox.SetActive(ws.HideInactive)

	hideInactiveCheckbox.Connect(constants.GtkSignalClicked, hideInactiveCheckBoxClickedHandler)

	return hideInactiveCheckbox
}

func GetConfEditButtons(ws *state.WinState) (*gtk.Button, *gtk.Button, *gtk.Button) {
	addConfItemBtn, err := gtk.ButtonNewWithMnemonic(constants.AddBtnLabel)
	if err != nil {
		log.Fatal("failed to create add conf item button:", err)
	}
	SetSpacerMarginsGtkBtn(addConfItemBtn)

	delConfItemBtn, err := gtk.ButtonNewWithMnemonic(constants.DelBtnLabel)
	if err != nil {
		log.Fatal("failed to create add conf item button:", err)
	}
	SetSpacerMarginsGtkBtn(delConfItemBtn)

	cloneConfItemBtn, err := gtk.ButtonNewWithMnemonic(constants.CloneBtnLabel)
	if err != nil {
		log.Fatal("failed to create clone conf item button:", err)
	}
	SetSpacerMarginsGtkBtn(cloneConfItemBtn)

	return addConfItemBtn, delConfItemBtn, cloneConfItemBtn
}

func GetConfigTab(ws *state.WinState) (*gtk.ScrolledWindow, *gtk.TreeView, *gtk.Label) {
	// TODO: refactor the config tree view list store using the same
	// approach that was used for the results list store
	configTreeView, err := GetConfigAsTreeView(ws)
	if err != nil {
		log.Fatalf("failed to get config as tree view: %v", err.Error())
	}

	configTreeView.SetEnableSearch(false)
	configTab, err := gtk.LabelNew(constants.ConfigTabLabel)
	if err != nil {
		log.Fatalf("failed to set config tab label: %v", err.Error())
	}

	configTreeSelection, err := configTreeView.GetSelection()
	if err != nil {
		log.Fatalf("failed to get results tree view sel: %v", err.Error())
	}

	configTreeSelection.SetMode(gtk.SELECTION_MULTIPLE)

	selectionChanged := func(s *gtk.TreeSelection) {
		// reset the map of selected id's
		ws.SelectedConfIDs = make(map[string]bool)

		rows := s.GetSelectedRows(ws.ConfigListStore)

		for l := rows; l != nil; l = l.Next() {
			path := l.Data().(*gtk.TreePath)

			id, err := oldutil.GetTXIDByListStorePath(ws.ConfigListStore, path)
			if err != nil {
				log.Printf("failed to retrieve id for path upon selection change: %v", err.Error())
			}

			if id == "" {
				log.Printf("warning: empty id for selection path=%v", path.String())
				continue
			}

			ws.SelectedConfIDs[id] = true
		}
	}

	configTreeSelection.Connect(constants.GtkSignalChanged, selectionChanged)
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

	return configSw, configTreeView, configTab
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
	p.Connect(constants.ActionClose, func() { p.Close() })
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
			}

			*ws.TX, err = oldutil.LoadConfig(ws.OpenFileName)
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
				newTX := lib.GetNewTX(time.Now())
				// newTX.Order = 1
				*ws.TX = []lib.TX{newTX}
			}

			SyncConfigListStore(ws)
			UpdateResults(ws, false)

			ws.Win.ShowAll()
			ws.Notebook.SetCurrentPage(constants.TAB_CONFIG)
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

		p.Close()
		p.Destroy()
	})
	p.Dialog.ShowAll()
}

// GetConfigBaseComponents returns the base components that are required in
// order to view the transaction tree view presented inside of a GTK notebook
// tab.
func GetConfigBaseComponents(ws *state.WinState) (*gtk.Grid, *gtk.ScrolledWindow, *gtk.TreeView, *gtk.Label) {
	configGrid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("failed to create config grid", err)
	}

	configGrid.SetOrientation(gtk.ORIENTATION_VERTICAL)
	configSw, configTreeView, configTab := GetConfigTab(ws)
	configGrid.Attach(configSw, 0, 0, constants.FullGridWidth, 2)

	return configGrid, configSw, configTreeView, configTab
}
