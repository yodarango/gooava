package models

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type ExpenseItem struct {
	Label  string  `json:"label"`
	Amount float64 `json:"amount"`
	Date   string  `json:"date"` // YYYY-MM-DD
}

type CategoryBreakdownItem struct {
	Category string        `json:"category"`
	Amount   float64       `json:"amount"`
	Color    string        `json:"color"`
	Items    []ExpenseItem `json:"items"`
}

type MonthBreakdown struct {
	Month     string                  `json:"month"` // "YYYY-MM"
	Label     string                  `json:"label"` // "Nov '26"
	Total     float64                 `json:"total"`
	Income    float64                 `json:"income"`
	Breakdown []CategoryBreakdownItem `json:"breakdown"`
}

type CategoryTotalToDate struct {
	Category           string        `json:"category"`
	Color              string        `json:"color"`
	TotalSpent         float64       `json:"total_spent"`
	PercentageOfIncome *float64      `json:"percentage_of_income"` // null when total income is 0
	Items              []ExpenseItem `json:"items"`
}

type SpendingBreakdownResponse struct {
	Months            []MonthBreakdown      `json:"months"`
	TotalsByCategory  []CategoryTotalToDate `json:"totals_by_category"`
	TotalSpentToDate  float64               `json:"total_spent_to_date"`
	TotalIncomeToDate float64               `json:"total_income_to_date"`
}

const defaultMiscColor = "#8c8c8c"

