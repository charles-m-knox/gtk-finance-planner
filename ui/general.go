package ui

import (
	"embed"
	"fmt"
	"log"

	c "finance-planner/constants"
	"finance-planner/lib"
	"finance-planner/state"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func createCheckboxColumn(title string, columnID int, radio bool, listStore *gtk.ListStore, ws *state.WinState) (tvc *gtk.TreeViewColumn, err error) {
	cellRenderer, err := gtk.CellRendererToggleNew()
	if err != nil {
		return tvc, fmt.Errorf("unable to create checkbox column renderer: %v", err.Error())
	}
	cellRenderer.SetActive(true)
	cellRenderer.SetRadio(radio)
	cellRenderer.SetActivatable(true)
	cellRenderer.SetVisible(true)
	cellRenderer.Connect("toggled", func(a *gtk.CellRendererToggle, path string) {
		ConfigChange(ws, path, columnID, a.GetActive())
	})

	column, err := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "active", columnID)
	if err != nil {
		return tvc, fmt.Errorf("unable to create checkbox cell column: %v", err.Error())
	}
	column.SetResizable(true)
	column.SetClickable(true)
	column.SetVisible(true)

	return column, nil
}

func ProcessInitialConfigLoad(win *gtk.ApplicationWindow, openFileName *string, userTX *[]lib.TX) {
	if *openFileName != "" {
		var err error
		*userTX, err = lib.LoadConfig(*openFileName)
		if err != nil {
			ConfigLoadErrorPromptFlow(win, *openFileName, userTX)
		}
	}

	if len(*userTX) == 0 {
		newTX := lib.GetNewTX()
		newTX.Order = 1
		*userTX = []lib.TX{newTX}
		EmptyConfigLoadSuccessDialog(win, *openFileName)
	}
}

// ConfigLoadErrorPromptFlow occurs when the application tries to load the user
// transactions from a config file, but the file is either invalid, empty,
// or any other error present when loading.
func ConfigLoadErrorPromptFlow(win *gtk.ApplicationWindow, openFileName string, userTX *[]lib.TX) {
	m := fmt.Sprintf(
		"Config does not exist (or is not accessible) at %v. Would you like to create a new one there now?",
		openFileName,
	)

	d := gtk.MessageDialogNew(win,
		gtk.DIALOG_MODAL,
		gtk.MESSAGE_QUESTION,
		gtk.BUTTONS_YES_NO,
		"%s",
		m,
	)

	log.Println(m)

	resp := d.Run()
	if resp == gtk.RESPONSE_YES {
		err := lib.SaveConfig(openFileName, *userTX)
		if err != nil {
			m := fmt.Sprintf(
				"Failed to save config upon window load - will proceed with a blank config. Here's the error: %v",
				err.Error(),
			)
			di := gtk.MessageDialogNew(
				win,
				gtk.DIALOG_MODAL,
				gtk.MESSAGE_ERROR,
				gtk.BUTTONS_OK,
				"%s",
				m,
			)
			log.Println(m)
			di.Run()
			di.Destroy()
		}
	}
	d.Destroy()
}

// EmptyConfigLoadSuccessDialog shows a dialog indicating that the user doesn't
// have any configured recurring transactions in their config file, but only
// if the openFileName input is not an empty string.
func EmptyConfigLoadSuccessDialog(win *gtk.ApplicationWindow, openFileName string) {
	if openFileName != "" {
		m := fmt.Sprintf(
			"Success! Loaded file \"%v\" successfully. The configuration was empty, so a sample recurring transaction has been added.",
			openFileName,
		)
		d := gtk.MessageDialogNew(
			win,
			gtk.DIALOG_MODAL,
			gtk.MESSAGE_INFO,
			gtk.BUTTONS_OK,
			"%s",
			m,
		)
		log.Println(m)
		d.Run()
		d.Destroy()
	}
}

