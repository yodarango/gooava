package models

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"gooava/internal/lib"
)

/* *************************************************************************************************
* The finance assistant's query catalog. The model never writes SQL — it picks one of these
* predefined, user-scoped queries and optionally narrows the date window. Tables holding
* credentials (users, plaid_items) or app internals are deliberately NOT queryable, and no
* query ever returns secrets.
* ************************************************************************************************
 */

// ChatTools is passed to every chat completion so the model fetches only the data the
// conversation actually needs.
var ChatTools = []lib.ToolDefinition{
	{
		Type: "function",
		Function: lib.ToolFunction{
			Name:        "get_balance",
			Description: "Money on hand right now: the manual bank balance (null when never set), whether a bank is linked, the linked checking total, and every linked account with its current balance.",
			Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		},
	},
	{
		Type: "function",
		Function: lib.ToolFunction{
			Name:        "get_upcoming_bills",
			Description: "Bill occurrences in a date window: label, category, amount, due date, paid flag, past-due flag, plus the unpaid total. Default window: today through today+60 days.",
			Parameters: json.RawMessage(`{
				"type":"object",
				"properties":{
					"from_date":{"type":"string","description":"YYYY-MM-DD, default today"},
					"to_date":{"type":"string","description":"YYYY-MM-DD, default today+60 days"},
					"status":{"type":"string","enum":["unpaid","all"],"description":"default unpaid"}
				}
			}`),
		},
	},
	{
		Type: "function",
		Function: lib.ToolFunction{
			Name:        "get_expense_definitions",
			Description: "The user's bill/expense definitions (label, category, amount, recurrence schedule) and their category list.",
			Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		},
	},
	{
		Type: "function",
		Function: lib.ToolFunction{
			Name:        "get_income_sources",
			Description: "Income sources with label, amount, recurrence schedule, and the next few occurrence dates computed from that schedule — i.e. the user's upcoming paycheck dates and amounts.",
			Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		},
	},
	{
		Type: "function",
		Function: lib.ToolFunction{
			Name:        "get_income_history",
			Description: "Recorded income entries (paychecks approved or rejected) in a date window. Default: the last 90 days.",
			Parameters: json.RawMessage(`{
				"type":"object",
				"properties":{
					"from_date":{"type":"string","description":"YYYY-MM-DD, default today-90 days"},
					"to_date":{"type":"string","description":"YYYY-MM-DD, default today"}
				}
			}`),
		},
	},
	{
		Type: "function",
		Function: lib.ToolFunction{
			Name:        "get_transactions",
			Description: "Synced bank transactions, newest first: date, merchant/name, amount (positive = money spent, negative = money received), category, pending flag. Default: last 30 days, at most 250 rows.",
			Parameters: json.RawMessage(`{
				"type":"object",
				"properties":{
					"from_date":{"type":"string","description":"YYYY-MM-DD, default today-30 days"},
					"to_date":{"type":"string","description":"YYYY-MM-DD, default today"},
					"category":{"type":"string","description":"exact category filter, e.g. FOOD_AND_DRINK"}
				}
			}`),
		},
	},
	{
		Type: "function",
		Function: lib.ToolFunction{
			Name:        "get_spending_summary",
			Description: "Spending totals grouped by category plus the top 10 merchants over a window (pending excluded, money-out only), with the grand total. Default: last 90 days.",
			Parameters: json.RawMessage(`{
				"type":"object",
				"properties":{
					"from_date":{"type":"string","description":"YYYY-MM-DD, default today-90 days"},
					"to_date":{"type":"string","description":"YYYY-MM-DD, default today"}
				}
			}`),
		},
	},
}

