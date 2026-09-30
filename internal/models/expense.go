package models

import (
	"encoding/json"
	"fmt"
)

type Expense struct {
	Id        int64           `json:"id"`
	UserID    uint            `json:"user_id"`
	Category  string          `json:"category"`
	Amount    float64         `json:"amount"`
	Label     string          `json:"label"`
	Recurring string          `json:"recurring"`  // "" (one-time), daily, weekly, biweekly, monthly, quarterly, yearly
	RecurRule json.RawMessage `json:"recur_rule"` // JSON blob with the recurrence details, NULL when not recurring/daily
	CreatedAt string          `json:"created_at"`
}

/**************************************************************************************
* Creates a manually entered expense for the user. The category defaults to "misc"
* until a categories section exists. The label is stored as a plain string so it can
* later be matched to Plaid transaction names by string equivalence.
*
* status: ✅
**************************************************************************************/
func (e *Expense) Create() error {
	if e.Category == "" {
		e.Category = "misc"
	}

	query := `
		INSERT INTO expenses(user_id, category, amount, label, recurring, recur_rule)
		VALUES(?,?,?,?,?,?)
	`

	// store NULLs instead of empty values for one-time expenses
	var recurring interface{}
	if e.Recurring != "" {
		recurring = e.Recurring
	}
	var recurRule interface{}
	if len(e.RecurRule) > 0 {
		recurRule = []byte(e.RecurRule)
	}

	result, err := ModelsRepo.DB.Conn.Exec(query, e.UserID, e.Category, e.Amount, e.Label, recurring, recurRule)
	if err != nil {
		return fmt.Errorf("could not create expense: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("could not get last expense id: %w", err)
	}

	e.Id = id

	return nil
}

/**************************************************************************************
* ListExpenses returns the user's expenses, newest first. recur_rule comes back as
* raw JSON (nil when NULL).
**************************************************************************************/
func ListExpenses(userID uint) ([]Expense, error) {
	query := `
		SELECT id, category, amount, label, COALESCE(recurring, ''), recur_rule
		FROM expenses
		WHERE user_id = ?
		ORDER BY id DESC
	`

	rows, err := ModelsRepo.DB.Conn.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("could not list expenses: %w", err)
	}
	defer rows.Close()

	expenses := make([]Expense, 0)
	for rows.Next() {
		var expense Expense
		var recurRule []byte
		err := rows.Scan(&expense.Id, &expense.Category, &expense.Amount, &expense.Label, &expense.Recurring, &recurRule)
		if err != nil {
			return nil, fmt.Errorf("could not scan expense: %w", err)
		}
		expense.RecurRule = json.RawMessage(recurRule)
		expenses = append(expenses, expense)
	}

	return expenses, nil
}

/**************************************************************************************
* Update overwrites every editable field of one of the user's expenses.
**************************************************************************************/
func (e *Expense) Update() error {
	if e.Category == "" {
		e.Category = "misc"
	}

	var recurring interface{}
	if e.Recurring != "" {
		recurring = e.Recurring
	}
	var recurRule interface{}
	if len(e.RecurRule) > 0 {
		recurRule = []byte(e.RecurRule)
	}

	query := `
		UPDATE expenses
		SET category = ?, amount = ?, label = ?, recurring = ?, recur_rule = ?
		WHERE id = ? AND user_id = ?
	`

	result, err := ModelsRepo.DB.Conn.Exec(query, e.Category, e.Amount, e.Label, recurring, recurRule, e.Id, e.UserID)
	if err != nil {
		return fmt.Errorf("could not update expense: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not verify expense update: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("expense not found")
	}

	return nil
}

/**************************************************************************************
* DeleteExpense removes one of the user's expenses by id.
**************************************************************************************/
func DeleteExpense(userID uint, id int64) error {
	result, err := ModelsRepo.DB.Conn.Exec("DELETE FROM expenses WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		return fmt.Errorf("could not delete expense: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not verify expense deletion: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("expense not found")
	}

	return nil
}
