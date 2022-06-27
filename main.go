package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"time"

	c "finance-planner/constants"
	"finance-planner/lib"
	"finance-planner/state"
	"finance-planner/ui"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

//go:embed assets/*.png
var embeddedIconFS embed.FS

func main() {
	application, err := gtk.ApplicationNew(c.GtkAppID, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		log.Fatal("failed to create gtk application:", err)
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
		ws := primary(application, defaultConfigFile)

		aNew := glib.SimpleActionNew(c.ActionNew, nil)
		aNew.Connect(c.GtkSignalActivate, func() {
			ws := primary(application, defaultConfigFile)
			ws.Win.ShowAll()
		})
		application.AddAction(aNew)

		aQuit := glib.SimpleActionNew(c.ActionQuit, nil)
		aQuit.Connect(c.GtkSignalActivate, func() {
			application.Quit()
		})
		application.AddAction(aQuit)

		ws.Win.ShowAll()
	})

	os.Exit(application.Run(os.Args))
}

// primary just creates an instance of a new finance planner window; there
// can be multiple windows per application session
func primary(application *gtk.Application, filename string) *state.WinState {
	// variables stored like this can be accessed via go routines if needed
	var (
		nSets = 1
		err   error
		ws    *state.WinState
	)

	ws = &state.WinState{
		HideInactive:        false,
		ConfigColumnSort:    c.None,
		OpenFileName:        filename,
		StartingBalance:     50000,
		StartDate:           lib.GetNowDateString(),
		EndDate:             lib.GetDefaultEndDateString(),
		SelectedConfigItems: []int{},
		TX:                  &[]lib.TX{},
		Results:             &[]lib.Result{},
		App:                 application,
	}

	// initialize some values
	ws.ResultsListStore, err = ui.GetNewResultsListStore()
	if err != nil {
		log.Fatalf("failed to initialize results list store: %v", err.Error())
	}

	ws.ConfigListStore, err = ui.GetNewConfigListStore()
	if err != nil {
		log.Fatalf("failed to initialize config list store: %v", err.Error())
	}

	win, rootBox, header, mbtn, menu := ui.GetMainWindowRootElements(application)
	ws.Win = win
	ws.Header = header

	// TODO: investigate svg pixbuf loading instead of png loading
	// pb, err := gdk.PixbufNewFromFile("./assets/icon-128.png")
	icon, err := embeddedIconFS.ReadFile(c.IconAssetPath)
	if err != nil {
		log.Fatalf("failed to load embedded app icon: %v", err.Error())
	}

	pb, err := gdk.PixbufNewFromBytesOnly(icon)
	if err != nil {
		log.Fatalf("failed to read app icon: %v", err.Error())
	}
	win.SetIcon(pb)

	win.SetTitle(c.FinancialPlanner)
	header.SetTitle(c.FinancialPlanner)
	header.SetShowCloseButton(true)
	mbtn.SetMenuModel(&menu.MenuModel)

	ui.ProcessInitialConfigLoad(win, &ws.OpenFileName, ws.TX)
	header.SetSubtitle(ws.OpenFileName)

	aClose := glib.SimpleActionNew(c.ActionClose, nil)
	aClose.Connect(c.GtkSignalActivate, func() {
		win.Close()
	})
	win.AddAction(aClose)

	saveConfAsFn := func() {
		ui.SaveConfAs(ws.Win, ws.Header, &ws.OpenFileName, ws.TX)
	}

	saveOpenConfFn := func() {
		ui.SaveOpenConf(ws.Win, ws.Header, &ws.OpenFileName, ws.TX)
	}

	saveResultsFn := func() {
		ui.SaveResults(ws.Win, ws.Header, ws.Results)
	}

	copyResultsFn := func() {
		ui.CopyResults(ws.Win, ws.Header, ws.Results)
	}

	saveConfAsAction := glib.SimpleActionNew(c.ActionSaveConfig, nil)
	saveOpenConfAction := glib.SimpleActionNew(c.ActionSaveOpenConfig, nil)
	saveResultsAction := glib.SimpleActionNew(c.ActionSaveResults, nil)
	copyResultsAction := glib.SimpleActionNew(c.ActionCopyResults, nil)
	loadConfCurrentWindowAction := glib.SimpleActionNew(c.ActionLoadConfigCurrentWindow, nil)
	loadConfNewWindowAction := glib.SimpleActionNew(c.ActionLoadConfigNewWindow, nil)

	saveConfAsAction.Connect(c.GtkSignalActivate, saveConfAsFn)
	saveOpenConfAction.Connect(c.GtkSignalActivate, saveOpenConfFn)
	saveResultsAction.Connect(c.GtkSignalActivate, saveResultsFn)
	copyResultsAction.Connect(c.GtkSignalActivate, copyResultsFn)

	// create and insert custom action group with prefix "fin" (for finances)
	finActionGroup := glib.SimpleActionGroupNew()
	finActionGroup.AddAction(saveConfAsAction)
	finActionGroup.AddAction(saveOpenConfAction)
	finActionGroup.AddAction(saveResultsAction)
	finActionGroup.AddAction(copyResultsAction)
	finActionGroup.AddAction(loadConfCurrentWindowAction)
	finActionGroup.AddAction(loadConfNewWindowAction)

	ws.Win.InsertActionGroup("fin", finActionGroup)
	ws.Win.AddAction(saveConfAsAction)
	ws.Win.AddAction(saveOpenConfAction)
	ws.Win.AddAction(saveResultsAction)
	ws.Win.AddAction(copyResultsAction)
	ws.Win.AddAction(loadConfCurrentWindowAction)
	ws.Win.AddAction(loadConfNewWindowAction)

	nb, err := gtk.NotebookNew()
	ws.Notebook = nb
	if err != nil {
		log.Fatal("failed to create notebook:", err)
	}

	startingBalanceInput, stDateInput, endDateInput := ui.GetResultsInputs(ws)

	grid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("failed to create grid:", err)
	}
	grid.SetOrientation(gtk.ORIENTATION_VERTICAL)

	nb.SetHExpand(true)
	nb.SetVExpand(true)

	hideInactiveCheckBoxClickedHandler := func(chkBtn *gtk.CheckButton) {
		ws.HideInactive = !ws.HideInactive
		chkBtn.SetActive(ws.HideInactive)
		ui.SyncConfigListStore(ws)
	}

	hideInactiveCheckbox, err := gtk.CheckButtonNewWithMnemonic(c.HideInactiveBtnLabel)
	if err != nil {
		log.Fatal("failed to create 'hide inactive' checkbox:", err)
	}
	ui.SetSpacerMarginsGtkCheckBtn(hideInactiveCheckbox)
	hideInactiveCheckbox.SetActive(ws.HideInactive)

	hideInactiveCheckbox.Connect(c.GtkSignalClicked, hideInactiveCheckBoxClickedHandler)

	addConfItemBtn, delConfItemBtn, cloneConfItemBtn := ui.GetConfEditButtons(ws)

	configGrid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("failed to create grid:", err)
	}

	configGrid.SetOrientation(gtk.ORIENTATION_VERTICAL)
	configSw, configTab := ui.GetConfigTab(ws)
	configGrid.Attach(configSw, 0, 0, 1, 1)

	nb.AppendPage(configGrid, configTab)

	*ws.Results, err = lib.GenerateResultsFromDateStrings(
		ws.TX,
		ws.StartingBalance,
		ws.StartDate,
		ws.EndDate,
	)
	if err != nil {
		log.Fatal("failed to generate results from date strings", err.Error())
	}
	sw, label, err := ui.GenerateResultsTab(ws.TX, *ws.Results, ws.ResultsListStore)
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

	delConfItemHandler := func() {
		ui.DelConfItem(ws)
	}

	addConfItemHandler := func() {
		ui.AddConfItem(ws)
	}

	cloneConfItemHandler := func() {
		ui.CloneConfItem(ws)
	}

	loadConfCurrentWindowFn := func() {
		ui.LoadConfig(ws, primary, false)
	}
	loadConfNewWindowFn := func() {
		ui.LoadConfig(ws, primary, true)
	}

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
		primary(application, "").Win.ShowAll()
	}

	getStats := func() {
		ui.GetStats(win, ws.Results)
	}

	updateResultsDefault := func() {
		ui.UpdateResults(ws, true)
	}

	updateResultsSilent := func() {
		ui.UpdateResults(ws, false)
	}

	loadConfCurrentWindowAction.Connect(c.GtkSignalActivate, loadConfCurrentWindowFn)
	loadConfNewWindowAction.Connect(c.GtkSignalActivate, loadConfNewWindowFn)

	addConfItemBtn.Connect(c.GtkSignalClicked, addConfItemHandler)
	delConfItemBtn.Connect(c.GtkSignalClicked, delConfItemHandler)
	cloneConfItemBtn.Connect(c.GtkSignalClicked, cloneConfItemHandler)

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

	return ws
}
