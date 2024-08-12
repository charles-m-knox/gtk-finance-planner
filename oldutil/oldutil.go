package oldutil

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"git.cmcode.dev/cmcode/gtk-finance-planner/constants"

	lib "git.cmcode.dev/cmcode/finance-planner-lib"

	"github.com/gotk3/gotk3/gtk"
	"gopkg.in/yaml.v3"
)

// FPConf is a configuration that is compatible with my other financial planning
// applications.
type Profile struct {
	TX []lib.TX `yaml:"transactions"`
	// Name            string   `yaml:"name"`
	// Modified        bool     `yaml:"-"`
	// SelectedRow     int      `yaml:"selectedRow"`
	// SelectedColumn  int      `yaml:"selectedColumn"`
	// StartingBalance string   `yaml:"startingBalance"`
	// StartDay        string   `yaml:"startDay"`
	// StartMonth      string   `yaml:"startMonth"`
	// StartYear       string   `yaml:"startYear"`
	// EndDay          string   `yaml:"endDay"`
	// EndMonth        string   `yaml:"endMonth"`
	// EndYear         string   `yaml:"endYear"`
}

// FPConf is a configuration that is compatible with my other financial planning
// applications. Note that gtk-finance-planner only supports one profile at
// this time.
type FPConf struct {
	Profiles []Profile `yaml:"profiles"`
}

var WeekdayIndex = map[string]int{
	"Monday":    0,
	"Tuesday":   1,
	"Wednesday": 2,
	"Thursday":  3,
	"Friday":    4,
	"Saturday":  5,
	"Sunday":    6,
}

// IsWeekday determines if a provided string corresponds to a weekday.
// TODO: refactor using constants
func IsWeekday(weekday string) bool {
	if weekday == "Monday" {
		return true
	}
	if weekday == "Tuesday" {
		return true
	}
	if weekday == "Wednesday" {
		return true
	}
	if weekday == "Thursday" {
		return true
	}
	if weekday == "Friday" {
		return true
	}
	if weekday == "Saturday" {
		return true
	}
	if weekday == "Sunday" {
		return true
	}
	return false
}

func GetTXIDByListStorePath(ls *gtk.ListStore, path *gtk.TreePath) (string, error) {
	ps := path.String()

	iter, err := ls.GetIter(path)
	if err != nil {
		return "", fmt.Errorf("failed to get iter for path %v: %v", ps, err.Error())
	}

	value, err := ls.GetValue(iter, constants.COLUMN_ID)
	if err != nil {
		return "", fmt.Errorf("failed to get value for path %v: %v", ps, err.Error())
	}

	id, err := value.GetString()
	if err != nil {
		return "", fmt.Errorf("failed to get string-value for id, by path %v: %v", ps, err.Error())
	}

	return id, nil
}

// LoadConfig can load from the same config files that finance-planner-tui
// uses, but it cannot save to that same file currently.
func LoadConfig(file string) (txs []lib.TX, err error) {
	if strings.HasSuffix(file, ".yml") || strings.HasSuffix(file, ".yaml") {
		b, err := os.ReadFile(file)
		if err != nil {
			return txs, fmt.Errorf("failed to read config json: %v", err.Error())
		}

		fpc := FPConf{}

		err = yaml.Unmarshal(b, &fpc)
		if err != nil {
			return txs, fmt.Errorf("failed to unmarshal config yaml: %v", err.Error())
		}

		if len(fpc.Profiles) == 0 {
			return txs, errors.New("config file %v has no profiles")
		}

		return fpc.Profiles[0].TX, nil
	}

	txJSON, err := os.ReadFile(file)
	if err != nil {
		return txs, fmt.Errorf("failed to read config json: %v", err.Error())
	}

	err = json.Unmarshal(txJSON, &txs)
	if err != nil {
		return txs, fmt.Errorf("failed to unmarshal config json: %v", err.Error())
	}

	// apply an automatic order to each of the transactions, starting from 1,
	// since the 0-value is default when undefined
	// for i := range txs {
	// 	if txs[i].Order == 0 {
	// 		txs[i].Order = i + 1
	// 	}
	// }

	return
}

