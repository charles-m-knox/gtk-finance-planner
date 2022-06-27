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

	application.Connect(c.GtkSignalActivate, func() {
		win := primary(application, defaultConfigFile)

		aNew := glib.SimpleActionNew(c.ActionNew, nil)
		aNew.Connect(c.GtkSignalActivate, func() {
			primary(application, defaultConfigFile).ShowAll()
		})
		application.AddAction(aNew)

		aQuit := glib.SimpleActionNew(c.ActionQuit, nil)
		aQuit.Connect(c.GtkSignalActivate, func() {
			application.Quit()
		})
		application.AddAction(aQuit)

		win.ShowAll()
	})

	os.Exit(application.Run(os.Args))
}

// primary just creates an instance of a new finance planner window; there
// can be multiple windows per application session
func primary(application *gtk.Application, filename string) *gtk.ApplicationWindow {
	// for go routine access
	var (
		nSets                     = 1
		startingBalance           = 50000
		startDate                 = lib.GetNowDateString()
		endDate                   = lib.GetDefaultEndDateString()
		selectedConfigItemIndexes []int
		openFileName              = filename
		userTX                    []lib.TX
		err                       error
		latestResults             []lib.Result
		resultsListStore          *gtk.ListStore
		configViewListStore       *gtk.ListStore
	)

	// initialize some values
	resultsListStore, err = ui.GetNewResultsListStore()
	if err != nil {
		log.Fatalf("failed to initialize results list store: %v", err.Error())
	}

	win, rootBox, header, mbtn, menu := ui.GetMainWindowRootElements(application)

	win.SetTitle(c.FinancialPlanner)
	header.SetTitle(c.FinancialPlanner)
	header.SetShowCloseButton(true)
	header.SetSubtitle(openFileName)
	mbtn.SetMenuModel(&menu.MenuModel)

	ui.ProcessInitialConfigLoad(
		win,
		header,
		openFileName,
		&userTX,
	)

	aClose := glib.SimpleActionNew(c.ActionClose, nil)
	aClose.Connect(c.GtkSignalActivate, func() {
		win.Close()
	})
	win.AddAction(aClose)

	// create and insert custom action group with prefix "fin" (for finances)
	finActionGroup := glib.SimpleActionGroupNew()
	win.InsertActionGroup("fin", finActionGroup)

	saveConfAsFn := func() {
		ui.SaveConfAs(win, header, &openFileName, &userTX)
	}

	saveOpenConfFn := func() {
		ui.SaveOpenConf(win, header, &openFileName, &userTX)
	}

	saveResultsFn := func() {
		ui.SaveResults(win, header, &latestResults)
	}

	copyResultsFn := func() {
		ui.CopyResults(win, header, &latestResults)
	}

	saveConfAsAction := glib.SimpleActionNew(c.ActionSaveConfig, nil)
	saveConfAsAction.Connect(c.GtkSignalActivate, saveConfAsFn)
	finActionGroup.AddAction(saveConfAsAction)
	win.AddAction(saveConfAsAction)

	saveOpenConfAction := glib.SimpleActionNew(c.ActionSaveOpenConfig, nil)
	saveOpenConfAction.Connect(c.GtkSignalActivate, saveOpenConfFn)
	finActionGroup.AddAction(saveOpenConfAction)
	win.AddAction(saveOpenConfAction)

	saveResultsAction := glib.SimpleActionNew(c.ActionSaveResults, nil)
	saveResultsAction.Connect(c.GtkSignalActivate, saveResultsFn)
	finActionGroup.AddAction(saveResultsAction)
	win.AddAction(saveResultsAction)

	copyResultsAction := glib.SimpleActionNew(c.ActionCopyResults, nil)
	copyResultsAction.Connect(c.GtkSignalActivate, copyResultsFn)
	finActionGroup.AddAction(copyResultsAction)
	win.AddAction(copyResultsAction)

	loadConfCurrentWindowAction := glib.SimpleActionNew(c.ActionLoadConfigCurrentWindow, nil)
	finActionGroup.AddAction(loadConfCurrentWindowAction)
	win.AddAction(loadConfCurrentWindowAction)

	loadConfNewWindowAction := glib.SimpleActionNew(c.ActionLoadConfigNewWindow, nil)
	finActionGroup.AddAction(loadConfNewWindowAction)
	win.AddAction(loadConfNewWindowAction)

	nb, err := gtk.NotebookNew()
	if err != nil {
		log.Fatal("failed to create notebook:", err)
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

		header.SetSubtitle(fmt.Sprintf("%v*", openFileName))
		// TODO: clear and re-populate the list store instead of removing the
		// entire tab page
		// nb.RemovePage(c.TAB_RESULTS)
		// newSw, newLabel, resultsListStore, err := ui.GenerateResultsTab(&userTX, latestResults)
		// if err != nil {
		// 	log.Fatalf("failed to generate results tab: %v", err.Error())
		// }
		// nb.InsertPage(newSw, newLabel, c.TAB_RESULTS)
		if resultsListStore != nil {
			err = ui.SyncResultsListStore(&latestResults, resultsListStore)
			if err != nil {
				log.Print("failed to sync results list store:", err.Error())
			}
			win.ShowAll()
			if switchTo {
				nb.SetCurrentPage(c.TAB_RESULTS)
			}
		}
	}

	updateResultsDefault := func() {
		updateResults(true)
	}

	updateResultsSilent := func() {
		updateResults(false)
	}

	startingBalanceInput, err := gtk.EntryNew()
	if err != nil {
		log.Fatal("failed to create starting balance input entry:", err)
	}
	ui.SetSpacerMarginsGtkEntry(startingBalanceInput)
	startingBalanceInput.SetPlaceholderText(c.BalanceInputPlaceholderText)
	updateStartingBalance := func() {
		s, _ := startingBalanceInput.GetText()
		if s == "" {
			return
		}
		startingBalance = int(lib.ParseDollarAmount(s, true))
		startingBalanceInput.SetText(lib.FormatAsCurrency(startingBalance))
		updateResults(true)
	}
	startingBalanceInput.Connect(c.GtkSignalActivate, updateStartingBalance)

	// text input for providing the starting date
	stDateInput, err := gtk.EntryNew()
	if err != nil {
		log.Fatal("failed to create start date input entry:", err)
	}
	ui.SetSpacerMarginsGtkEntry(stDateInput)
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
	stDateInput.Connect(c.GtkSignalActivate, stDateInputUpdate)
	stDateInput.Connect("focus-out-event", stDateInputUpdate)

	// text input for providing the ending date
	endDateInput, err := gtk.EntryNew()
	if err != nil {
		log.Fatal("failed to create end date input entry:", err)
	}
	ui.SetSpacerMarginsGtkEntry(endDateInput)
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
	endDateInput.Connect(c.GtkSignalActivate, endDateInputUpdate)
	endDateInput.Connect("focus-out-event", endDateInputUpdate)

	grid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("failed to create grid:", err)
	}
	grid.SetOrientation(gtk.ORIENTATION_VERTICAL)

	nb.SetHExpand(true)
	nb.SetVExpand(true)

	// now build the config tab page
	genConfigView := func() (*gtk.ScrolledWindow, *gtk.Label) {
		// TODO: refactor the config tree view list store using the same
		// approach that was used for the results list store
		configTreeView, configListStore, err := ui.GetConfigAsTreeView(&userTX, updateResultsSilent)
		if err != nil {
			log.Fatalf("failed to get config as tree view: %v", err.Error())
		}
		// so we can refer to them later
		configViewListStore = configListStore

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
		p.Connect(c.ActionClose, func() { p.Close() })
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
					nb.RemovePage(c.TAB_CONFIG)
					newConfigSw, newLabel := genConfigView()
					nb.InsertPage(newConfigSw, newLabel, c.TAB_CONFIG)
					updateResults(false)
					win.ShowAll()
					nb.SetCurrentPage(c.TAB_CONFIG)
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
				// TODO: remove this logging, but only once you've gotten rid
				// of weird edge cases - for example, clearing the tree view's
				// list store with any selected items causes a large number of
				// selection changes (this is probalby very inefficient). To
				// address this, you need to figure out how to un-select values
				log.Println(i, selectedConfigItemIndexes, selectedConfigItemIndexes[i])
				userTX = lib.RemoveTXAtIndex(userTX, selectedConfigItemIndexes[i])
			}

			updateResults(false)
			ui.SyncListStore(&userTX, configViewListStore)

			// TODO: old code - jittery and inefficient
			// nb.RemovePage(c.TAB_CONFIG)
			// newConfigSw, newLabel := genConfigView()
			// nb.InsertPage(newConfigSw, newLabel, c.TAB_CONFIG)
			// updateResults(false)
			// win.ShowAll()
			// nb.SetCurrentPage(c.TAB_CONFIG)

			selectedConfigItemIndexes = []int{}
		}
	}

	// initialize some things
	ui.SetHideInactive(HideInactivePtr)

	hideInactiveCheckBoxClickedHandler := func(chkBtn *gtk.CheckButton) {
		HideInactive = !HideInactive
		chkBtn.SetActive(HideInactive)
		updateResults(false)
		ui.SyncListStore(&userTX, configViewListStore)
	}

	addConfItemHandler := func() {
		// nb.RemovePage(c.TAB_CONFIG)

		// TODO: create/use helper function that generates new TX instances
		// TODO: refactor
		userTX = append(userTX, lib.TX{
			Order:     len(userTX) + 1,
			Amount:    -500,
			Active:    true,
			Name:      "New",
			Frequency: "WEEKLY",
			Interval:  1,
		})

		updateResults(false)
		ui.SyncListStore(&userTX, configViewListStore)

		// old code - this is less efficient and jitters the view
		// newConfigSw, newLabel := genConfigView()
		// nb.InsertPage(newConfigSw, newLabel, c.TAB_CONFIG)
		// win.ShowAll()
		// nb.SetCurrentPage(c.TAB_CONFIG)
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
				userTX = append(userTX, userTX[selectedConfigItemIndexes[i]])
				userTX[len(userTX)-1].Order = len(userTX)
			}

			updateResults(false)
			ui.SyncListStore(&userTX, configViewListStore)

			// TODO: old code - less efficient and jittery
			// nb.RemovePage(c.TAB_CONFIG)
			// newConfigSw, newLabel := genConfigView()
			// nb.InsertPage(newConfigSw, newLabel, c.TAB_CONFIG)
			// updateResults(false)
			// win.ShowAll()
			// nb.SetCurrentPage(c.TAB_CONFIG)
		}
		// selectedConfigItemIndexes = []int{}
	}

	hideInactiveCheckbox, err := gtk.CheckButtonNewWithMnemonic(c.HideInactiveBtnLabel)
	if err != nil {
		log.Fatal("failed to create 'hide inactive' checkbox:", err)
	}
	ui.SetSpacerMarginsGtkCheckBtn(hideInactiveCheckbox)
	hideInactiveCheckbox.SetActive(HideInactive)

	hideInactiveCheckbox.Connect(c.GtkSignalClicked, hideInactiveCheckBoxClickedHandler)

	addConfItemBtn, err := gtk.ButtonNewWithLabel(c.AddBtnLabel)
	if err != nil {
		log.Fatal("failed to create add conf item button:", err)
	}
	ui.SetSpacerMarginsGtkBtn(addConfItemBtn)

	delConfItemBtn, err := gtk.ButtonNewWithLabel(c.DelBtnLabel)
	if err != nil {
		log.Fatal("failed to create add conf item button:", err)
	}
	ui.SetSpacerMarginsGtkBtn(delConfItemBtn)

	cloneConfItemBtn, err := gtk.ButtonNewWithLabel(c.CloneBtnLabel)
	if err != nil {
		log.Fatal("failed to create clone conf item button:", err)
	}
	ui.SetSpacerMarginsGtkBtn(cloneConfItemBtn)

	configGrid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("failed to create grid:", err)
	}

	configGrid.SetOrientation(gtk.ORIENTATION_VERTICAL)
	configSw, configTab := genConfigView()
	configGrid.Attach(configSw, 0, 0, 1, 1)
	delConfItemBtn.Connect(c.GtkSignalClicked, delConfItemHandler)
	addConfItemBtn.Connect(c.GtkSignalClicked, addConfItemHandler)
	cloneConfItemBtn.Connect(c.GtkSignalClicked, cloneConfItemHandler)
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
	sw, label, err := ui.GenerateResultsTab(&userTX, latestResults, resultsListStore)
	if err != nil {
		log.Fatalf("failed to generate results tab: %v", err.Error())
	}
	nb.AppendPage(sw, label)

	// Native GTK is not thread safe, and thus, gotk3's GTK bindings may not
	// be used from other goroutines.  Instead, glib.IdleAdd() must be used
	// to add a function to run in the GTK main loop when it is in an idle
	// state.
	//
	// If the function passed to glib.IdleAdd() returns one argument, and
	// that argument is a bool, this return value will be used in the same
	// manner as a native g_idle_add() call.  If this return value is false,
	// the function will be removed from executing in the GTK main loop's
	// idle state.  If the return value is true, the function will continue
	// to execute when the GTK main loop is in this state.
	idleTextSet := func() bool {
		// mainBtn.SetLabel(fmt.Sprintf("Set a label %d time(s)!", nSets))
		log.Printf("go routine %v loops", nSets)

		// Returning false here is unnecessary, as anything but returning true
		// will remove the function from being called by the GTK main loop.
		return false
	}
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
		nb.SetCurrentPage(c.TAB_CONFIG)
	}

	setTabToResults := func() {
		nb.SetCurrentPage(c.TAB_RESULTS)
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
		ui.GetStats(win, &latestResults)
	}

	loadConfCurrentWindowAction.Connect(c.GtkSignalActivate, loadConfCurrentWindowFn)
	loadConfNewWindowAction.Connect(c.GtkSignalActivate, loadConfNewWindowFn)

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

	grid.Attach(nb, 0, 0, c.FullGridWidth, 2)
	grid.Attach(hideInactiveCheckbox, 0, 2, 1, 1)
	grid.Attach(addConfItemBtn, 0, 3, 1, 1)
	grid.Attach(delConfItemBtn, 1, 3, 1, 1)
	grid.Attach(cloneConfItemBtn, 2, 3, 1, 1)
	grid.Attach(startingBalanceInput, 0, 4, 1, 1)
	grid.Attach(stDateInput, 1, 4, 1, 1)
	grid.Attach(endDateInput, 2, 4, 1, 1)

	header.PackStart(mbtn)
	rootBox.PackStart(grid, true, true, 0)
	win.Add(rootBox)
	win.AddAccelGroup(accelerators)
	win.SetTitlebar(header)
	win.SetPosition(gtk.WIN_POS_CENTER)
	win.SetDefaultSize(1366, 768)
	win.ShowAll()

	return win
}
