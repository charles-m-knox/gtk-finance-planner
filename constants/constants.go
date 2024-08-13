package constants

// var DefaultConfFilePath = filepath.FromSlash(".config/finance-planner")

const (
	// Warning: do not remove this line; the makefile/build script relies on it
	VERSION      = "0.1.1"
	AboutMessage = "Finance Planner\n\nA way to manage recurring transactions.\n\nSource code: https://git.cmcode.dev/cmcode/gtk-finance-planner"

	APP_CONF_DIR      = "finance-planner"
	APP_CONF_FILENAME = "conf.json"

	Day     = "Day"
	Weekly  = "Weekly"
	Monthly = "Monthly"
	Yearly  = "Yearly"

	WEEKLY  = "WEEKLY"
	MONTHLY = "MONTHLY"
	YEARLY  = "YEARLY"

	Y = "Y"
	W = "W"
	M = "M"

	New = "New"

	FinancialPlanner = "Financial Planner"

	UISpacer = 10 // allows consistent spacing between all elements

	// TODO: get a proper reverse fqdn for this eventually
	GtkAppID = "dev.cmcode.gtk-finance-planner"

	IconAssetPath = "assets/icon-128.png"

	BalanceInputPlaceholderText = "$500.00 - Enter a balance to start with."
	FullGridWidth               = 2
	HalfGridWidth               = 1
	ScrolledWindowGridHeight    = 4
	ControlsGridHeight          = 1

	GtkSignalClicked      = "clicked"
	GtkSignalActivate     = "activate"
	GtkSignalChanged      = "changed"
	GtkSignalFocusOut     = "focus-out-event"
	GtkSignalEditingStart = "editing-started"
	GtkSignalEdited       = "edited"

	ActionGroupFin = "fin"
	ActionGroupApp = "app"
	ActionGroupWin = "win"

	ActionClose                   = "close"
	ActionNew                     = "new"
	ActionQuit                    = "quit"
	ActionSaveConfig              = "saveConfig"
	ActionSaveOpenConfig          = "saveOpenConfig"
	ActionSaveResults             = "saveResults"
	ActionCopyResults             = "copyResults"
	ActionLoadConfigCurrentWindow = "loadConfigCurrentWindow"
	ActionLoadConfigNewWindow     = "loadConfigNewWindow"
	ActionGetStats                = "getStats"
	ActionAbout                   = "showAboutDialog"

	MenuItemSave          = "Save"
	MenuItemSaveAs        = "Save as..."
	MenuItemOpen          = "Open..."
	MenuItemOpenNewWindow = "Open in new window..."
	MenuItemSaveResults   = "Save results..."
	MenuItemCopyResults   = "Copy results to clipboard"
	MenuItemShowStats     = "Show statistics"
	MenuItemAbout         = "About"
	MenuItemNewWindow     = "New Window"
	MenuItemCloseWindow   = "Close Window"
	MenuItemQuit          = "Quit"

	HideInactiveBtnLabel = "_Hide inactive"
	CloneBtnLabel        = "_Clone"
	AddBtnLabel          = "_+"
	DelBtnLabel          = "_-"
	ConfigTabLabel       = "Config"

	// user-facing messages
	MsgInvalidDateInput          = "Enter a valid date in the format YYYY-MM-DD."
	MsgStartBalanceCannotBeEmpty = "Enter a non-empty currency-like value for the starting balance."
	MsgInvalidRecurrence         = "Please enter one of the following values: y/m/w/monthly/weekly/yearly"

	// error codes - generate new ones with "uuidgen | cut -b 1-6"
	ErrorCodeSyncConfigListStore                      = "9a0fab"
	ErrorCodeSyncConfigListStoreAfterColumnSortChange = "a6bbb2"
)

const (
	TAB_CONFIG = iota
	TAB_RESULTS
)

// A zebra-like pattern helps visually parse values in the "day transaction
// names" column in the results view. These are blue-ish colors.
var ResultsTXNameColorSequences = []string{
	"#d9e7fd",
	// "#b4cffb",
	// "#8eb7f9",
	// "#699ff7",
	"#4387f5",
}

// results page values

const (
	ColumnDate                = "Date"
	ColumnBalance             = "Balance"
	ColumnCumulativeIncome    = "CumulativeIncome"
	ColumnCumulativeExpenses  = "CumulativeExpenses"
	ColumnDayExpenses         = "DayExpenses"
	ColumnDayIncome           = "DayIncome"
	ColumnDayNet              = "DayNet"
	ColumnDiffFromStart       = "DiffFromStart"
	ColumnDayTransactionNames = "DayTransactionNames"
)

