package lib

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	totp "github.com/pquerna/otp/totp"
	"github.com/teambition/rrule-go"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	Day     = "Day"
	Weekly  = "Weekly"
	Monthly = "Monthly"
	Yearly  = "Yearly"
)

var WeekdayIndex = map[string]int{
	"Monday":    0,
	"Tuesday":   1,
	"Wednesday": 2,
	"Thursday":  3,
	"Friday":    4,
	"Saturday":  5,
	"Sunday":    6,
}

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

func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%02d/%02d/%d", month, day, year)
}

func FormatAsCurrency(a int) string {
	// convert to float and dump as currency string
	//  TODO: print the integer and clip the last two digits instead of
	// using floats
	amt := float64(a)
	amt = amt / 100
	p := message.NewPrinter(language.English)
	return p.Sprintf("$%.2f", amt)
}

type TX struct { // transaction
	Order  int    `json:"order"`  // manual ordering
	Amount int    `json:"amount"` // in cents; 500 = $5.00
	Active bool   `json:"active"`
	Name   string `json:"name"`
	Note   string `json:"note"`
	// for examples of rrules:
	// https://github.com/teambition/rrule-go/blob/f71921a2b0a18e6e73c74dea155f3a549d71006d/rrule.go#L91
	// https://github.com/teambition/rrule-go/blob/master/rruleset_test.go
	// https://labix.org/python-dateutil/#head-88ab2bc809145fcf75c074817911575616ce7caf
	RRule string `json:"rrule"`
	// for when users don't want to use the rrules:
	Frequency   string `json:"frequency"`
	Interval    int    `json:"interval"`
	Weekdays    []int  `json:"weekdays"` // monday starts on 0
	StartsDay   int    `json:"startsDay"`
	StartsMonth int    `json:"startsMonth"`
	StartsYear  int    `json:"startsYear"`
	EndsDay     int    `json:"endsDay"`
	EndsMonth   int    `json:"endsMonth"`
	EndsYear    int    `json:"endsYear"`
}

type PreCalculatedResult struct {
	Date                  time.Time
	DayTransactionNames   []string
	DayTransactionAmounts []int
}

type Result struct { // csv/table output row
	Record              int
	Date                time.Time
	Balance             int
	CumulativeIncome    int
	CumulativeExpenses  int
	DayExpenses         int
	DayIncome           int
	DayNet              int
	DayTransactionNames string
	DiffFromStart       int
}

type User struct {
	ID       string // uuid
	Password string
	TOTP     string
}

func AddDayToWeekdays(weekdays []int, weekday int) []int {
	if weekday < 0 || weekday > 6 {
		return weekdays
	}
	for i := range weekdays {
		if weekdays[i] == weekday {
			return weekdays // do nothing
		}
	}
	weekdays = append(weekdays, weekday)
	return weekdays
}

func RemoveDayFromWeekdays(weekdays []int, weekday int) []int {
	if weekday < 0 || weekday > 6 {
		return weekdays
	}
	for i := range weekdays {
		if weekdays[i] != weekday {
			weekdays = append(weekdays, weekday)
		}
	}
	return weekdays
}

func ToggleDayFromWeekdays(weekdays []int, weekday int) []int {
	if weekday < 0 || weekday > 6 {
		return weekdays
	}
	foundWeekday := false
	returnValue := []int{}
	for i := range weekdays {
		if weekdays[i] == weekday {
			foundWeekday = true
		} else {
			returnValue = append(returnValue, weekdays[i])
		}
	}
	if !foundWeekday {
		returnValue = append(returnValue, weekday)
	}
	sort.Ints(returnValue)
	return returnValue
}

func GetTOTP(userID string, issuer string) (string, string, error) {
	key, err := totp.Generate(
		totp.GenerateOpts{
			Issuer: issuer,
			AccountName: fmt.Sprintf(
				"%v - %v",
				userID,
				issuer,
			),
		},
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to gen totp: %v", err.Error())
	}

	return key.Secret(), key.String(), nil
}