/**************************************************************************************
* ExecuteChatTool runs one catalog query, always scoped to the user, and returns compact
* JSON. Errors come back as short text the model can adapt to — tool chatter is never
* persisted.
**************************************************************************************/
func ExecuteChatTool(userID uint, name, argumentsJSON string) (string, error) {
	args := map[string]interface{}{}
	if strings.TrimSpace(argumentsJSON) != "" {
		if err := json.Unmarshal([]byte(argumentsJSON), &args); err != nil {
			return "", fmt.Errorf("invalid arguments: %v", err)
		}
	}

	switch name {
	case "get_balance":
		return chatToolBalance(userID)
	case "get_upcoming_bills":
		from, to, err := chatToolRange(args, 0, 60)
		if err != nil {
			return "", err
		}
		return chatToolBills(userID, from, to, argString(args, "status"))
	case "get_expense_definitions":
		return chatToolExpenseDefinitions(userID)
	case "get_income_sources":
		return chatToolIncomeSources(userID)
	case "get_income_history":
		from, to, err := chatToolRange(args, -90, 0)
		if err != nil {
			return "", err
		}
		return chatToolIncomeHistory(userID, from, to)
	case "get_transactions":
		from, to, err := chatToolRange(args, -30, 0)
		if err != nil {
			return "", err
		}
		return chatToolTransactions(userID, from, to, argString(args, "category"))
	case "get_spending_summary":
		from, to, err := chatToolRange(args, -90, 0)
		if err != nil {
			return "", err
		}
		return chatToolSpendingSummary(userID, from, to)
	default:
		return "", fmt.Errorf("unknown tool %q", name)
	}
}

func argString(args map[string]interface{}, key string) string {
	if v, ok := args[key].(string); ok {
		return strings.TrimSpace(v)
	}
	return ""
}

/**************************************************************************************
* chatToolRange resolves the from_date/to_date args (YYYY-MM-DD) against defaults given
* as day offsets from today, and clamps the span to 366 days.
**************************************************************************************/
func chatToolRange(args map[string]interface{}, fromOffsetDays, toOffsetDays int) (string, string, error) {
	now := time.Now()
	from := now.AddDate(0, 0, fromOffsetDays).Format("2006-01-02")
	to := now.AddDate(0, 0, toOffsetDays).Format("2006-01-02")

	if v := argString(args, "from_date"); v != "" {
		if _, err := time.Parse("2006-01-02", v); err != nil {
			return "", "", fmt.Errorf("from_date must be YYYY-MM-DD")
		}
		from = v
	}
	if v := argString(args, "to_date"); v != "" {
		if _, err := time.Parse("2006-01-02", v); err != nil {
			return "", "", fmt.Errorf("to_date must be YYYY-MM-DD")
		}
		to = v
	}
	if from > to {
		return "", "", fmt.Errorf("from_date must be on or before to_date")
	}

	fromTime, _ := time.Parse("2006-01-02", from)
	toTime, _ := time.Parse("2006-01-02", to)
	if toTime.Sub(fromTime) > 366*24*time.Hour {
		to = fromTime.AddDate(0, 0, 366).Format("2006-01-02")
	}

	return from, to, nil
}

func marshalChatToolResult(v interface{}) (string, error) {
	encoded, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("could not encode tool result: %w", err)
	}
	return string(encoded), nil
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

/**************************************************************************************
* get_balance: manual balance + linked accounts + checking total.
**************************************************************************************/
func chatToolBalance(userID uint) (string, error) {
	balance, err := GetBankBalance(userID)
	if err != nil {
		return "", err
	}
	accounts, err := ListPlaidAccounts(userID)
	if err != nil {
		return "", err
	}

	type account struct {
		Name    string  `json:"name"`
		Mask    string  `json:"mask"`
		Subtype string  `json:"subtype"`
		Balance float64 `json:"balance"`
	}
	result := struct {
		ManualBalance *float64  `json:"manual_balance"`
		HasLinkedBank bool      `json:"has_linked_bank"`
		CheckingTotal float64   `json:"checking_total"`
		Accounts      []account `json:"accounts"`
	}{ManualBalance: balance, Accounts: make([]account, 0)}

	for _, a := range accounts {
		result.HasLinkedBank = true
		if a.Subtype == "checking" {
			result.CheckingTotal += a.CurrentBalance
		}
		result.Accounts = append(result.Accounts, account{
			Name: a.Name, Mask: a.Mask, Subtype: a.Subtype, Balance: a.CurrentBalance,
		})
	}
	result.CheckingTotal = round2(result.CheckingTotal)

	return marshalChatToolResult(result)
}