const (
	ColumnDateIndex = iota
	ColumnBalanceIndex
	ColumnCumulativeIncomeIndex
	ColumnCumulativeExpensesIndex
	ColumnDayExpensesIndex
	ColumnDayIncomeIndex
	ColumnDayNetIndex
	ColumnDiffFromStartIndex
	ColumnDayTransactionNamesIndex
)

var ResultsColumns = []string{
	ColumnDate,
	ColumnBalance,
	ColumnCumulativeIncome,
	ColumnCumulativeExpenses,
	ColumnDayExpenses,
	ColumnDayIncome,
	ColumnDayNet,
	ColumnDiffFromStart,
	ColumnDayTransactionNames,
}

// make ResultsColumnsIndexes the same length as the "columns" variable
var ResultsColumnsIndexes = []int{
	ColumnDateIndex,
	ColumnBalanceIndex,
	ColumnCumulativeIncomeIndex,
	ColumnCumulativeExpensesIndex,
	ColumnDayExpensesIndex,
	ColumnDayIncomeIndex,
	ColumnDayNetIndex,
	ColumnDiffFromStartIndex,
	ColumnDayTransactionNamesIndex,
}

// values for the config page

const (
	// ColumnOrder     = "Order"

	ColumnAmount    = "Amount"    // int in cents; 500 = $5.00
	ColumnActive    = "Active"    // bool true/false
	ColumnName      = "Name"      // editable string
	ColumnFrequency = "Frequency" // dropdown, monthly/daily/weekly/yearly
	ColumnInterval  = "Interval"  // integer, occurs every x frequency
	ColumnMonday    = "Monday"    // bool
	ColumnTuesday   = "Tuesday"   // bool
	ColumnWednesday = "Wednesday" // bool
	ColumnThursday  = "Thursday"  // bool
	ColumnFriday    = "Friday"    // bool
	ColumnSaturday  = "Saturday"  // bool
	ColumnSunday    = "Sunday"    // bool
	ColumnStarts    = "Starts"    // string
	ColumnEnds      = "Ends"      // string
	ColumnNote      = "Note"      // editable string
	ColumnID        = "ID"
	ColumnCreatedAt = "CreatedAt"
	ColumnUpdatedAt = "UpdatedAt"

	WeekdayMonday    = "Monday"
	WeekdayTuesday   = "Tuesday"
	WeekdayWednesday = "Wednesday"
	WeekdayThursday  = "Thursday"
	WeekdayFriday    = "Friday"
	WeekdaySaturday  = "Saturday"
	WeekdaySunday    = "Sunday"
)

var ConfigColumns = []string{
	// ColumnOrder,     // int
	ColumnAmount,    // int in cents; 500 = $5.00
	ColumnActive,    // bool true/false
	ColumnName,      // editable string
	ColumnFrequency, // dropdown, monthly/daily/weekly/yearly
	ColumnInterval,  // integer, occurs every x frequency
	ColumnMonday,    // bool
	ColumnTuesday,   // bool
	ColumnWednesday, // bool
	ColumnThursday,  // bool
	ColumnFriday,    // bool
	ColumnSaturday,  // bool
	ColumnSunday,    // bool
	ColumnStarts,    // string
	ColumnEnds,      // string
	ColumnNote,      // editable string
	ColumnID,
	ColumnCreatedAt,
	ColumnUpdatedAt,
}

var Weekdays = []string{
	WeekdayMonday,
	WeekdayTuesday,
	WeekdayWednesday,
	WeekdayThursday,
	WeekdayFriday,
	WeekdaySaturday,
	WeekdaySunday,
}

const (
	WeekdayMondayInt = iota
	WeekdayTuesdayInt
	WeekdayWednesdayInt
	WeekdayThursdayInt
	WeekdayFridayInt
	WeekdaySaturdayInt
	WeekdaySundayInt
)

const (
	// COLUMN_ORDER     = iota // int
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
	COLUMN_ID               // non-editable strings
	COLUMN_CREATEDAT        // non-editable strings
	COLUMN_UPDATEDAT        // non-editable strings
)

const (
	None = "none"
	Desc = "Desc"
	Asc  = "Asc"
)
