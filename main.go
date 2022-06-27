package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	c "finance-planner/constants"
	"finance-planner/lib"
	"finance-planner/ui"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// important at some point:
// https://github.com/gotk3/gotk3-examples/blob/master/gtk-examples/goroutines/goroutines.go
// https://github.com/gotk3/gotk3-examples/blob/master/gtk-examples/statusicon/main.go
// https://github.com/gotk3/gotk3-examples/blob/master/gtk-examples/textview/textview.go
// https://golangdocs.com/aes-encryption-decryption-in-golang

const (
	TAB_CONFIG = iota
	TAB_RESULTS
)

var (
	HideInactive    = false
	HideInactivePtr = &HideInactive
)

func main() {
	// TODO: change appID
	const appID = "com.charlesmknox.finance-planner"
	application, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatal("Could not create application:", err)
	}
	// gtk.Init(nil)

	currentUser := lib.GetUser()
	homeDir := currentUser.HomeDir
	defaultConfigFile := c.DefaultConfFileName
	if homeDir != "" {
		defaultConfigFile = fmt.Sprintf(
			"%v/%v/%v",
			homeDir,
			c.DefaultConfFilePath,
			c.DefaultConfFileName,
		)
	}

	application.Connect("activate", func() {
		win := primary(application, defaultConfigFile)

		aNew := glib.SimpleActionNew("new", nil)
		aNew.Connect("activate", func() {
			primary(application, defaultConfigFile).ShowAll()
		})
		application.AddAction(aNew)

		aQuit := glib.SimpleActionNew("quit", nil)
		aQuit.Connect("activate", func() {
			application.Quit()
		})
		application.AddAction(aQuit)

		win.ShowAll()
	})

	os.Exit(application.Run(os.Args))
}