/**************************************************************************************
* get_upcoming_bills: bill occurrences in the window with the unpaid grand total.
**************************************************************************************/
func chatToolBills(userID uint, from, to, status string) (string, error) {
	query := `SELECT e.label, e.category, eo.amount, DATE_FORMAT(eo.due_on, '%Y-%m-%d'), eo.paid
		FROM expense_occurrences eo
		JOIN expenses e ON e.id = eo.expense_id
		WHERE eo.user_id = ? AND eo.due_on >= ? AND eo.due_on <= ?`
	if status != "all" {
		query += " AND eo.paid = 0"
	}
	query += " ORDER BY eo.due_on, eo.id"

	rows, err := ModelsRepo.DB.Conn.Query(query, userID, from, to)
	if err != nil {
		return "", fmt.Errorf("could not list bills: %w", err)
	}
	defer rows.Close()

	type bill struct {
		Label    string  `json:"label"`
		Category string  `json:"category"`
		Amount   float64 `json:"amount"`
		DueOn    string  `json:"due_on"`
		Paid     bool    `json:"paid"`
		PastDue  bool    `json:"past_due"`
	}

	today := time.Now().Format("2006-01-02")
	bills := make([]bill, 0)
	unpaidTotal := 0.0
	for rows.Next() {
		var b bill
		if err := rows.Scan(&b.Label, &b.Category, &b.Amount, &b.DueOn, &b.Paid); err != nil {
			return "", fmt.Errorf("could not scan bill: %w", err)
		}
		b.PastDue = !b.Paid && b.DueOn < today
		if !b.Paid {
			unpaidTotal += b.Amount
		}
		bills = append(bills, b)
	}

	return marshalChatToolResult(map[string]interface{}{
		"from": from, "to": to,
		"unpaid_total": round2(unpaidTotal),
		"bills":        bills,
	})
}

/**************************************************************************************
* get_expense_definitions: the bill definitions (with recurrence rules) + category list.
**************************************************************************************/
func chatToolExpenseDefinitions(userID uint) (string, error) {
	expenses, err := ListExpenses(userID)
	if err != nil {
		return "", err
	}

	type expenseDef struct {
		Label     string          `json:"label"`
		Category  string          `json:"category"`
		Amount    float64         `json:"amount"`
		Recurring string          `json:"recurring"`
		RecurRule json.RawMessage `json:"recur_rule"`
	}
	defs := make([]expenseDef, 0, len(expenses))
	for _, e := range expenses {
		defs = append(defs, expenseDef{
			Label: e.Label, Category: e.Category, Amount: e.Amount,
			Recurring: e.Recurring, RecurRule: e.RecurRule,
		})
	}

	rows, err := ModelsRepo.DB.Conn.Query(
		`SELECT label, COALESCE(description, '') FROM expense_categories
		 WHERE user_id = ? ORDER BY label`, userID)
	if err != nil {
		return "", fmt.Errorf("could not list expense categories: %w", err)
	}
	defer rows.Close()

	type category struct {
		Label       string `json:"label"`
		Description string `json:"description"`
	}
	categories := make([]category, 0)
	for rows.Next() {
		var c category
		if err := rows.Scan(&c.Label, &c.Description); err != nil {
			return "", fmt.Errorf("could not scan expense category: %w", err)
		}
		categories = append(categories, c)
	}

	return marshalChatToolResult(map[string]interface{}{
		"recur_rule_legend": "weekly {day_of_week 0-6, Sun=0}; biweekly {days_of_month [d1,d2]}; monthly {day_of_month}; quarterly {day, months [1-12]}; yearly {month, day}",
		"expenses":          defs,
		"categories":        categories,
	})
}