func GetMainWindowRootElements(application *gtk.Application) (
	*gtk.ApplicationWindow,
	*gtk.Box,
	*gtk.HeaderBar,
	*gtk.MenuButton,
	*glib.Menu,
) {
	win, err := gtk.ApplicationWindowNew(application)
	if err != nil {
		log.Fatal("unable to create window:", err)
	}
	rootBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, c.UISpacer)
	if err != nil {
		log.Fatal("unable to create root box:", err)
	}

	header, err := gtk.HeaderBarNew()
	if err != nil {
		log.Fatal("unable to create header bar:", err)
	}

	// Create a new menu button
	mbtn, err := gtk.MenuButtonNew()
	if err != nil {
		log.Fatal("unable to create menu button:", err)
	}

	// Set up the menu model for the button
	menu := glib.MenuNew()
	if menu == nil {
		log.Fatal("unable to create menu (nil pointer)")
	}

	// Actions with the prefix 'app' reference actions on the application
	// Actions with the prefix 'win' reference actions on the current window (specific to ApplicationWindow)
	// Other prefixes can be added to widgets via InsertActionGroup
	// example:
	// menu.Append("Custom Panic", "custom.panic")
	menu.Append("Save", "fin.saveOpenConfig")
	menu.Append("Save as...", "fin.saveConfig")
	menu.Append("Open...", "fin.loadConfigCurrentWindow")
	menu.Append("Open in new window...", "fin.loadConfigNewWindow")
	menu.Append("Save results...", "fin.saveResults")
	menu.Append("Copy results to clipboard", "fin.copyResults")
	menu.Append("Show statistics", "fin.getStats")
	menu.Append("New Window", "app.new")
	menu.Append("Close Window", "win.close")
	menu.Append("Quit", "app.quit")

	return win, rootBox, header, mbtn, menu
}

// TODO: refactor dialog code
// TODO: clean up logging
func SaveOpenConf(win *gtk.ApplicationWindow, header *gtk.HeaderBar, openFileName *string, userTX *[]lib.TX) {
	// write the config to the target file path
	err := lib.SaveConfig(*openFileName, *userTX)
	if err != nil {
		m := fmt.Sprintf(
			"Failed to save config to file \"%v\": %v",
			openFileName,
			err.Error(),
		)
		d := gtk.MessageDialogNew(
			win,
			gtk.DIALOG_MODAL,
			gtk.MESSAGE_ERROR,
			gtk.BUTTONS_OK,
			"%s",
			m,
		)
		log.Println(m)
		d.Run()
		d.Destroy()
		return
	}
	header.SetSubtitle(*openFileName)
}

// TODO: refactor dialog code
// TODO: clean up logging
func SaveConfAs(win *gtk.ApplicationWindow, header *gtk.HeaderBar, openFileName *string, userTX *[]lib.TX) {
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
	p.Connect("close", func(pdialog *gtk.FileChooserDialog) {
		pdialog.Close()
	})
	p.Connect("response", func(dialog *gtk.FileChooserDialog, resp int) {
		if resp == int(gtk.RESPONSE_OK) {
			// folder, _ := dialog.FileChooser.GetCurrentFolder()
			// GetFilename includes the full path and file name
			*openFileName = dialog.FileChooser.GetFilename()
			// write the config to the target file path
			err := lib.SaveConfig(*openFileName, *userTX)
			if err != nil {
				m := fmt.Sprintf(
					"Failed to save config to file \"%v\": %v",
					*openFileName,
					err.Error(),
				)
				d := gtk.MessageDialogNew(
					win,
					gtk.DIALOG_MODAL,
					gtk.MESSAGE_ERROR,
					gtk.BUTTONS_OK,
					"%s",
					m,
				)
				log.Println(m)
				d.Run()
				d.Destroy()
				return
			}
			header.SetSubtitle(*openFileName)
		}
		p.Close()
	})
	p.Dialog.ShowAll()
}

