package models

import (
	"encoding/json"
	"fmt"
)

type Income struct {
	Id        int64           `json:"id"`
	UserID    uint            `json:"user_id"`
	Label     string          `json:"label"`
	Amount    float64         `json:"amount"`
	Recurring string          `json:"recurring"`  // "" (one-time), weekly, biweekly, monthly, quarterly, semiannually, annually
	RecurRule json.RawMessage `json:"recur_rule"` // recurrence details; {"date":"YYYY-MM-DD"} for one-time incomes
	CreatedAt string          `json:"created_at"`
}

// IncomeEntry is the approval-ledger record for one income occurrence.
type IncomeEntry struct {
	Id         int64   `json:"id"`
	UserID     uint    `json:"user_id"`
	IncomeID   int64   `json:"income_id"`
	OccurredOn string  `json:"occurred_on"` // YYYY-MM-DD
	Amount     float64 `json:"amount"`
	Status     string  `json:"status"` // approved | rejected
	CreatedAt  string  `json:"created_at"`
}

/**************************************************************************************
* Creates an income source for the user.
*
* status: ✅
**************************************************************************************/
func (i *Income) Create() error {
	query := `
		INSERT INTO incomes(user_id, label, amount, recurring, recur_rule)
		VALUES(?,?,?,?,?)
	`

	var recurring interface{}
	if i.Recurring != "" {
		recurring = i.Recurring
	}
	var recurRule interface{}
	if len(i.RecurRule) > 0 {
		recurRule = []byte(i.RecurRule)
	}

	result, err := ModelsRepo.DB.Conn.Exec(query, i.UserID, i.Label, i.Amount, recurring, recurRule)
	if err != nil {
		return fmt.Errorf("could not create income: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("could not get last income id: %w", err)
	}

	i.Id = id
	return nil
}

/**************************************************************************************
* ListIncomes returns the user's income sources, newest first.
**************************************************************************************/
func ListIncomes(userID uint) ([]Income, error) {
	query := `
		SELECT id, label, amount, COALESCE(recurring, ''), recur_rule
		FROM incomes
		WHERE user_id = ?
		ORDER BY id DESC
	`

	rows, err := ModelsRepo.DB.Conn.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("could not list incomes: %w", err)
	}
	defer rows.Close()

	incomes := make([]Income, 0)
	for rows.Next() {
		var income Income
		var recurRule []byte
		if err := rows.Scan(&income.Id, &income.Label, &income.Amount, &income.Recurring, &recurRule); err != nil {
			return nil, fmt.Errorf("could not scan income: %w", err)
		}
		income.RecurRule = json.RawMessage(recurRule)
		incomes = append(incomes, income)
	}

	return incomes, nil
}

/**************************************************************************************
* Update overwrites every editable field of one of the user's incomes.
**************************************************************************************/
func (i *Income) Update() error {
	var recurring interface{}
	if i.Recurring != "" {
		recurring = i.Recurring
	}
	var recurRule interface{}
	if len(i.RecurRule) > 0 {
		recurRule = []byte(i.RecurRule)
	}

	query := `
		UPDATE incomes
		SET label = ?, amount = ?, recurring = ?, recur_rule = ?
		WHERE id = ? AND user_id = ?
	`

	result, err := ModelsRepo.DB.Conn.Exec(query, i.Label, i.Amount, recurring, recurRule, i.Id, i.UserID)
	if err != nil {
		return fmt.Errorf("could not update income: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not verify income update: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("income not found")
	}

	return nil
}

/**************************************************************************************
* DeleteIncome removes one of the user's incomes by id (its entries cascade).
**************************************************************************************/
func DeleteIncome(userID uint, id int64) error {
	result, err := ModelsRepo.DB.Conn.Exec("DELETE FROM incomes WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		return fmt.Errorf("could not delete income: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not verify income deletion: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("income not found")
	}

	return nil
}

/**************************************************************************************
* ListIncomeEntries returns the user's approval-ledger entries, newest first.
**************************************************************************************/
func ListIncomeEntries(userID uint) ([]IncomeEntry, error) {
	query := `
		SELECT id, income_id, DATE_FORMAT(occurred_on, '%Y-%m-%d'), amount, status
		FROM income_entries
		WHERE user_id = ?
		ORDER BY occurred_on DESC, id DESC
	`

	rows, err := ModelsRepo.DB.Conn.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("could not list income entries: %w", err)
	}
	defer rows.Close()

	entries := make([]IncomeEntry, 0)
	for rows.Next() {
		var entry IncomeEntry
		if err := rows.Scan(&entry.Id, &entry.IncomeID, &entry.OccurredOn, &entry.Amount, &entry.Status); err != nil {
			return nil, fmt.Errorf("could not scan income entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

/**************************************************************************************
* ApproveIncomeEntry records an approved occurrence. Any rejected rows for the same
* income+date are removed first so re-approving a rejected occurrence works.
**************************************************************************************/
func ApproveIncomeEntry(userID uint, incomeID int64, occurredOn string, amount float64) error {
	if _, err := ModelsRepo.DB.Conn.Exec(
		"DELETE FROM income_entries WHERE income_id = ? AND occurred_on = ? AND status = 'rejected'",
		incomeID, occurredOn,
	); err != nil {
		return fmt.Errorf("could not clear rejected income entry: %w", err)
	}

	_, err := ModelsRepo.DB.Conn.Exec(
		"INSERT INTO income_entries(user_id, income_id, occurred_on, amount, status) VALUES(?,?,?,?,'approved')",
		userID, incomeID, occurredOn, amount,
	)
	if err != nil {
		return fmt.Errorf("could not approve income entry: %w", err)
	}

	return nil
}

/**************************************************************************************
* RejectIncomeEntry records a rejected occurrence so it stops showing up as pending.
**************************************************************************************/
func RejectIncomeEntry(userID uint, incomeID int64, occurredOn string, amount float64) error {
	_, err := ModelsRepo.DB.Conn.Exec(
		"INSERT INTO income_entries(user_id, income_id, occurred_on, amount, status) VALUES(?,?,?,?,'rejected')",
		userID, incomeID, occurredOn, amount,
	)
	if err != nil {
		return fmt.Errorf("could not reject income entry: %w", err)
	}

	return nil
}