/**************************************************************************************
* get_income_sources: income definitions + computed next occurrence (paycheck) dates.
**************************************************************************************/
func chatToolIncomeSources(userID uint) (string, error) {
	incomes, err := ListIncomes(userID)
	if err != nil {
		return "", err
	}

	type incomeSource struct {
		Label           string          `json:"label"`
		Amount          float64         `json:"amount"`
		Recurring       string          `json:"recurring"`
		RecurRule       json.RawMessage `json:"recur_rule"`
		NextOccurrences []string        `json:"next_occurrences"`
	}

	now := time.Now()
	sources := make([]incomeSource, 0, len(incomes))
	for _, i := range incomes {
		next := nextIncomeDates(i, now, 4)
		if next == nil {
			next = []string{}
		}
		sources = append(sources, incomeSource{
			Label: i.Label, Amount: i.Amount, Recurring: i.Recurring,
			RecurRule: i.RecurRule, NextOccurrences: next,
		})
	}

	return marshalChatToolResult(map[string]interface{}{
		"today":             now.Format("2006-01-02"),
		"recur_rule_legend": "one-time {date}; weekly {day_of_week 0-6, Sun=0}; biweekly {mode every_other_week, anchor YYYY-MM-DD} or {mode days_of_month, days_of_month [d1,d2]}; monthly {day_of_month}; quarterly/semiannually {day_of_month, every 3rd/6th month from the first occurrence}; annually {month, day_of_month}",
		"incomes":           sources,
	})
}

/* *************************************************************************************************
* Income recurrence — a Go port of web/src/utils/income.js so paycheck dates are computed
* identically everywhere. recur_rule shapes:
*   "" (one-time): { date: "YYYY-MM-DD" }
*   weekly:        { day_of_week: 0-6 }                    (0 = Sunday)
*   biweekly:      { mode: "every_other_week", anchor: "YYYY-MM-DD" }  (anchor is a real payday)
*               or { mode: "days_of_month", days_of_month: [d1, d2] }
*   monthly:       { day_of_month: d }
*   quarterly:     { day_of_month: d }  (every 3rd month from the first occurrence)
*   semiannually:  { day_of_month: d }  (every 6th month from the first occurrence)
*   annually:      { month: m, day_of_month: d }
* ************************************************************************************************
 */
type incomeRule struct {
	Date        string `json:"date"`
	DayOfWeek   *int   `json:"day_of_week"`
	Mode        string `json:"mode"`
	Anchor      string `json:"anchor"`
	DaysOfMonth []int  `json:"days_of_month"`
	DayOfMonth  *int   `json:"day_of_month"`
	Month       int    `json:"month"`
}

// "YYYY-MM-DD" as a LOCAL date (time stripped) — string parsing avoids timezone surprises.
func parseLocalDate(s string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", s)
	if err != nil {
		return parsed, err
	}
	return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.Local), nil
}

// matches when the date's day-of-month equals `day`, clamped to the month's real
// length (e.g. the 31st becomes Feb 28)
func isDayOfMonth(date time.Time, day int) bool {
	last := time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, time.Local).Day()
	if day > last {
		day = last
	}
	return date.Day() == day
}

