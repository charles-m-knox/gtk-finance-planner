package constants

import "github.com/gotk3/gotk3/glib"

const (
	Day     = "Day"
	Weekly  = "Weekly"
	Monthly = "Monthly"
	Yearly  = "Yearly"

	FinancialPlanner = "Financial Planner"

	UISpacer = 10 // allows consistent spacing between all elements

	DefaultConfFileName = "conf.json"
	DefaultConfFilePath = ".config/finance-planner"

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

	ActionClose                   = "close"
	ActionNew                     = "new"
	ActionQuit                    = "quit"
	ActionSaveConfig              = "saveConfig"
	ActionSaveOpenConfig          = "saveOpenConfig"
	ActionSaveResults             = "saveResults"
	ActionCopyResults             = "copyResults"
	ActionLoadConfigCurrentWindow = "loadConfigCurrentWindow"
	ActionLoadConfigNewWindow     = "loadConfigNewWindow"

	HideInactiveBtnLabel = "_Hide inactive"
	CloneBtnLabel        = "_Clone"
	AddBtnLabel          = "_+"
	DelBtnLabel          = "_-"
	ConfigTabLabel       = "Config"
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
	ColumnOrder     = "Order"
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

	WeekdayMonday    = "Monday"
	WeekdayTuesday   = "Tuesday"
	WeekdayWednesday = "Wednesday"
	WeekdayThursday  = "Thursday"
	WeekdayFriday    = "Friday"
	WeekdaySaturday  = "Saturday"
	WeekdaySunday    = "Sunday"
)

var ConfigColumns = []string{
	ColumnOrder,     // int
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
	COLUMN_ORDER     = iota // int
	COLUMN_AMOUNT           // int in cents; 500 = $5.00
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
)

var ConfigColumnTypes = []glib.Type{
	glib.TYPE_STRING,
	glib.TYPE_STRING,
	glib.TYPE_BOOLEAN,
	glib.TYPE_STRING,
	glib.TYPE_STRING,
	glib.TYPE_STRING,
	glib.TYPE_BOOLEAN,
	glib.TYPE_BOOLEAN,
	glib.TYPE_BOOLEAN,
	glib.TYPE_BOOLEAN,
	glib.TYPE_BOOLEAN,
	glib.TYPE_BOOLEAN,
	glib.TYPE_BOOLEAN,
	glib.TYPE_STRING,
	glib.TYPE_STRING,
	glib.TYPE_STRING,
}

const (
	None = "none"
	Desc = "Desc"
	Asc  = "Asc"
)
