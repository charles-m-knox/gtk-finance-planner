package main

import (
	"embed"
	"fmt"
	"log"
	"os"

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
		err error
		ws  *state.WinState
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
		ws.Win.Close()
		ws.Win.Destroy()
	}

	setTabToConfig := func() {
		ws.Notebook.SetCurrentPage(c.TAB_CONFIG)
	}

	setTabToResults := func() {
		ws.Notebook.SetCurrentPage(c.TAB_RESULTS)
	}

	nextTab := func() {
		ws.Notebook.NextPage()
	}
	prevTab := func() {
		ws.Notebook.PrevPage()
	}

	newWindow := func() {
		primary(application, "").Win.ShowAll()
	}

	getStats := func() {
		ui.GetStats(ws.Win, ws.Results)
	}

	updateResultsDefault := func() {
		ui.UpdateResults(ws, true)
	}

	updateResultsSilent := func() {
		ui.UpdateResults(ws, false)
	}

	// instantiation of graphical components begins next

	win, rootBox, header, mbtn, menu := ui.GetMainWindowRootElements(application)
	ws.Win = win
	ws.Header = header

	ui.ProcessInitialConfigLoad(ws.Win, &ws.OpenFileName, ws.TX)

	nb, grid := ui.GetStructuralComponents(ws)
	ws.Notebook = nb

	cfgGrid, resultsGrid := ui.SetupNotebookPages(ws)
	ui.SetWinIcon(ws, embeddedIconFS)

	ws.Win.SetTitle(c.FinancialPlanner)
	ws.Header.SetTitle(c.FinancialPlanner)
	ws.Header.SetSubtitle(ws.OpenFileName)
	ws.Header.SetShowCloseButton(true)
	mbtn.SetMenuModel(&menu.MenuModel)

	startingBalanceInput, stDateInput, endDateInput := ui.GetResultsInputs(ws)
	addConfItemBtn, delConfItemBtn, cloneConfItemBtn := ui.GetConfEditButtons(ws)
	hideInactiveCheckbox := ui.GetHideInactiveCheckbox(ws)

	// all graphical components have been instantiated now - the next part
	// is to connect signals, functions, and accelerators

	// actions (triggered via menus or accelerators)
	closeWinAction := glib.SimpleActionNew(c.ActionClose, nil)
	saveConfAsAction := glib.SimpleActionNew(c.ActionSaveConfig, nil)
	saveOpenConfAction := glib.SimpleActionNew(c.ActionSaveOpenConfig, nil)
	saveResultsAction := glib.SimpleActionNew(c.ActionSaveResults, nil)
	copyResultsAction := glib.SimpleActionNew(c.ActionCopyResults, nil)
	loadConfCurrentWindowAction := glib.SimpleActionNew(c.ActionLoadConfigCurrentWindow, nil)
	loadConfNewWindowAction := glib.SimpleActionNew(c.ActionLoadConfigNewWindow, nil)

	// create and insert custom action group with prefix "fin" (for finances)
	finActionGroup := glib.SimpleActionGroupNew()
	finActionGroup.AddAction(saveConfAsAction)
	finActionGroup.AddAction(saveOpenConfAction)
	finActionGroup.AddAction(saveResultsAction)
	finActionGroup.AddAction(copyResultsAction)
	finActionGroup.AddAction(loadConfCurrentWindowAction)
	finActionGroup.AddAction(loadConfNewWindowAction)

	ws.Win.InsertActionGroup("fin", finActionGroup)
	ws.Win.AddAction(closeWinAction)
	ws.Win.AddAction(saveConfAsAction)
	ws.Win.AddAction(saveOpenConfAction)
	ws.Win.AddAction(saveResultsAction)
	ws.Win.AddAction(copyResultsAction)
	ws.Win.AddAction(loadConfCurrentWindowAction)
	ws.Win.AddAction(loadConfNewWindowAction)

	closeWinAction.Connect(c.GtkSignalActivate, closeWindow)
	saveConfAsAction.Connect(c.GtkSignalActivate, saveConfAsFn)
	saveOpenConfAction.Connect(c.GtkSignalActivate, saveOpenConfFn)
	saveResultsAction.Connect(c.GtkSignalActivate, saveResultsFn)
	copyResultsAction.Connect(c.GtkSignalActivate, copyResultsFn)
	loadConfCurrentWindowAction.Connect(c.GtkSignalActivate, loadConfCurrentWindowFn)
	loadConfNewWindowAction.Connect(c.GtkSignalActivate, loadConfNewWindowFn)

	// buttons
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

	// everything has been connected up and instantiated - now we can proceed
	// to attach components to a grid and render them

	cfgGrid.Attach(hideInactiveCheckbox, 0, c.ScrolledWindowGridHeight, c.HalfGridWidth, c.ControlsGridHeight)
	cfgGrid.Attach(addConfItemBtn, 0, c.ScrolledWindowGridHeight+1, c.HalfGridWidth, c.ControlsGridHeight)
	cfgGrid.Attach(delConfItemBtn, 1, c.ScrolledWindowGridHeight+1, c.HalfGridWidth, c.ControlsGridHeight)
	cfgGrid.Attach(cloneConfItemBtn, 0, c.ScrolledWindowGridHeight+2, c.FullGridWidth, c.ControlsGridHeight)

	resultsGrid.Attach(startingBalanceInput, 0, c.ScrolledWindowGridHeight, c.FullGridWidth, c.ControlsGridHeight)
	resultsGrid.Attach(stDateInput, 0, c.ScrolledWindowGridHeight+1, c.HalfGridWidth, c.ControlsGridHeight)
	resultsGrid.Attach(endDateInput, 1, c.ScrolledWindowGridHeight+1, c.HalfGridWidth, c.ControlsGridHeight)

	grid.Attach(ws.Notebook, 0, 0, c.FullGridWidth, c.ScrolledWindowGridHeight)

	ws.Header.PackStart(mbtn)
	rootBox.PackStart(grid, true, true, 0)
	ws.Win.Add(rootBox)
	ws.Win.AddAccelGroup(accelerators)
	ws.Win.SetTitlebar(ws.Header)
	ws.Win.SetPosition(gtk.WIN_POS_CENTER)
	ws.Win.SetDefaultSize(1366, 768)
	ws.Win.ShowAll()

	return ws
}
