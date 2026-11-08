package models

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"
)

// FinanceSnapshot is the compact, JSON-serialized picture of the user's money that
// the chat system prompt grounds every answer in (~a few KB).
type FinanceSnapshot struct {
	Today              string                  `json:"today"`
	ManualBalance      *float64                `json:"manual_balance"` // null when never set
	HasLinkedBank      bool                    `json:"has_linked_bank"`
	CheckingTotal      float64                 `json:"checking_total"` // sum of linked checking accounts
	Accounts           []snapshotAccount       `json:"accounts"`
	Incomes            []snapshotIncome        `json:"incomes"`
	Expenses           []snapshotExpense       `json:"expenses"`
	UpcomingBills      []snapshotBill          `json:"upcoming_bills"` // unpaid, due today-30d → today+60d
	SpendByCategory90d []snapshotCategoryTotal `json:"spend_by_category_90d"`
	TopMerchants90d    []snapshotMerchantTotal `json:"top_merchants_90d"`
}

type snapshotAccount struct {
	Name    string  `json:"name"`
	Mask    string  `json:"mask"`
	Subtype string  `json:"subtype"`
	Balance float64 `json:"balance"`
}

type snapshotIncome struct {
	Label     string  `json:"label"`
	Amount    float64 `json:"amount"`
	Recurring string  `json:"recurring"`
}

type snapshotExpense struct {
	Label     string  `json:"label"`
	Category  string  `json:"category"`
	Amount    float64 `json:"amount"`
	Recurring string  `json:"recurring"`
}

type snapshotBill struct {
	Label   string  `json:"label"`
	Amount  float64 `json:"amount"`
	DueOn   string  `json:"due_on"`
	PastDue bool    `json:"past_due"`
}

type snapshotCategoryTotal struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
}

type snapshotMerchantTotal struct {
	Merchant string  `json:"merchant"`
	Total    float64 `json:"total"`
}

/**************************************************************************************
* BuildFinanceSnapshot gathers the user's real data (balance, bills, income, recent
* spending) into a compact JSON string for the chat system prompt.
**************************************************************************************/
func BuildFinanceSnapshot(userID uint) (string, error) {
	snapshot := FinanceSnapshot{
		Today:              time.Now().Format("2006-01-02"),
		Accounts:           make([]snapshotAccount, 0),
		Incomes:            make([]snapshotIncome, 0),
		Expenses:           make([]snapshotExpense, 0),
		UpcomingBills:      make([]snapshotBill, 0),
		SpendByCategory90d: make([]snapshotCategoryTotal, 0),
		TopMerchants90d:    make([]snapshotMerchantTotal, 0),
	}

	// manual "in the bank" balance (null when never saved)
	balance, err := GetBankBalance(userID)
	if err != nil {
		return "", err
	}
	snapshot.ManualBalance = balance

	// linked bank accounts + checking total
	accounts, err := ListPlaidAccounts(userID)
	if err != nil {
		return "", err
	}
	for _, a := range accounts {
		snapshot.HasLinkedBank = true
		if a.Subtype == "checking" {
			snapshot.CheckingTotal += a.CurrentBalance
		}
		snapshot.Accounts = append(snapshot.Accounts, snapshotAccount{
			Name: a.Name, Mask: a.Mask, Subtype: a.Subtype, Balance: a.CurrentBalance,
		})
	}
	snapshot.CheckingTotal = math.Round(snapshot.CheckingTotal*100) / 100

	// income sources
	incomes, err := ListIncomes(userID)
	if err != nil {
		return "", err
	}
	for _, i := range incomes {
		snapshot.Incomes = append(snapshot.Incomes, snapshotIncome{
			Label: i.Label, Amount: i.Amount, Recurring: i.Recurring,
		})
	}

	// expense (bill) definitions
	expenses, err := ListExpenses(userID)
	if err != nil {
		return "", err
	}
	for _, e := range expenses {
		snapshot.Expenses = append(snapshot.Expenses, snapshotExpense{
			Label: e.Label, Category: e.Category, Amount: e.Amount, Recurring: e.Recurring,
		})
	}

	// unpaid bills from 30 days back through 60 days out (past-due flagged)
	bills, err := listUnpaidBillsForSnapshot(userID)
	if err != nil {
		return "", err
	}
	snapshot.UpcomingBills = bills

	// 90-day spending rollups from synced transactions
	transactions, err := ListPlaidTransactions(userID, 500)
	if err != nil {
		return "", err
	}
	cutoff := time.Now().AddDate(0, 0, -90).Format("2006-01-02")
	categoryTotals := map[string]float64{}
	merchantTotals := map[string]float64{}
	for _, t := range transactions {
		if t.Pending || t.Date < cutoff || t.Amount <= 0 {
			continue // positive amount = money out (Plaid convention)
		}
		category := t.Category
		if category == "" {
			category = "UNCATEGORIZED"
		}
		categoryTotals[category] += t.Amount

		merchant := t.MerchantName
		if merchant == "" {
			merchant = t.Name
		}
		merchantTotals[merchant] += t.Amount
	}
	for category, total := range categoryTotals {
		snapshot.SpendByCategory90d = append(snapshot.SpendByCategory90d,
			snapshotCategoryTotal{Category: category, Total: math.Round(total*100) / 100})
	}
	sort.Slice(snapshot.SpendByCategory90d, func(i, j int) bool {
		return snapshot.SpendByCategory90d[i].Total > snapshot.SpendByCategory90d[j].Total
	})
	for merchant, total := range merchantTotals {
		snapshot.TopMerchants90d = append(snapshot.TopMerchants90d,
			snapshotMerchantTotal{Merchant: merchant, Total: math.Round(total*100) / 100})
	}
	sort.Slice(snapshot.TopMerchants90d, func(i, j int) bool {
		return snapshot.TopMerchants90d[i].Total > snapshot.TopMerchants90d[j].Total
	})
	if len(snapshot.TopMerchants90d) > 10 {
		snapshot.TopMerchants90d = snapshot.TopMerchants90d[:10]
	}

	encoded, err := json.Marshal(snapshot)
	if err != nil {
		return "", fmt.Errorf("could not encode finance snapshot: %w", err)
	}

	return string(encoded), nil
}

/**************************************************************************************
* listUnpaidBillsForSnapshot loads unpaid occurrences due between today-30d and
* today+60d, each with its expense label; past-due rows are flagged.
**************************************************************************************/
func listUnpaidBillsForSnapshot(userID uint) ([]snapshotBill, error) {
	now := time.Now()
	from := now.AddDate(0, 0, -30).Format("2006-01-02")
	to := now.AddDate(0, 0, 60).Format("2006-01-02")
	today := now.Format("2006-01-02")

	rows, err := ModelsRepo.DB.Conn.Query(
		`SELECT e.label, eo.amount, DATE_FORMAT(eo.due_on, '%Y-%m-%d')
		 FROM expense_occurrences eo
		 JOIN expenses e ON e.id = eo.expense_id
		 WHERE eo.user_id = ? AND eo.paid = 0 AND eo.due_on >= ? AND eo.due_on <= ?
		 ORDER BY eo.due_on, eo.id`, userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("could not list unpaid bills for snapshot: %w", err)
	}
	defer rows.Close()

	bills := make([]snapshotBill, 0)
	for rows.Next() {
		var bill snapshotBill
		if err := rows.Scan(&bill.Label, &bill.Amount, &bill.DueOn); err != nil {
			return nil, fmt.Errorf("could not scan unpaid bill: %w", err)
		}
		bill.PastDue = bill.DueOn < today
		bills = append(bills, bill)
	}

	return bills, nil
}