func GetResults(tx []TX, startDate time.Time, endDate time.Time, startBalance int) ([]Result, error) {
	if startDate.After(endDate) {
		return []Result{}, fmt.Errorf("start date is after end date")
	}

	// start by quickly generating an index of every single date from startDate to endDate
	dates := make(map[int64]Result)
	preCalculatedDates := make(map[int64]PreCalculatedResult)

	r, err := rrule.NewRRule(
		rrule.ROption{
			Freq:    rrule.DAILY,
			Dtstart: startDate,
			Until:   endDate,
		},
	)
	if err != nil {
		return []Result{}, fmt.Errorf("failed to construct rrule for results date window: %v", err.Error())
	}
	allDates := r.All()

	for i, dt := range allDates {
		dtInt := dt.Unix()
		dates[dtInt] = Result{
			Record: i,
			Date:   dt,
		}
		preCalculatedDates[dtInt] = PreCalculatedResult{
			Date: dt,
		}
	}

	emptyDate := time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)

	// iterate over every TX definition, starting with its start date
	for _, txi := range tx {
		if !txi.Active {
			continue
		}

		var allOccurrences []time.Time

		if txi.RRule != "" {
			s, err := rrule.StrToRRuleSet(txi.RRule)
			if err != nil {
				return []Result{}, fmt.Errorf(
					"failed to process rrule for tx %v: %v",
					txi.Name,
					err.Error(),
				)
			}

			allOccurrences = s.Between(
				startDate,
				endDate,
				true,
			)
		} else {
			txiStartsDate := time.Date(txi.StartsYear, time.Month(txi.StartsMonth), txi.StartsDay, 0, 0, 0, 0, time.UTC)
			txiEndsDate := time.Date(txi.EndsYear, time.Month(txi.EndsMonth), txi.EndsDay, 0, 0, 0, 0, time.UTC)
			// input validation: if the end date for the transaction definition is after
			// the final end date, then just use the ending date.
			// also, if the transaction definition's end date is unset (equal to emptyDate),
			// then default to the ending date as well
			if txiEndsDate.After(endDate) || txiEndsDate == emptyDate {
				txiEndsDate = endDate
			}
			// input validation: if the transaction definition's start date is
			// unset (equal to emptyDate), then default to the start date
			if txiStartsDate == emptyDate {
				txiStartsDate = startDate
			}
			// convert the user input frequency to a value that rrule lib
			// will accept
			freq := rrule.DAILY
			if txi.Frequency == rrule.YEARLY.String() {
				freq = rrule.YEARLY
			} else if txi.Frequency == rrule.MONTHLY.String() {
				freq = rrule.MONTHLY
			}
			// convert the user input weekdays into a value that rrule lib will
			// accept
			weekdays := []rrule.Weekday{}
			for _, weekday := range txi.Weekdays {
				if weekday == rrule.MO.Day() {
					weekdays = append(weekdays, rrule.MO)
				} else if weekday == rrule.TU.Day() {
					weekdays = append(weekdays, rrule.TU)
				} else if weekday == rrule.WE.Day() {
					weekdays = append(weekdays, rrule.WE)
				} else if weekday == rrule.TH.Day() {
					weekdays = append(weekdays, rrule.TH)
				} else if weekday == rrule.FR.Day() {
					weekdays = append(weekdays, rrule.FR)
				} else if weekday == rrule.SA.Day() {
					weekdays = append(weekdays, rrule.SA)
				} else if weekday == rrule.SU.Day() {
					weekdays = append(weekdays, rrule.SU)
				}
			}
			// create the rule based on the input parameters from the user
			s, err := rrule.NewRRule(
				rrule.ROption{
					Freq:      freq,
					Interval:  txi.Interval,
					Dtstart:   txiStartsDate,
					Until:     txiEndsDate,
					Byweekday: weekdays,
				},
			)
			if err != nil {
				return []Result{}, fmt.Errorf(
					"failed to construct rrule for tx %v: %v",
					txi.Name,
					err.Error(),
				)
			}
			allOccurrences = s.Between(
				startDate,
				endDate,
				true,
			)
		}

		for _, dt := range allOccurrences {
			dtInt := dt.Unix()
			newResult := preCalculatedDates[dtInt]
			newResult.Date = dt
			newResult.DayTransactionAmounts = append(newResult.DayTransactionAmounts, txi.Amount)
			newResult.DayTransactionNames = append(newResult.DayTransactionNames, txi.Name)
			preCalculatedDates[dtInt] = newResult
		}
	}

	results := []Result{}
	for _, result := range dates {
		results = append(results, result)
	}

	sort.SliceStable(
		results,
		func(i, j int) bool {
			return results[j].Date.After(results[i].Date)
		},
	)

	// now that it's sorted, we can roll-out the calculations
	currentBalance := startBalance
	diff := 0
	cumulativeIncome := 0
	cumulativeExpenses := 0
	for i := range results {
		resultsDateInt := results[i].Date.Unix()
		// if for some reason not all transaction names and amounts match up,
		// exit now
		if len(preCalculatedDates[resultsDateInt].DayTransactionAmounts) != len(preCalculatedDates[resultsDateInt].DayTransactionNames) {
			return results, fmt.Errorf("there was a different number of transaction amounts versus transaction names for date %v", resultsDateInt)
		}

		// log.Printf("preCalculatedDates[results[i].Date]=%v, len=%v", preCalculatedDates[resultsDateInt], len(preCalculatedDates))

		for j := range preCalculatedDates[resultsDateInt].DayTransactionAmounts {
			// determine if the amount is an expense or income
			amt := preCalculatedDates[resultsDateInt].DayTransactionAmounts[j]
			if amt >= 0 {
				results[i].DayIncome += amt
				cumulativeIncome += amt
			} else {
				results[i].DayExpenses += amt
				cumulativeExpenses += amt
			}

			// basically just doing a join on a slice of strings, should
			// use the proper method for this in the future
			name := preCalculatedDates[resultsDateInt].DayTransactionNames[j]
			if results[i].DayTransactionNames == "" {
				results[i].DayTransactionNames = name
			} else {
				results[i].DayTransactionNames += fmt.Sprintf("; %v", name)
			}

			results[i].DayNet += amt
			diff += amt
			currentBalance += amt
			// log.Printf("day %v: amt=%v, date=%v", i, amt, resultsDateInt)
		}

		results[i].Balance = currentBalance
		results[i].CumulativeIncome = cumulativeIncome
		results[i].CumulativeExpenses = cumulativeExpenses
		results[i].DiffFromStart = diff
		// log.Printf("results[%v]=%v", i, results[i])
	}

	return results, nil
}