// TODO: refactor dialog code
// TODO: clean up logging
func SaveResults(win *gtk.ApplicationWindow, header *gtk.HeaderBar, latestResults *[]lib.Result) {
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
			err := lib.SaveResultsCSV(f, latestResults)
			if err != nil {
				m := fmt.Sprintf(
					"Failed to save results as CSV to file \"%v\": %v",
					f,
					err.Error(),
				)
				d := gtk.MessageDialogNew(
					win,
					gtk.DIALOG_MODAL,
					gtk.MESSAGE_ERROR,
					gtk.BUTTONS_OK,
					"%s",
					m,
				)
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

// TODO: refactor dialog code
// TODO: clean up logging
func CopyResults(win *gtk.ApplicationWindow, header *gtk.HeaderBar, latestResults *[]lib.Result) {
	r := lib.GetResultsCSVString(latestResults)
	clipboard, err := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
	if err != nil {
		log.Print("failed to get clipboard", err.Error())
	}
	clipboard.SetText(r)
	m := fmt.Sprintf(
		"Success! Copied %v result records to the clipboard.",
		len(*latestResults),
	)
	d := gtk.MessageDialogNew(win,
		gtk.DIALOG_MODAL,
		gtk.MESSAGE_INFO,
		gtk.BUTTONS_OK,
		"%s",
		m,
	)
	log.Println(m)
	d.Run()
	d.Destroy()
}

// TODO: refactor dialog code
// TODO: clean up logging
func GetStats(win *gtk.ApplicationWindow, latestResults *[]lib.Result) {
	stats, err := lib.GetStats(*latestResults)
	if err != nil {
		d := gtk.MessageDialogNew(
			win,
			gtk.DIALOG_MODAL,
			gtk.MESSAGE_WARNING,
			gtk.BUTTONS_OK,
			err.Error(),
		)
		log.Println(err.Error())
		d.Run()
		d.Destroy()
		return
	}
	m := fmt.Sprintf("%v\n\nWould you like to copy these stats to your clipboard?", stats)
	d := gtk.MessageDialogNew(win,
		gtk.DIALOG_MODAL,
		gtk.MESSAGE_INFO,
		gtk.BUTTONS_YES_NO,
		m,
	)
	log.Println(stats)
	resp := d.Run()
	d.Destroy()
	if resp == gtk.RESPONSE_YES {
		clipboard, err := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
		if err != nil {
			m = fmt.Sprintf("Failed to access the clipboard: %v", err.Error())
			d := gtk.MessageDialogNew(
				win,
				gtk.DIALOG_MODAL,
				gtk.MESSAGE_ERROR,
				gtk.BUTTONS_OK,
				m,
			)
			log.Println(m)
			d.Run()
			d.Destroy()
			return
		}
		clipboard.SetText(stats)
		m = ("Success! Copied the stats to the clipboard.")
		d := gtk.MessageDialogNew(
			win,
			gtk.DIALOG_MODAL,
			gtk.MESSAGE_INFO,
			gtk.BUTTONS_OK,
			m,
		)
		log.Println(m)
		d.Run()
		d.Destroy()
		return
	}
}

// GetStructuralComponents builds and returns pointers to the essential
// components that form the foundation for the window, regardless of which tab
// you're viewing. Returns a notebook and a grid.
func GetStructuralComponents(ws *state.WinState) (*gtk.Notebook, *gtk.Grid) {
	nb, err := gtk.NotebookNew()
	ws.Notebook = nb
	if err != nil {
		log.Fatal("failed to create notebook:", err)
	}

	grid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("failed to create grid:", err)
	}

	grid.SetOrientation(gtk.ORIENTATION_VERTICAL)
	ws.Notebook.SetHExpand(true)
	ws.Notebook.SetVExpand(true)

	return nb, grid
}

// SetupNotebookPages creates the Config and Results tabs by using pointers
// to existing ListStores (and lots of other things from the provided WinState)
// to enable each of the shown tab pages to show what they're supposed to show.
// This is an important function!
func SetupNotebookPages(ws *state.WinState) (*gtk.Grid, *gtk.Grid) {
	var err error

	configGrid, configSw, configTreeView, configTab := GetConfigBaseComponents(ws)
	ws.ConfigScrolledWindow = configSw
	ws.ConfigTreeView = configTreeView

	*ws.Results, err = lib.GenerateResultsFromDateStrings(
		ws.TX,
		ws.StartingBalance,
		ws.StartDate,
		ws.EndDate,
	)
	if err != nil {
		log.Fatal("failed to generate results from date strings", err.Error())
	}

	resultsGrid, label, err := GenerateResultsTab(
		ws.TX,
		*ws.Results,
		ws.ResultsListStore,
	)
	if err != nil {
		log.Fatalf("failed to generate results tab: %v", err.Error())
	}

	ws.Notebook.AppendPage(configGrid, configTab)
	ws.Notebook.AppendPage(resultsGrid, label)

	return configGrid, resultsGrid
}

func SetWinIcon(ws *state.WinState, embeddedIconFS embed.FS) {
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
	ws.Win.SetIcon(pb)
}