// incomeOccursOn mirrors incomeOccursOn in web/src/utils/income.js. anchor pins
// quarterly/semiannual cycles to the first occurrence's month; nil anchor matches any month.
func incomeOccursOn(recurring string, rule incomeRule, date time.Time, anchor *time.Time) bool {
	switch recurring {
	case "weekly":
		return rule.DayOfWeek != nil && int(date.Weekday()) == *rule.DayOfWeek

	case "biweekly":
		if rule.Mode == "days_of_month" {
			for _, day := range rule.DaysOfMonth {
				if isDayOfMonth(date, day) {
					return true
				}
			}
			return false
		}
		if rule.Mode == "every_other_week" && rule.Anchor != "" {
			anchorDate, err := parseLocalDate(rule.Anchor)
			if err != nil {
				return false
			}
			diff := int(math.Round(date.Sub(anchorDate).Hours() / 24))
			return ((diff%14)+14)%14 == 0
		}
		return false

	case "monthly":
		return rule.DayOfMonth != nil && isDayOfMonth(date, *rule.DayOfMonth)

	case "quarterly", "semiannually":
		if rule.DayOfMonth == nil || !isDayOfMonth(date, *rule.DayOfMonth) {
			return false
		}
		if anchor == nil {
			return true
		}
		step := 3
		if recurring == "semiannually" {
			step = 6
		}
		monthDiff := (date.Year()-anchor.Year())*12 + int(date.Month()) - int(anchor.Month())
		return ((monthDiff%step)+step)%step == 0

	case "annually":
		return rule.Month == int(date.Month()) && rule.DayOfMonth != nil && isDayOfMonth(date, *rule.DayOfMonth)
	}

	return false
}

/**************************************************************************************
* nextIncomeDates returns the income's next `count` occurrence dates on/after from —
* the upcoming paycheck dates. One-time incomes occur at most once, never again once past.
**************************************************************************************/
func nextIncomeDates(income Income, from time.Time, count int) []string {
	var rule incomeRule
	if len(income.RecurRule) > 0 {
		_ = json.Unmarshal(income.RecurRule, &rule)
	}
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.Local)

	if income.Recurring == "" {
		if rule.Date == "" {
			return nil
		}
		date, err := parseLocalDate(rule.Date)
		if err != nil || date.Before(from) {
			return nil
		}
		return []string{rule.Date}
	}

	// the first hit (matched with no anchor) anchors quarterly/semiannual cycles
	dates := make([]string, 0, count)
	var anchor *time.Time
	cursor := from
	limit := from.AddDate(2, 0, 0)
	for cursor.Before(limit) && len(dates) < count {
		if incomeOccursOn(income.Recurring, rule, cursor, anchor) {
			if anchor == nil {
				first := cursor
				anchor = &first
			}
			dates = append(dates, cursor.Format("2006-01-02"))
		}
		cursor = cursor.AddDate(0, 0, 1)
	}

	return dates
}

/**************************************************************************************
* get_income_history: recorded paycheck entries in the window, newest first.
**************************************************************************************/
func chatToolIncomeHistory(userID uint, from, to string) (string, error) {
	rows, err := ModelsRepo.DB.Conn.Query(
		`SELECT i.label, DATE_FORMAT(ie.occurred_on, '%Y-%m-%d'), ie.amount, ie.status
		 FROM income_entries ie
		 JOIN incomes i ON i.id = ie.income_id
		 WHERE ie.user_id = ? AND ie.occurred_on >= ? AND ie.occurred_on <= ?
		 ORDER BY ie.occurred_on DESC, ie.id DESC`, userID, from, to)
	if err != nil {
		return "", fmt.Errorf("could not list income entries: %w", err)
	}
	defer rows.Close()

	type entry struct {
		Label      string  `json:"label"`
		OccurredOn string  `json:"occurred_on"`
		Amount     float64 `json:"amount"`
		Status     string  `json:"status"`
	}
	entries := make([]entry, 0)
	for rows.Next() {
		var e entry
		if err := rows.Scan(&e.Label, &e.OccurredOn, &e.Amount, &e.Status); err != nil {
			return "", fmt.Errorf("could not scan income entry: %w", err)
		}
		entries = append(entries, e)
	}

	return marshalChatToolResult(map[string]interface{}{
		"from": from, "to": to, "entries": entries,
	})
}