func GetNowDateString() string {
	now := time.Now()
	return fmt.Sprintf(
		"%v-%v-%v",
		now.Year(),
		int(now.Month()),
		now.Day(),
	)
}

func GetDefaultEndDateString() string {
	now := time.Now()
	return fmt.Sprintf(
		"%v-%v-%v",
		now.Year()+1,
		int(now.Month()),
		now.Day(),
	)
}

func ParseYearMonthDateString(input string) (int, int, int) {
	vals := strings.Split(input, "-")
	if len(vals) != 3 {
		return 0, 0, 0
	}
	yr, _ := strconv.ParseInt(vals[0], 10, 64)
	mo, _ := strconv.ParseInt(vals[1], 10, 64)
	day, _ := strconv.ParseInt(vals[2], 10, 64)
	return int(yr), int(mo), int(day)
}

func ParseDollarAmount(input string, assumePositive bool) int64 {
	cents := int64(0)
	whole := int64(0)
	multiplier := int64(-1)
	r := regexp.MustCompile(`[^\d.]*`)
	s := r.ReplaceAllString(input, "")
	// all values are assumed negative, unless it starts with a + character
	if strings.Index(input, "+") == 0 || strings.Index(input, "$+") == 0 || assumePositive {
		multiplier = int64(1)
	}
	// in the event that the user is entering the starting balance,
	// they may want to set a negative starting balance. So basically just reverse
	// from above logic, since the user will have to be typing a negative sign in front.
	if assumePositive && (strings.Index(input, "$-") == 0 || strings.Index(input, "-") == 0) {
		multiplier = int64(-1)
	}
	// check if the user entered a period
	ss := strings.Split(s, ".")
	if len(ss) == 2 {
		cents, _ = strconv.ParseInt(ss[1], 10, 64)
		// if the user types e.g. 10.2, they meant $10.20
		// but not if the value started with a 0
		if strings.Index(ss[1], "0") != 0 && cents < 10 {
			cents = cents * 10
		}
		// if they put in too many numbers, zero it out
		if cents >= 100 {
			cents = 0
		}
	}
	whole, _ = strconv.ParseInt(ss[0], 10, 64)
	// pi := strings.Index(input, ".")
	// if pi != -1 {
	// 	// split and treat the values accordingly
	// }

	// r := regexp.MustCompile(`[^-\d.]*`)
	// s := r.ReplaceAllString(input, "")
	log.Println("ParseDollarAmount", r, s, input)
	// i, _ := strconv.ParseInt(s, 10, 64)
	// padded := fmt.Sprintf("%6d", i)
	// j, _ := strconv.ParseInt(padded, 10, 64)

	// account for the negative case when re-combining the two values
	if whole < 0 {
		return multiplier * (whole*100 - cents)
	}
	return multiplier * (whole*100 + cents)
}