func primary(application *gtk.Application, filename string) *gtk.ApplicationWindow {
	// for go routine access
	var (
		mainBtn                   *gtk.Button
		nSets                     = 1
		startingBalance           = 50000
		startDate                 = lib.GetNowDateString()
		endDate                   = lib.GetDefaultEndDateString()
		selectedConfigItemIndexes []int
		openFileName              = filename
		userTX                    []lib.TX
		err                       error
		latestResults             []lib.Result
	)

	// start the gtk window - not needed since we're bootstrapping the application
	// differently
	// gtk.Init(nil)

	win, err := gtk.ApplicationWindowNew(application)
	// win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("unable to create window:", err)
	}
	rootBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, c.UISpacer)
	if err != nil {
		log.Fatal("unable to create root box:", err)
	}

	win.SetTitle("Financial Planner")

	// TODO: this can be moved to elsewhere
	if openFileName != "" {
		userTX, err = lib.LoadConfig(openFileName)
		if err != nil {
			m := fmt.Sprintf(
				"Config does not exist at %v. Would you like to create a new one there now?",
				openFileName,
			)
			d := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_QUESTION, gtk.BUTTONS_YES_NO, "%s", m)
			log.Println(m)
			resp := d.Run()
			if resp == gtk.RESPONSE_YES {
				err := lib.SaveConfig(openFileName, userTX)
				if err != nil {
					m := fmt.Sprintf("Failed to save config upon window load - will proceed with a blank config. Here's the error: %v", err.Error())
					di := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "%s", m)
					log.Println(m)
					di.Run()
					di.Destroy()
				}
			}
			d.Destroy()
		}
	}

	// win.Connect("destroy", quit)

	// Create a header bar
	header, err := gtk.HeaderBarNew()
	if err != nil {
		log.Fatal("Could not create header bar:", err)
	}
	header.SetShowCloseButton(true)
	header.SetTitle("Financial Planner")
	header.SetSubtitle(openFileName)

	if len(userTX) == 0 {
		userTX = []lib.TX{{
			Order:     1,
			Amount:    -500,
			Active:    true,
			Name:      "New",
			Frequency: "WEEKLY",
			Interval:  1,
		}}
		header.SetSubtitle(fmt.Sprintf("%v*", openFileName))
		if openFileName != "" {
			m := fmt.Sprintf(
				"Success! Loaded file \"%v\" successfully. The configuration was empty, so a sample recurring transaction has been added.",
				openFileName,
			)
			d := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_OK, "%s", m)
			log.Println(m)
			d.Run()
			d.Destroy()
		}
	}

	// Create a new menu button
	mbtn, err := gtk.MenuButtonNew()
	if err != nil {
		log.Fatal("Could not create menu button:", err)
	}
	// Set up the menu model for the button
	menu := glib.MenuNew()
	if menu == nil {
		log.Fatal("Could not create menu (nil)")
	}
	// Actions with the prefix 'app' reference actions on the application
	// Actions with the prefix 'win' reference actions on the current window (specific to ApplicationWindow)
	// Other prefixes can be added to widgets via InsertActionGroup
	// menu.Append("Custom Panic", "custom.panic")
	menu.Append("Save", "fin.saveOpenConfig")
	menu.Append("Save as...", "fin.saveConfig")
	menu.Append("Open...", "fin.loadConfigCurrentWindow")
	menu.Append("Open in new window...", "fin.loadConfigNewWindow")
	menu.Append("Save results...", "fin.saveResults")
	menu.Append("Copy results to clipboard", "fin.copyResults")
	menu.Append("New Window", "app.new")
	menu.Append("Close Window", "win.close")
	menu.Append("Quit", "app.quit")

	// gtk.AccelMapGet().Connect("appquitplease", func() {
	// 	application.Quit()
	// })

	// gtk.AccelMapAddEntry(
	// 	"appquitplease",
	// 	keyQ,
	// 	modCtrl,
	// )

	// create keyboard shortcuts window some time later
	// shortcutsWindow := gtk.ShortcutsWindow{Window}
	// shortcuts := gtk.ShortcutsSection{Box: *rootBox}
	// shortcutsMainGroup := gtk.ShortcutsGroup{Box: shortcuts.Box}
	// err = shortcutsMainGroup.SetProperty("title", "Primary")
	// if err != nil {
	// 	log.Println("failed to set shortcuts main group title property", err.Error())
	// }
	// shortcutSave := gtk.ShortcutsShortcut

	// Create the action "win.close"
	aClose := glib.SimpleActionNew("close", nil)
	aClose.Connect("activate", func() {
		win.Close()
	})
	win.AddAction(aClose)

	// Create and insert custom action group with prefix "fin" (for finances)
	finActionGroup := glib.SimpleActionGroupNew()
	win.InsertActionGroup("fin", finActionGroup)

	saveConfAsFn := func() {
		p, err := gtk.FileChooserDialogNewWith2Buttons(
			"Save config",
			win,
			gtk.FILE_CHOOSER_ACTION_SAVE,
			"_Save",
			gtk.RESPONSE_OK,
			"_Cancel",
			gtk.RESPONSE_CANCEL,
		)
		if err != nil {
			log.Fatal("failed to create save config file picker", err.Error())
		}
		p.Connect("close", func() {
			p.Close()
		})
		p.Connect("response", func(dialog *gtk.FileChooserDialog, resp int) {
			if resp == int(gtk.RESPONSE_OK) {
				// folder, _ := dialog.FileChooser.GetCurrentFolder()
				// GetFilename includes the full path and file name
				openFileName = dialog.FileChooser.GetFilename()
				// write the config to the target file path
				err := lib.SaveConfig(openFileName, userTX)
				if err != nil {
					m := fmt.Sprintf("Failed to save config to file \"%v\": %v", openFileName, err.Error())
					d := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "%s", m)
					log.Println(m)
					d.Run()
					d.Destroy()
					return
				}
				header.SetSubtitle(openFileName)
			}
			p.Close()
		})
		p.Dialog.ShowAll()
	}

	saveConfAsAction := glib.SimpleActionNew("saveConfig", nil)
	saveConfAsAction.Connect("activate", saveConfAsFn)
	finActionGroup.AddAction(saveConfAsAction)
	win.AddAction(saveConfAsAction)

	saveOpenConfFn := func() {
		// write the config to the target file path
		err := lib.SaveConfig(openFileName, userTX)
		if err != nil {
			m := fmt.Sprintf("Failed to save config to file \"%v\": %v", openFileName, err.Error())
			d := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "%s", m)
			log.Println(m)
			d.Run()
			d.Destroy()
			return
		}
		header.SetSubtitle(openFileName)
	}

	saveOpenConfAction := glib.SimpleActionNew("saveOpenConfig", nil)
	saveOpenConfAction.Connect("activate", saveOpenConfFn)
	finActionGroup.AddAction(saveOpenConfAction)
	win.AddAction(saveOpenConfAction)

	saveResultsFn := func() {
		p, err := gtk.FileChooserDialogNewWith2Buttons(
			"Save config",
			win,
			gtk.FILE_CHOOSER_ACTION_SAVE,
			"_Save",
			gtk.RESPONSE_OK,
			"_Cancel",
			gtk.RESPONSE_CANCEL,
		)
		if err != nil {
			log.Fatal("failed to create save results file picker", err.Error())
		}
		p.Connect("close", func() {
			p.Close()
		})
		p.Connect("response", func(dialog *gtk.FileChooserDialog, resp int) {
			if resp == int(gtk.RESPONSE_OK) {
				// folder, _ := dialog.FileChooser.GetCurrentFolder()
				// GetFilename includes the full path and file name
				f := dialog.FileChooser.GetFilename()
				// write the config to the target file path
				err := lib.SaveResultsCSV(f, &latestResults)
				if err != nil {
					m := fmt.Sprintf("Failed to save results as CSV to file \"%v\": %v", f, err.Error())
					d := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "%s", m)
					log.Println(m)
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

	saveResultsAction := glib.SimpleActionNew("saveResults", nil)
	saveResultsAction.Connect("activate", saveResultsFn)
	finActionGroup.AddAction(saveResultsAction)
	win.AddAction(saveResultsAction)

	copyResultsFn := func() {
		r := lib.GetResultsCSVString(&latestResults)
		clipboard, err := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
		if err != nil {
			log.Print("failed to get clipboard", err.Error())
		}
		clipboard.SetText(r)
		m := fmt.Sprintf("Success! Copied %v result records to the clipboard.", len(latestResults))
		d := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_OK, "%s", m)
		log.Println(m)
		d.Run()
		d.Destroy()
	}

	copyResultsAction := glib.SimpleActionNew("copyResults", nil)
	copyResultsAction.Connect("activate", copyResultsFn)
	finActionGroup.AddAction(copyResultsAction)
	win.AddAction(copyResultsAction)

	loadConfCurrentWindowAction := glib.SimpleActionNew("loadConfigCurrentWindow", nil)
	finActionGroup.AddAction(loadConfCurrentWindowAction)
	win.AddAction(loadConfCurrentWindowAction)

	loadConfNewWindowAction := glib.SimpleActionNew("loadConfigNewWindow", nil)
	finActionGroup.AddAction(loadConfNewWindowAction)
	win.AddAction(loadConfNewWindowAction)

	mbtn.SetMenuModel(&menu.MenuModel)

	mainBtn, err = gtk.ButtonNewWithLabel("Update Results")
	if err != nil {
		log.Fatal("Unable to create button:", err)
	}
	mainBtn.SetMarginTop(c.UISpacer)
	mainBtn.SetMarginBottom(c.UISpacer)
	mainBtn.SetMarginStart(c.UISpacer)
	mainBtn.SetMarginEnd(c.UISpacer)

	// spnBtn, err := gtk.SpinButtonNewWithRange(0.0, 1.0, 0.001)
	// if err != nil {
	// 	log.Fatal("Unable to create spin button:", err)
	// }

	nb, err := gtk.NotebookNew()
	if err != nil {
		log.Fatal("Unable to create notebook:", err)
	}

	updateResults := func(switchTo bool) {
		latestResults, err = lib.GenerateResultsFromDateStrings(
			&userTX,
			startingBalance,
			startDate,
			endDate,
		)
		if err != nil {
			log.Fatal("failed to generate results from date strings", err.Error())
		}
		nb.RemovePage(TAB_RESULTS)
		newSw, newLabel, err := ui.GenerateResultsTab(&userTX, latestResults)
		if err != nil {
			log.Fatalf("failed to generate results tab: %v", err.Error())
		}
		nb.InsertPage(newSw, newLabel, TAB_RESULTS)
		header.SetSubtitle(fmt.Sprintf("%v*", openFileName))
		win.ShowAll()
		if switchTo {
			nb.SetCurrentPage(TAB_RESULTS)
		}
	}

	updateResultsDefault := func() {
		updateResults(true)
	}

	updateResultsSilent := func() {
		updateResults(false)
	}

	// text startingBalanceInput for providing the starting balance
	startingBalanceInput, err := gtk.EntryNew()
	if err != nil {
		log.Fatal("Unable to create entry:", err)
	}
	startingBalanceInput.SetMarginTop(c.UISpacer)
	startingBalanceInput.SetMarginBottom(c.UISpacer)
	startingBalanceInput.SetMarginStart(c.UISpacer)
	startingBalanceInput.SetMarginEnd(c.UISpacer)
	startingBalanceInput.SetPlaceholderText("$500.00 - Enter a balance to start with.")
	updateStartingBalance := func() {
		s, _ := startingBalanceInput.GetText()
		if s == "" {
			return
		}
		startingBalance = int(lib.ParseDollarAmount(s, true))
		startingBalanceInput.SetText(lib.FormatAsCurrency(startingBalance))
		updateResults(true)
	}
	startingBalanceInput.Connect("activate", updateStartingBalance)
	// startingBalanceInput.Connect("focus-out-event", updateStartingBalance)

	// text input for providing the starting date
	stDateInput, err := gtk.EntryNew()
	if err != nil {
		log.Fatal("Unable to create entry:", err)
	}
	stDateInput.SetMarginTop(c.UISpacer)
	stDateInput.SetMarginBottom(c.UISpacer)
	stDateInput.SetMarginStart(c.UISpacer)
	stDateInput.SetMarginEnd(c.UISpacer)
	stDateInput.SetPlaceholderText(startDate)
	stDateInputUpdate := func() {
		s, _ := stDateInput.GetText()
		y, m, d := lib.ParseYearMonthDateString(s)
		if s == "" || (y == 0 && m == 0 && d == 0) {
			return
		}
		startDate = fmt.Sprintf("%v-%v-%v", y, m, d)
		stDateInput.SetText(startDate)
		updateResults(true)
	}
	stDateInput.Connect("activate", stDateInputUpdate)
	stDateInput.Connect("focus-out-event", stDateInputUpdate)

	// text input for providing the ending date
	endDateInput, err := gtk.EntryNew()
	if err != nil {
		log.Fatal("Unable to create entry:", err)
	}
	endDateInput.SetMarginTop(c.UISpacer)
	endDateInput.SetMarginBottom(c.UISpacer)
	endDateInput.SetMarginStart(c.UISpacer)
	endDateInput.SetMarginEnd(c.UISpacer)
	endDateInput.SetPlaceholderText(endDate)
	endDateInputUpdate := func() {
		s, _ := endDateInput.GetText()
		y, m, d := lib.ParseYearMonthDateString(s)
		if s == "" || (y == 0 && m == 0 && d == 0) {
			return
		}
		endDate = fmt.Sprintf("%v-%v-%v", y, m, d)
		endDateInput.SetText(endDate)
		updateResults(true)
	}
	endDateInput.Connect("activate", endDateInputUpdate)
	endDateInput.Connect("focus-out-event", endDateInputUpdate)

	// Calling (*gtk.Container).Add() with a gtk.Grid will add widgets next
	// to each other, in the order they were added, to the right side of the
	// last added widget when the grid is in a horizontal orientation, and
	// at the bottom of the last added widget if the grid is in a vertial
	// orientation.  Using a grid in this manner works similar to a gtk.Box,
	// but unlike gtk.Box, a gtk.Grid will respect its child widget's expand
	// and margin properties.
	// grid.Add(btn)
	// grid.Add(lab)
	// grid.Add(entry)
	// grid.Add(spnBtn)

	// Widgets may also be added by calling (*gtk.Grid).Attach() to specify
	// where to place the widget in the grid, and optionally how many rows
	// and columns to span over.
	//
	// Additional rows and columns are automatically added to the grid as
	// necessary when new widgets are added with (*gtk.Container).Add(), or,
	// as shown in this case, using (*gtk.Grid).Attach().
	//
	// In this case, a notebook is added beside the widgets inserted above.
	// The notebook widget is inserted with a left position of 1, a top
	// position of 1 (starting at the same vertical position as the button),
	// a width of 1 column, and a height of 2 rows (spanning down to the
	// same vertical position as the entry).
	//
	// This example also demonstrates how not every area of the grid must
	// contain a widget.  In particular, the area to the right of the label
	// and the right of spin button have contain no widgets.
	// scrollable component view
	// const GRID_HEIGHT = 3
	// const GRID_WIDTH = 1
	// Create a new grid widget to arrange child widgets
	grid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create grid:", err)
	}

	// gtk.Grid embeds an Orientable struct to simulate the GtkOrientable
	// GInterface. Set the orientation from the default horizontal to
	// vertical.
	grid.SetOrientation(gtk.ORIENTATION_VERTICAL)

	nb.SetHExpand(true)
	nb.SetVExpand(true)

	var configViewListStore *gtk.ListStore
	// var configViewTreeView *gtk.TreeView

	// now build the config tab page
	genConfigView := func() (*gtk.ScrolledWindow, *gtk.Label) {
		configTreeView, configListStore, err := ui.GetConfigAsTreeView(&userTX, updateResultsSilent)
		if err != nil {
			log.Fatalf("failed to get config as tree view: %v", err.Error())
		}
		// so we can refer to them later
		configViewListStore = configListStore
		// configViewTreeView = configTreeView

		configTreeView.SetEnableSearch(false)
		configTab, err := gtk.LabelNew("Config")
		if err != nil {
			log.Fatalf("failed to set config tab label: %v", err.Error())
		}
		configTreeSelection, err := configTreeView.GetSelection()
		if err != nil {
			log.Fatalf("failed to get results tree vie sel: %v", err.Error())
		}
		configTreeSelection.SetMode(gtk.SELECTION_MULTIPLE)

		// Handler of "changed" signal of TreeView's selection
		selectionChanged := func(s *gtk.TreeSelection) {
			rows := s.GetSelectedRows(configListStore)
			// items := make([]string, 0, rows.Length())
			selectedConfigItemIndexes = []int{}

			for l := rows; l != nil; l = l.Next() {
				path := l.Data().(*gtk.TreePath)
				i, _ := strconv.ParseInt(path.String(), 10, 64)
				selectedConfigItemIndexes = append(selectedConfigItemIndexes, int(i))
				// iter, _ := configListStore.GetIter(path)
				// value, _ := configListStore.GetValue(iter, 0)
				// log.Println(path, iter, value)
				// str, _ := value.GetString()
				// items = append(items, str)
			}
			log.Printf("selection changed: %v, %v", s, selectedConfigItemIndexes)
		}

		configTreeSelection.Connect("changed", selectionChanged)
		configSw, err := gtk.ScrolledWindowNew(nil, nil)
		if err != nil {
			log.Fatal("Unable to create scrolled window:", err)
		}
		configSw.Add(configTreeView)
		configSw.SetHExpand(true)
		configSw.SetVExpand(true)
		configSw.SetMarginTop(c.UISpacer)
		configSw.SetMarginBottom(c.UISpacer)
		configSw.SetMarginStart(c.UISpacer)
		configSw.SetMarginEnd(c.UISpacer)

		return configSw, configTab
	}

	// this has to be defined here since it requires updateResults and genConfigView
	loadConfFn := func(openInNewWindow bool) {
		p, err := gtk.FileChooserDialogNewWith2Buttons(
			"Load config",
			win,
			gtk.FILE_CHOOSER_ACTION_OPEN,
			"_Load",
			gtk.RESPONSE_OK,
			"_Cancel",
			gtk.RESPONSE_CANCEL,
		)
		if err != nil {
			log.Fatal("failed to create load config file picker", err.Error())
		}
		p.Connect("close", func() { p.Close() })
		p.Connect("response", func(dialog *gtk.FileChooserDialog, resp int) {
			if resp == int(gtk.RESPONSE_OK) {
				// folder, _ := dialog.FileChooser.GetCurrentFolder()
				// GetFilename includes the full path and file name
				openFileName = dialog.FileChooser.GetFilename()
				if openInNewWindow {
					p.Close()
					p.Destroy()
					primary(application, openFileName).ShowAll()
					return
				} else {
					userTX, err = lib.LoadConfig(openFileName)
					if err != nil {
						m := fmt.Sprintf("Failed to load config \"%v\": %v", openFileName, err.Error())
						d := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "%s", m)
						log.Println(m)
						d.Run()
						d.Destroy()
						p.Close()
						p.Destroy()
						return
					}
					extraDialogMessageText := ""
					if len(userTX) == 0 {
						extraDialogMessageText = " The configuration was empty, so a sample recurring transaction has been added."
						userTX = []lib.TX{{
							Order:     1,
							Amount:    -500,
							Active:    true,
							Name:      "New",
							Frequency: "WEEKLY",
							Interval:  1,
						}}
					}
					nb.RemovePage(TAB_CONFIG)
					newConfigSw, newLabel := genConfigView()
					nb.InsertPage(newConfigSw, newLabel, TAB_CONFIG)
					updateResults(false)
					win.ShowAll()
					nb.SetCurrentPage(TAB_CONFIG)
					header.SetSubtitle(openFileName)
					m := fmt.Sprintf("Success! Loaded file \"%v\" successfully.%v", openFileName, extraDialogMessageText)
					d := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_OK, "%s", m)
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
	loadConfCurrentWindowFn := func() {
		loadConfFn(false)
	}
	loadConfNewWindowFn := func() {
		loadConfFn(true)
	}

	delConfItemHandler := func() {
		if len(selectedConfigItemIndexes) > 0 && len(userTX) > 1 {
			// remove the conf item from the UserTX config
			log.Println(selectedConfigItemIndexes)
			// newUserTX := []lib.TX{}
			// for _, i := range selectedConfigItemIndexes {
			for i := len(selectedConfigItemIndexes) - 1; i >= 0; i-- {
				if selectedConfigItemIndexes[i] > len(userTX) {
					return
				}
				log.Println(i, selectedConfigItemIndexes, selectedConfigItemIndexes[i])
				// https://www.golangprograms.com/how-to-delete-an-element-from-a-slice-in-golang.html
				userTX = lib.RemoveTXAtIndex(userTX, selectedConfigItemIndexes[i])
			}
			nb.RemovePage(TAB_CONFIG)
			newConfigSw, newLabel := genConfigView()
			nb.InsertPage(newConfigSw, newLabel, TAB_CONFIG)
			updateResults(false)
			win.ShowAll()
			nb.SetCurrentPage(TAB_CONFIG)
			selectedConfigItemIndexes = []int{}
		}
	}

	// initialize some things
	ui.SetHideInactive(HideInactivePtr)

	hideInactiveCheckBoxClickedHandler := func(chkBtn *gtk.CheckButton) {
		HideInactive = !HideInactive
		chkBtn.SetActive(HideInactive)
		updateResults(false)
		configViewListStore.Clear()
		ui.SyncListStore(&userTX, configViewListStore)
	}

	addConfItemHandler := func() {
		// nb.RemovePage(TAB_CONFIG)
		userTX = append(userTX, lib.TX{
			Order:     len(userTX) + 1,
			Amount:    -500,
			Active:    true,
			Name:      "New",
			Frequency: "WEEKLY",
			Interval:  1,
		})
		updateResults(false)
		configViewListStore.Clear()
		ui.SyncListStore(&userTX, configViewListStore)

		// old code - this is less efficient and jitters the view
		// newConfigSw, newLabel := genConfigView()
		// nb.InsertPage(newConfigSw, newLabel, TAB_CONFIG)
		// win.ShowAll()
		// nb.SetCurrentPage(TAB_CONFIG)
	}

	cloneConfItemHandler := func() {
		if len(selectedConfigItemIndexes) > 0 {
			log.Println(selectedConfigItemIndexes)
			// for _, i := range selectedConfigItemIndexes {
			for i := len(selectedConfigItemIndexes) - 1; i >= 0; i-- {
				if selectedConfigItemIndexes[i] > len(userTX) {
					return
				}
				log.Println(i, selectedConfigItemIndexes, selectedConfigItemIndexes[i])
				// https://www.golangprograms.com/how-to-delete-an-element-from-a-slice-in-golang.html
				// userTX = lib.RemoveTXAtIndex(userTX, selectedConfigItemIndexes[i])
				userTX = append(userTX, userTX[selectedConfigItemIndexes[i]])
				userTX[len(userTX)-1].Order = len(userTX)
			}
			nb.RemovePage(TAB_CONFIG)
			newConfigSw, newLabel := genConfigView()
			nb.InsertPage(newConfigSw, newLabel, TAB_CONFIG)
			updateResults(false)
			win.ShowAll()
			nb.SetCurrentPage(TAB_CONFIG)
		}
		// selectedConfigItemIndexes = []int{}
	}

	hideInactiveCheckbox, err := gtk.CheckButtonNewWithMnemonic("Hide inactive")
	if err != nil {
		log.Fatal("failed to create 'hide inactive' checkbox:", err)
	}
	hideInactiveCheckbox.SetMarginTop(c.UISpacer)
	hideInactiveCheckbox.SetMarginBottom(c.UISpacer)
	hideInactiveCheckbox.SetMarginStart(c.UISpacer)
	hideInactiveCheckbox.SetMarginEnd(c.UISpacer)
	hideInactiveCheckbox.SetActive(HideInactive)

	hideInactiveCheckbox.Connect("clicked", hideInactiveCheckBoxClickedHandler)

	addConfItemBtn, err := gtk.ButtonNewWithLabel("+")
	if err != nil {
		log.Fatal("Unable to create add conf item button:", err)
	}
	addConfItemBtn.SetMarginTop(c.UISpacer)
	addConfItemBtn.SetMarginBottom(c.UISpacer)
	addConfItemBtn.SetMarginStart(c.UISpacer)
	addConfItemBtn.SetMarginEnd(c.UISpacer)

	delConfItemBtn, err := gtk.ButtonNewWithLabel("-")
	if err != nil {
		log.Fatal("Unable to create add conf item button:", err)
	}
	delConfItemBtn.SetMarginTop(c.UISpacer)
	delConfItemBtn.SetMarginBottom(c.UISpacer)
	delConfItemBtn.SetMarginStart(c.UISpacer)
	delConfItemBtn.SetMarginEnd(c.UISpacer)

	cloneConfItemBtn, err := gtk.ButtonNewWithLabel("Clone")
	if err != nil {
		log.Fatal("Unable to create clone conf item button:", err)
	}
	cloneConfItemBtn.SetMarginTop(c.UISpacer)
	cloneConfItemBtn.SetMarginBottom(c.UISpacer)
	cloneConfItemBtn.SetMarginStart(c.UISpacer)
	cloneConfItemBtn.SetMarginEnd(c.UISpacer)

	// saveConfBtn, err := gtk.ButtonNewWithLabel("Save Config")
	// if err != nil {
	// 	log.Fatal("Unable to create save conf item button:", err)
	// }
	// saveConfBtn.SetMarginTop(c.UISpacer)
	// saveConfBtn.SetMarginBottom(c.UISpacer)
	// saveConfBtn.SetMarginStart(c.UISpacer)
	// saveConfBtn.SetMarginEnd(c.UISpacer)

	// loadConfBtn, err := gtk.ButtonNewWithLabel("Load Config")
	// if err != nil {
	// 	log.Fatal("Unable to create load conf item button:", err)
	// }
	// loadConfBtn.SetMarginTop(c.UISpacer)
	// loadConfBtn.SetMarginBottom(c.UISpacer)
	// loadConfBtn.SetMarginStart(c.UISpacer)
	// loadConfBtn.SetMarginEnd(c.UISpacer)

	// saveResultsItemBtn, err := gtk.ButtonNewWithLabel("Save Results")
	// if err != nil {
	// 	log.Fatal("Unable to create save results item button:", err)
	// }
	// saveResultsItemBtn.SetMarginTop(c.UISpacer)
	// saveResultsItemBtn.SetMarginBottom(c.UISpacer)
	// saveResultsItemBtn.SetMarginStart(c.UISpacer)
	// saveResultsItemBtn.SetMarginEnd(c.UISpacer)

	configGrid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Unable to create grid:", err)
	}

	configGrid.SetOrientation(gtk.ORIENTATION_VERTICAL)
	configSw, configTab := genConfigView()
	configGrid.Attach(configSw, 0, 0, 1, 1)
	delConfItemBtn.Connect("clicked", delConfItemHandler)
	addConfItemBtn.Connect("clicked", addConfItemHandler)
	cloneConfItemBtn.Connect("clicked", cloneConfItemHandler)
	// saveConfBtn.Connect("clicked", saveConfHandler)
	// loadConfBtn.Connect("clicked", loadConfHandler)
	nb.AppendPage(configGrid, configTab)

	latestResults, err = lib.GenerateResultsFromDateStrings(
		&userTX,
		startingBalance,
		startDate,
		endDate,
	)
	if err != nil {
		log.Fatal("failed to generate results from date strings", err.Error())
	}
	sw, label, err := ui.GenerateResultsTab(&userTX, latestResults)
	if err != nil {
		log.Fatalf("failed to generate results tab: %v", err.Error())
	}
	nb.AppendPage(sw, label)
	mainBtn.Connect("clicked", updateResultsDefault)

	idleTextSet := func() bool {
		// mainBtn.SetLabel(fmt.Sprintf("Set a label %d time(s)!", nSets))
		log.Printf("go routine %v loops", nSets)

		// Returning false here is unnecessary, as anything but returning true
		// will remove the function from being called by the GTK main loop.
		return false
	}

	// Native GTK is not thread safe, and thus, gotk3's GTK bindings may not
	// be used from other goroutines.  Instead, glib.IdleAdd() must be used
	// to add a function to run in the GTK main loop when it is in an idle
	// state.
	//
	// Two examples of using glib.IdleAdd() are shown below.  The first runs
	// a user created function, LabelSetTextIdle, and passes it two
	// arguments for a label and the text to set it with.  The second calls
	// (*gtk.Label).SetText directly, passing in only the text as an
	// argument.
	//
	// If the function passed to glib.IdleAdd() returns one argument, and
	// that argument is a bool, this return value will be used in the same
	// manner as a native g_idle_add() call.  If this return value is false,
	// the function will be removed from executing in the GTK main loop's
	// idle state.  If the return value is true, the function will continue
	// to execute when the GTK main loop is in this state.
	go func() {
		for {
			time.Sleep(60 * time.Second)
			_ = glib.IdleAdd(idleTextSet)
			nSets++
		}
	}()

	quitApp := func() {
		application.Quit()
	}
	closeWindow := func() {
		win.Close()
		win.Destroy()
	}

	setTabToConfig := func() {
		log.Println(nb.GetCurrentPage())
		nb.SetCurrentPage(TAB_CONFIG)
	}

	setTabToResults := func() {
		log.Println(nb.GetCurrentPage())
		nb.SetCurrentPage(TAB_RESULTS)
	}

	nextTab := func() {
		nb.NextPage()
	}
	prevTab := func() {
		nb.PrevPage()
	}

	newWindow := func() {
		primary(application, "").ShowAll()
	}

	getStats := func() {
		stats, err := lib.GetStats(latestResults)
		if err != nil {
			d := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_WARNING, gtk.BUTTONS_OK, err.Error())
			log.Println(err.Error())
			d.Run()
			d.Destroy()
			return
		}
		m := fmt.Sprintf("%v\n\nWould you like to copy these stats to your clipboard?", stats)
		d := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_YES_NO, m)
		log.Println(stats)
		resp := d.Run()
		d.Destroy()
		if resp == gtk.RESPONSE_YES {
			clipboard, err := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
			if err != nil {
				m = fmt.Sprintf("Failed to access the clipboard: %v", err.Error())
				d := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, m)
				log.Println(m)
				d.Run()
				d.Destroy()
				return
			}
			clipboard.SetText(stats)
			m = ("Success! Copied the stats to the clipboard.")
			d := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_OK, m)
			log.Println(m)
			d.Run()
			d.Destroy()
			return
		}
	}

	loadConfCurrentWindowAction.Connect("activate", loadConfCurrentWindowFn)
	loadConfNewWindowAction.Connect("activate", loadConfNewWindowFn)

	accelerators, _ := gtk.AccelGroupNew()
	keyQ, modCtrl := gtk.AcceleratorParse("<Control>q")
	keyS, modCtrlShift := gtk.AcceleratorParse("<Control><Shift>s")
	keyB, _ := gtk.AcceleratorParse("b")
	keyI, _ := gtk.AcceleratorParse("i")
	keyD, _ := gtk.AcceleratorParse("d")
	keyR, _ := gtk.AcceleratorParse("r")
	keyO, _ := gtk.AcceleratorParse("o")
	keyP, _ := gtk.AcceleratorParse("p")
	keyW, _ := gtk.AcceleratorParse("w")
	keyC, _ := gtk.AcceleratorParse("c")
	keyN, _ := gtk.AcceleratorParse("n")
	key1, modAlt := gtk.AcceleratorParse("<alt>1")
	key2, _ := gtk.AcceleratorParse("2")
	accelerators.Connect(keyQ, modCtrl, gtk.ACCEL_VISIBLE, quitApp)
	accelerators.Connect(keyW, modCtrlShift, gtk.ACCEL_VISIBLE, quitApp)
	accelerators.Connect(keyW, modCtrl, gtk.ACCEL_VISIBLE, closeWindow)
	accelerators.Connect(keyN, modCtrlShift, gtk.ACCEL_VISIBLE, newWindow)
	accelerators.Connect(keyN, modCtrl, gtk.ACCEL_VISIBLE, newWindow)
	accelerators.Connect(keyS, modCtrlShift, gtk.ACCEL_VISIBLE, saveConfAsFn)
	accelerators.Connect(keyS, modCtrl, gtk.ACCEL_VISIBLE, saveOpenConfFn)
	accelerators.Connect(keyP, modCtrl, gtk.ACCEL_VISIBLE, saveResultsFn)
	accelerators.Connect(keyP, modCtrlShift, gtk.ACCEL_VISIBLE, copyResultsFn)
	accelerators.Connect(keyC, modCtrlShift, gtk.ACCEL_VISIBLE, copyResultsFn)
	accelerators.Connect(gdk.KEY_ISO_Enter, modCtrl, gtk.ACCEL_VISIBLE, updateResultsSilent)
	accelerators.Connect(keyR, modCtrl, gtk.ACCEL_VISIBLE, updateResultsDefault)
	accelerators.Connect(keyO, modCtrl, gtk.ACCEL_VISIBLE, loadConfCurrentWindowFn)
	accelerators.Connect(keyO, modCtrlShift, gtk.ACCEL_VISIBLE, loadConfNewWindowFn)
	accelerators.Connect(keyD, modCtrl, gtk.ACCEL_VISIBLE, delConfItemHandler)
	accelerators.Connect(keyB, modCtrl, gtk.ACCEL_VISIBLE, addConfItemHandler)
	accelerators.Connect(keyD, modCtrlShift, gtk.ACCEL_VISIBLE, cloneConfItemHandler)
	accelerators.Connect(keyI, modCtrl, gtk.ACCEL_VISIBLE, getStats)
	accelerators.Connect(key1, modAlt, gtk.ACCEL_VISIBLE, setTabToConfig)
	accelerators.Connect(key2, modAlt, gtk.ACCEL_VISIBLE, setTabToResults)
	accelerators.Connect(gdk.KEY_KP_Page_Down, modCtrlShift, gtk.ACCEL_VISIBLE, prevTab)
	accelerators.Connect(gdk.KEY_KP_Page_Up, modCtrlShift, gtk.ACCEL_VISIBLE, nextTab)
	accelerators.Connect(gdk.KEY_Tab, modCtrl, gtk.ACCEL_VISIBLE, nextTab)
	accelerators.Connect(gdk.KEY_Tab, modCtrlShift, gtk.ACCEL_VISIBLE, prevTab)

	const FULL_GRID_WIDTH = 3
	grid.Attach(nb, 0, 0, FULL_GRID_WIDTH, 2)
	grid.Attach(hideInactiveCheckbox, 0, 2, 1, 1)
	grid.Attach(addConfItemBtn, 0, 3, 1, 1)
	grid.Attach(delConfItemBtn, 1, 3, 1, 1)
	grid.Attach(cloneConfItemBtn, 2, 3, 1, 1)
	grid.Attach(startingBalanceInput, 0, 4, 1, 1)
	grid.Attach(stDateInput, 1, 4, 1, 1)
	grid.Attach(endDateInput, 2, 4, 1, 1)
	// grid.Attach(mainBtn, 3, 3, 1, 1)
	// grid.Attach(saveResultsItemBtn, 4, 3, 1, 1)

	header.PackStart(mbtn)
	rootBox.PackStart(grid, true, true, 0)
	win.Add(rootBox)
	win.AddAccelGroup(accelerators)
	win.SetTitlebar(header)
	win.SetPosition(gtk.WIN_POS_CENTER)
	win.SetDefaultSize(1366, 768)
	win.ShowAll()

	// gtk.Main()

	// r.GET("/", func(c *gin.Context) {
	// 	now := time.Now()
	// 	results, err := lib.GetResults(
	// 		userTX,
	// 		time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
	// 		time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC),
	// 		50000,
	// 	)

	// 	if err != nil {
	// 		log.Printf("failed to get results: %v", err.Error())
	// 	}

	// 	// log.Printf("results: %v", results)

	// 	c.HTML(http.StatusOK, "index.html", map[string]interface{}{
	// 		"now":     time.Date(2017, 07, 01, 0, 0, 0, 0, time.UTC),
	// 		"Results": results,
	// 	})
	// })
	// r.GET("/static/bootstrap.min.css", func(c *gin.Context) {
	// 	c.File("./static/bootstrap.min.css")
	// })
	// r.GET("/static/styles.css", func(c *gin.Context) {
	// 	c.File("./static/styles.css")
	// })

	// err = r.Run("0.0.0.0:10982")
	// if err != nil {
	// 	log.Printf("error: %v", err.Error())
	// }
	return win
}