func SaveConfig(file string, txs []lib.TX) error {
	txJSON, err := json.Marshal(txs)
	if err != nil {
		return fmt.Errorf("failed to parse tx json: %v", err.Error())
	}
	dir := path.Dir(file)
	log.Println(dir)
	err = os.MkdirAll(dir, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create parent directory \"%v\" for saving tx json: %v", dir, err.Error())
	}
	err = os.WriteFile(file, txJSON, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write json to file %v: %v", file, err.Error())
	}
	return nil
}

func SaveResultsCSV(file string, results *[]lib.Result) (err error) {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	for _, r := range *results {
		var record []string
		record = append(record, lib.GetNowDateString(r.Date))
		record = append(record, lib.FormatAsCurrency(r.Balance))
		record = append(record, lib.FormatAsCurrency(r.CumulativeIncome))
		record = append(record, lib.FormatAsCurrency(r.CumulativeExpenses))
		record = append(record, lib.FormatAsCurrency(r.DayExpenses))
		record = append(record, lib.FormatAsCurrency(r.DayIncome))
		record = append(record, lib.FormatAsCurrency(r.DayNet))
		record = append(record, lib.FormatAsCurrency(r.DiffFromStart))
		record = append(record, r.DayTransactionNames)
		_ = w.Write(record)
	}
	w.Flush()
	return nil
}

// TODO: refactor w/ constants for the color hex code values
func CurrencyMarkup(input int) string {
	currency := lib.FormatAsCurrency(input)
	if input == 0 {
		return fmt.Sprintf(`<i><span foreground="#CCCCCC">%v</span></i>`, currency)
	}
	if input > 0 {
		return fmt.Sprintf(`<span foreground="#c2e1b5">%v</span>`, currency)
	}
	if input < 0 {
		return fmt.Sprintf(`<span foreground="#dda49e">%v</span>`, currency)
	}

	return currency
}

// MarkupColorSequence takes an input string slice and converts it into a semi-
// colon separated string, as well as slowly shifting the color of each semi-
// colon separated value in the string itself to help users differentiate the
// different entries in the CSV string.
func MarkupColorSequence(input []string) string {
	result := new(strings.Builder)
	if len(input) > 0 {
		result.WriteString(fmt.Sprintf("(%v) ", len(input)))
	}
	for i, name := range input {
		colorSequenceIndex := i % len(constants.ResultsTXNameColorSequences)
		result.WriteString(fmt.Sprintf(`<u><span foreground="%v">%v</span></u>; `, constants.ResultsTXNameColorSequences[colorSequenceIndex], name))
	}
	return result.String()
}

// GetCSVString produces a simple semi-colon-separated value string.
func GetCSVString(input []string) string {
	result := new(strings.Builder)
	if len(input) > 0 {
		result.WriteString(fmt.Sprintf("(%v) ", len(input)))
	}
	for _, name := range input {
		result.WriteString(fmt.Sprintf(`%v; `, name))
	}
	return result.String()
}

// GetListStoreValue retrieves a value from a GTK list store during iteration
// of a tree.
func GetListStoreValue(
	ls *gtk.ListStore,
	iter *gtk.TreeIter,
	col int,
) (result interface{}, err error) {
	gv, err := ls.GetValue(iter, col)
	if err != nil {
		return result, fmt.Errorf(
			"failed to get value from config list store: %v",
			err.Error(),
		)
	}

	// marshal the value into a Go-native data type
	val, err := gv.GoValue()
	if err != nil {
		return result, fmt.Errorf(
			"failed to get val string: %v",
			err.Error(),
		)
	}

	return val, nil
}

// GetTXIDByGTKIndex attempts to find the ID corresponding to the provided
// index value `i` within the GTK ListStore. For example, if row 5 is selected,
// this will return the ID of the TX at row 5 (as currently displayed).
// Returns an empty string if the value is not found.
// TODO: This is unused but may be useful in the future.
func GetTXIDByGTKIndex(ls *gtk.ListStore, i int) string {
	id := ""

	iterFn := func(model *gtk.TreeModel, searchPath *gtk.TreePath, iter *gtk.TreeIter) bool {
		if searchPath.String() != fmt.Sprintf("%v", i) {
			return false
		}

		val, err := GetListStoreValue(ls, iter, constants.COLUMN_ID)
		if err != nil {
			log.Printf("get TX ID by list store index: %v", err.Error())
		}

		id = val.(string)

		return true
	}

	ls.ForEach(iterFn)

	return id
}
