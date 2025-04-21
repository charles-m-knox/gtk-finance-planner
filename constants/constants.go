package constants

// var DefaultConfFilePath = filepath.FromSlash(".config/finance-planner")

const (
	// Warning: do not remove this line; the makefile/build script relies on it
	VERSION      = "0.1.5"
	AboutMessage = "Finance Planner\n\nA way to manage recurring transactions.\n\nSource code: https://github.com/charles-m-knox/gtk-finance-planner"

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

	GtkAppID = "com.charlesmknox.gtk-finance-planner"

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

// This is the app icon but encoded via:
// cat assets/icon-128.png | base64
const AppIconBase64Png = `
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
AAAAAElFTkSuQmCC
`