/**************************************************************************************
* GetSpendingBreakdown computes the user's spending breakdown over the past 36 months,
* grouped by month and category, plus to-date category totals with % of total income.
* Expenses not matching any user category are classified as "misc".
**************************************************************************************/
func GetSpendingBreakdown(userID uint) (*SpendingBreakdownResponse, error) {
	now := time.Now()

	// 1. Build chronological slice of the past 36 months ending with current month
	type monthSlot struct {
		ym    string
		label string
	}
	monthsList := make([]monthSlot, 36)
	startMonth := time.Date(now.Year(), now.Month()-35, 1, 0, 0, 0, 0, time.Local)
	for i := 0; i < 36; i++ {
		t := startMonth.AddDate(0, i, 0)
		monthsList[i] = monthSlot{
			ym:    t.Format("2006-01"),
			label: t.Format("Jan '06"),
		}
	}
	fromDate := startMonth.Format("2006-01-02")
	toDate := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, time.Local).Format("2006-01-02")

	// 2. Load user's expense categories
	catColors := make(map[string]string)
	definedCats := make([]string, 0)
	catRows, err := ModelsRepo.DB.Conn.Query(
		"SELECT label, color FROM expense_categories WHERE user_id = ? ORDER BY label", userID)
	if err != nil {
		return nil, fmt.Errorf("could not list categories: %w", err)
	}
	defer catRows.Close()

	for catRows.Next() {
		var label, color string
		if err := catRows.Scan(&label, &color); err != nil {
			return nil, fmt.Errorf("could not scan category: %w", err)
		}
		catColors[label] = color
		definedCats = append(definedCats, label)
	}

	// 3. Query paid occurrences in the 36-month window
	monthlyCatSpending := make(map[string]map[string]float64)
	monthlyCatItems := make(map[string]map[string][]ExpenseItem)
	occRows, err := ModelsRepo.DB.Conn.Query(`
		SELECT DATE_FORMAT(eo.due_on, '%Y-%m') AS ym, COALESCE(e.category, 'misc'), e.label, eo.amount, DATE_FORMAT(eo.due_on, '%Y-%m-%d')
		FROM expense_occurrences eo
		JOIN expenses e ON e.id = eo.expense_id
		WHERE eo.user_id = ? AND eo.paid = 1 AND eo.due_on >= ? AND eo.due_on <= ?
		ORDER BY eo.amount DESC`, userID, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("could not query paid occurrences: %w", err)
	}
	defer occRows.Close()

	for occRows.Next() {
		var ym, rawCat, label, date string
		var amount float64
		if err := occRows.Scan(&ym, &rawCat, &label, &amount, &date); err != nil {
			return nil, fmt.Errorf("could not scan occurrence row: %w", err)
		}
		cat := rawCat
		if _, ok := catColors[cat]; !ok {
			cat = "misc"
		}
		if monthlyCatSpending[ym] == nil {
			monthlyCatSpending[ym] = make(map[string]float64)
			monthlyCatItems[ym] = make(map[string][]ExpenseItem)
		}
		monthlyCatSpending[ym][cat] += amount
		monthlyCatItems[ym][cat] = append(monthlyCatItems[ym][cat], ExpenseItem{
			Label:  label,
			Amount: round2(amount),
			Date:   date,
		})
	}

	// 4. Query all-time paid occurrences for totals to date
	allTimeCatTotal := make(map[string]float64)
	allTimeCatItems := make(map[string][]ExpenseItem)
	var totalSpentToDate float64
	allTimeRows, err := ModelsRepo.DB.Conn.Query(`
		SELECT COALESCE(e.category, 'misc'), e.label, eo.amount, DATE_FORMAT(eo.due_on, '%Y-%m-%d')
		FROM expense_occurrences eo
		JOIN expenses e ON e.id = eo.expense_id
		WHERE eo.user_id = ? AND eo.paid = 1
		ORDER BY eo.amount DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("could not query all-time occurrences: %w", err)
	}
	defer allTimeRows.Close()

	for allTimeRows.Next() {
		var rawCat, label, date string
		var amount float64
		if err := allTimeRows.Scan(&rawCat, &label, &amount, &date); err != nil {
			return nil, fmt.Errorf("could not scan all-time occurrence row: %w", err)
		}
		cat := rawCat
		if _, ok := catColors[cat]; !ok {
			cat = "misc"
		}
		allTimeCatTotal[cat] += amount
		allTimeCatItems[cat] = append(allTimeCatItems[cat], ExpenseItem{
			Label:  label,
			Amount: round2(amount),
			Date:   date,
		})
		totalSpentToDate += amount
	}

	// 5. Query income (approved entries with Plaid fallback)
	monthlyIncome := make(map[string]float64)
	incRows, err := ModelsRepo.DB.Conn.Query(`
		SELECT DATE_FORMAT(occurred_on, '%Y-%m') AS ym, SUM(amount)
		FROM income_entries
		WHERE user_id = ? AND status = 'approved' AND occurred_on >= ? AND occurred_on <= ?
		GROUP BY ym`, userID, fromDate, toDate)
	if err == nil {
		for incRows.Next() {
			var ym string
			var amt float64
			if err := incRows.Scan(&ym, &amt); err == nil {
				monthlyIncome[ym] = amt
			}
		}
		incRows.Close()
	}

	plaidIncRows, err := ModelsRepo.DB.Conn.Query(`
		SELECT DATE_FORMAT(date, '%Y-%m') AS ym, SUM(ABS(amount))
		FROM plaid_transactions
		WHERE user_id = ? AND pending = 0 AND (category = 'INCOME' OR amount < 0) AND date >= ? AND date <= ?
		GROUP BY ym`, userID, fromDate, toDate)
	if err == nil {
		for plaidIncRows.Next() {
			var ym string
			var amt float64
			if err := plaidIncRows.Scan(&ym, &amt); err == nil {
				if monthlyIncome[ym] == 0 {
					monthlyIncome[ym] = amt
				}
			}
		}
		plaidIncRows.Close()
	}

	var totalIncome float64
	_ = ModelsRepo.DB.Conn.QueryRow(
		"SELECT COALESCE(SUM(amount), 0) FROM income_entries WHERE user_id = ? AND status = 'approved'",
		userID).Scan(&totalIncome)
	if totalIncome == 0 {
		_ = ModelsRepo.DB.Conn.QueryRow(`
			SELECT COALESCE(SUM(ABS(amount)), 0) FROM plaid_transactions
			WHERE user_id = ? AND pending = 0 AND (category = 'INCOME' OR amount < 0)`,
			userID).Scan(&totalIncome)
	}

	// 6. Build TotalsByCategory (sorted from biggest to smallest)
	totalsByCategory := make([]CategoryTotalToDate, 0)
	for _, cat := range definedCats {
		spent := round2(allTimeCatTotal[cat])
		var pct *float64
		if totalIncome > 0 {
			p := math.Round((spent/totalIncome)*1000) / 10
			pct = &p
		}
		items := allTimeCatItems[cat]
		if items == nil {
			items = []ExpenseItem{}
		}
		totalsByCategory = append(totalsByCategory, CategoryTotalToDate{
			Category:           cat,
			Color:              catColors[cat],
			TotalSpent:         spent,
			PercentageOfIncome: pct,
			Items:              items,
		})
	}
	if miscSpent := allTimeCatTotal["misc"]; miscSpent > 0 || len(definedCats) == 0 {
		spent := round2(miscSpent)
		var pct *float64
		if totalIncome > 0 {
			p := math.Round((spent/totalIncome)*1000) / 10
			pct = &p
		}
		items := allTimeCatItems["misc"]
		if items == nil {
			items = []ExpenseItem{}
		}
		totalsByCategory = append(totalsByCategory, CategoryTotalToDate{
			Category:           "misc",
			Color:              defaultMiscColor,
			TotalSpent:         spent,
			PercentageOfIncome: pct,
			Items:              items,
		})
	}

	// Sort categories from biggest to smallest
	sort.Slice(totalsByCategory, func(a, b int) bool {
		return totalsByCategory[a].TotalSpent > totalsByCategory[b].TotalSpent
	})

	// 7. Build Months
	months := make([]MonthBreakdown, 36)
	for i, slot := range monthsList {
		catMap := monthlyCatSpending[slot.ym]
		itemMap := monthlyCatItems[slot.ym]
		breakdown := make([]CategoryBreakdownItem, 0)
		monthTotal := 0.0

		for cat, amount := range catMap {
			if amount <= 0 {
				continue
			}
			color := catColors[cat]
			if color == "" {
				color = defaultMiscColor
			}
			items := itemMap[cat]
			if items == nil {
				items = []ExpenseItem{}
			}
			breakdown = append(breakdown, CategoryBreakdownItem{
				Category: cat,
				Amount:   round2(amount),
				Color:    color,
				Items:    items,
			})
			monthTotal += amount
		}

		sort.Slice(breakdown, func(a, b int) bool {
			return breakdown[a].Amount > breakdown[b].Amount
		})

		months[i] = MonthBreakdown{
			Month:     slot.ym,
			Label:     slot.label,
			Total:     round2(monthTotal),
			Income:    round2(monthlyIncome[slot.ym]),
			Breakdown: breakdown,
		}
	}

	return &SpendingBreakdownResponse{
		Months:            months,
		TotalsByCategory:  totalsByCategory,
		TotalSpentToDate:  round2(totalSpentToDate),
		TotalIncomeToDate: round2(totalIncome),
	}, nil
}