func RemoveTXAtIndex(txs []TX, i int) []TX {
	return append(txs[:i], txs[i+1:]...)
}

func GenerateResultsFromDateStrings(txs *[]TX, bal int, startDt string, endDt string) ([]Result, error) {
	now := time.Now()
	stYr, stMo, stDay := ParseYearMonthDateString(startDt)
	endYr, endMo, endDay := ParseYearMonthDateString(endDt)
	if startDt == "0-0-0" || startDt == "" {
		stYr = now.Year()
		stMo = int(now.Month())
		stDay = now.Day()
	}
	if endDt == "0-0-0" || endDt == "" {
		endYr = now.Year() + 1
		endMo = int(now.Month())
		endDay = now.Day()
	}
	res, err := GetResults(
		*txs,
		time.Date(stYr, time.Month(stMo), stDay, 0, 0, 0, 0, time.UTC),
		time.Date(endYr, time.Month(endMo), endDay, 0, 0, 0, 0, time.UTC),
		bal,
	)
	if err != nil {
		return []Result{}, fmt.Errorf("failed to get results: %v", err.Error())
	}

	return res, nil
}

func GetStats(results []Result) (string, error) {
	// daily average spend/income
	// monthly (30 day) average spend/income
	// yearly (365 day) average spend/income
	// later on:
	// amount spent in each month of the year
	count := len(results)
	i := 365
	if count > i {
		b := new(strings.Builder)
		b.WriteString("Here are some statistics about your finances.\n\n")

		dailySpendingAvg := results[i].CumulativeExpenses / i
		dailyIncomeAvg := results[i].CumulativeIncome / i

		b.WriteString(fmt.Sprintf(
			"Daily spending: %v\nDaily income: %v\nDaily net: %v",
			FormatAsCurrency(dailySpendingAvg),
			FormatAsCurrency(dailyIncomeAvg),
			FormatAsCurrency(dailySpendingAvg+dailyIncomeAvg),
		))
		moSpendingAvg := results[i].CumulativeExpenses / 12
		moIncomeAvg := results[i].CumulativeIncome / 12
		b.WriteString(fmt.Sprintf(
			"\nMonthly spending: %v\nMonthly income: %v\nMonthly net: %v",
			FormatAsCurrency(moSpendingAvg),
			FormatAsCurrency(moIncomeAvg),
			FormatAsCurrency(moSpendingAvg+moIncomeAvg),
		))
		yrSpendingAvg := results[i].CumulativeExpenses
		yrIncomeAvg := results[i].CumulativeIncome
		b.WriteString(fmt.Sprintf(
			"\nYearly spending: %v\nYearly income: %v\nYearly net: %v",
			FormatAsCurrency(yrSpendingAvg),
			FormatAsCurrency(yrIncomeAvg),
			FormatAsCurrency(yrSpendingAvg+yrIncomeAvg),
		))

		return b.String(), nil
	}
	return "", fmt.Errorf("You need at least one year between your start date and end date to get statistics about your finances.")
}
