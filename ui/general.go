package ui

import (
	"fmt"
	"log"
	"strconv"

	c "finance-planner/constants"
	"finance-planner/lib"
	"finance-planner/state"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func createCheckboxColumn(title string, id int, radio bool, listStore *gtk.ListStore, ws *state.WinState) (tvc *gtk.TreeViewColumn, err error) {
	cellRenderer, err := gtk.CellRendererToggleNew()
	if err != nil {
		return tvc, fmt.Errorf("unable to create checkbox column renderer: %v", err.Error())
	}
	cellRenderer.SetActive(true)
	cellRenderer.SetRadio(radio)
	cellRenderer.SetActivatable(true)
	cellRenderer.SetVisible(true)
	// TODO: consider refactoring this a little bit
	cellRenderer.Connect("toggled", func(a *gtk.CellRendererToggle, path string) {
		// TODO: using nested structures results in a path that looks
		// like 1:2:5 -  parse accordingly
		i, err := strconv.ParseInt(path, 10, 64)
		if err != nil {
			log.Printf("failed to parse path \"%v\" as an int: %v", path, err.Error())
		}

		if lib.IsWeekday(c.ConfigColumns[id]) {
			weekday := lib.WeekdayIndex[c.ConfigColumns[id]]
			(*ws.TX)[i].Weekdays = lib.ToggleDayFromWeekdays((*ws.TX)[i].Weekdays, weekday)

			listStore.ForEach(func(model *gtk.TreeModel, searchPath *gtk.TreePath, iter *gtk.TreeIter) bool {
				if searchPath.String() == path {
					listStore.Set(
						iter,
						[]int{id},
						[]interface{}{
							(*ws.TX)[i].DoesTXHaveWeekday(weekday),
						})
					return true
				}
				return false
			})
			UpdateResults(ws, false)
			// note: calling SyncConfigListStore is unnecessary here, because the
			// above listStore.ForEach query actually syncs it for us. Also,
			// calling SyncConfigListStore actually causes some annoying UI behavior.
			// err := SyncConfigListStore(txs, listStore)
			// if err != nil {
			// 	log.Printf("failed to sync list store: %v", err.Error())
			// }
		} else if c.ConfigColumns[id] == c.ColumnActive {
			(*ws.TX)[i].Active = !(*ws.TX)[i].Active
			listStore.ForEach(func(model *gtk.TreeModel, searchPath *gtk.TreePath, iter *gtk.TreeIter) bool {
				if searchPath.String() == path {
					listStore.Set(
						iter,
						[]int{id},
						[]interface{}{(*ws.TX)[i].Active})
					return true
				}
				return false
			})
			UpdateResults(ws, false)
			// note: calling SyncConfigListStore is unnecessary here, because the
			// above listStore.ForEach query actually syncs it for us. Also,
			// calling SyncConfigListStore actually causes some annoying UI behavior.
			// err := SyncConfigListStore(txs, listStore)
			// if err != nil {
			// 	log.Printf("failed to sync list store: %v", err.Error())
			// }
		}
	})

	column, err := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "active", id)
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
		// TODO: create helper function that creates a fresh TX
		*userTX = []lib.TX{{
			Order:     1,
			Amount:    -500,
			Active:    true,
			Name:      "New",
			Frequency: "WEEKLY",
			Interval:  1,
		}}
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
