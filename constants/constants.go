package constants

const (
	Day     = "Day"
	Weekly  = "Weekly"
	Monthly = "Monthly"
	Yearly  = "Yearly"

	UISpacer = 10 // allows consistent spacing between all elements

	DefaultConfFileName = "conf.json"
	DefaultConfFilePath = ".config/finance-planner"
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