/**************************************************************************************
* get_transactions: synced bank transactions in the window (newest first, max 250).
**************************************************************************************/
func chatToolTransactions(userID uint, from, to, category string) (string, error) {
	query := `SELECT DATE_FORMAT(t.date, '%Y-%m-%d'), t.name, COALESCE(t.merchant_name, ''),
	                 t.amount, COALESCE(t.category, ''), t.pending
	          FROM plaid_transactions t
	          WHERE t.user_id = ? AND t.date >= ? AND t.date <= ?`
	params := []interface{}{userID, from, to}
	if category != "" {
		query += " AND t.category = ?"
		params = append(params, category)
	}
	query += " ORDER BY t.date DESC, t.id DESC LIMIT 250"

	rows, err := ModelsRepo.DB.Conn.Query(query, params...)
	if err != nil {
		return "", fmt.Errorf("could not list transactions: %w", err)
	}
	defer rows.Close()

	type transaction struct {
		Date     string  `json:"date"`
		Name     string  `json:"name"`
		Merchant string  `json:"merchant"`
		Amount   float64 `json:"amount"` // positive = money spent (Plaid convention)
		Category string  `json:"category"`
		Pending  bool    `json:"pending"`
	}
	transactions := make([]transaction, 0)
	for rows.Next() {
		var t transaction
		if err := rows.Scan(&t.Date, &t.Name, &t.Merchant, &t.Amount, &t.Category, &t.Pending); err != nil {
			return "", fmt.Errorf("could not scan transaction: %w", err)
		}
		transactions = append(transactions, t)
	}

	return marshalChatToolResult(map[string]interface{}{
		"from": from, "to": to, "transactions": transactions,
	})
}

/**************************************************************************************
* get_spending_summary: category totals + top 10 merchants over the window
* (pending excluded, money-out only), with the grand total.
**************************************************************************************/
func chatToolSpendingSummary(userID uint, from, to string) (string, error) {
	categoryRows, err := ModelsRepo.DB.Conn.Query(
		`SELECT COALESCE(NULLIF(t.category, ''), 'UNCATEGORIZED'), SUM(t.amount)
		 FROM plaid_transactions t
		 WHERE t.user_id = ? AND t.pending = 0 AND t.amount > 0 AND t.date >= ? AND t.date <= ?
		 GROUP BY 1
		 ORDER BY 2 DESC`, userID, from, to)
	if err != nil {
		return "", fmt.Errorf("could not summarize spending by category: %w", err)
	}
	defer categoryRows.Close()

	type categoryTotal struct {
		Category string  `json:"category"`
		Total    float64 `json:"total"`
	}
	byCategory := make([]categoryTotal, 0)
	totalSpent := 0.0
	for categoryRows.Next() {
		var c categoryTotal
		if err := categoryRows.Scan(&c.Category, &c.Total); err != nil {
			return "", fmt.Errorf("could not scan category total: %w", err)
		}
		c.Total = round2(c.Total)
		totalSpent += c.Total
		byCategory = append(byCategory, c)
	}

	merchantRows, err := ModelsRepo.DB.Conn.Query(
		`SELECT COALESCE(NULLIF(t.merchant_name, ''), t.name), SUM(t.amount)
		 FROM plaid_transactions t
		 WHERE t.user_id = ? AND t.pending = 0 AND t.amount > 0 AND t.date >= ? AND t.date <= ?
		 GROUP BY 1
		 ORDER BY 2 DESC
		 LIMIT 10`, userID, from, to)
	if err != nil {
		return "", fmt.Errorf("could not summarize spending by merchant: %w", err)
	}
	defer merchantRows.Close()

	type merchantTotal struct {
		Merchant string  `json:"merchant"`
		Total    float64 `json:"total"`
	}
	topMerchants := make([]merchantTotal, 0)
	for merchantRows.Next() {
		var m merchantTotal
		if err := merchantRows.Scan(&m.Merchant, &m.Total); err != nil {
			return "", fmt.Errorf("could not scan merchant total: %w", err)
		}
		m.Total = round2(m.Total)
		topMerchants = append(topMerchants, m)
	}

	return marshalChatToolResult(map[string]interface{}{
		"from": from, "to": to,
		"total_spent":   round2(totalSpent),
		"by_category":   byCategory,
		"top_merchants": topMerchants,
	})
}
