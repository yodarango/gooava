package models

import (
	"fmt"
)

// ExpenseOccurrence is one concrete bill (a row in expense_occurrences) for an expense.
// The `expenses` row is the "type"/schedule; this is a single payable instance of it.
type ExpenseOccurrence struct {
	Id                int64   `json:"id"`
	UserID            uint    `json:"user_id"`
	ExpenseID         int64   `json:"expense_id"`
	DueOn             string  `json:"due_on"` // YYYY-MM-DD
	Amount            float64 `json:"amount"`
	Paid              bool    `json:"paid"`
	PaidSource        string  `json:"paid_source"` // "", "auto", "manual"
	PaidTransactionID string  `json:"paid_transaction_id"`
	PaidAt            string  `json:"paid_at"`
}

/**************************************************************************************
* EnsureExpenseOccurrences inserts any missing occurrence rows for the given
* (expense_id, due_on) pairs, then returns every occurrence for the user in the range.
* Existing rows are left untouched (a paid row is never overwritten).
**************************************************************************************/
func EnsureExpenseOccurrences(userID uint, items []struct {
	ExpenseID int64
	DueOn     string
	Amount    float64
}) error {
	for _, item := range items {
		_, err := ModelsRepo.DB.Conn.Exec(
			`INSERT IGNORE INTO expense_occurrences(user_id, expense_id, due_on, amount)
			 SELECT ?, e.id, ?, e.amount
			 FROM expenses e
			 WHERE e.id = ? AND e.user_id = ?`,
			userID, item.DueOn, item.ExpenseID, userID,
		)
		if err != nil {
			return fmt.Errorf("could not ensure expense occurrence (%v on %v): %w", item.ExpenseID, item.DueOn, err)
		}
	}
	return nil
}

/**************************************************************************************
* ListExpenseOccurrences returns every occurrence for the user with due_on in
* [fromDate, toDate] (inclusive), joined with the expense label for display.
* Ordered by due date.
**************************************************************************************/
func ListExpenseOccurrences(userID uint, fromDate, toDate string) ([]ExpenseOccurrence, error) {
	query := `
		SELECT id, expense_id, DATE_FORMAT(due_on, '%Y-%m-%d'), amount, paid,
		       COALESCE(paid_source, ''), COALESCE(paid_transaction_id, ''),
		       COALESCE(DATE_FORMAT(paid_at, '%Y-%m-%d %H:%i:%s'), '')
		FROM expense_occurrences
		WHERE user_id = ? AND due_on >= ? AND due_on <= ?
		ORDER BY due_on, id
	`

	rows, err := ModelsRepo.DB.Conn.Query(query, userID, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("could not list expense occurrences: %w", err)
	}
	defer rows.Close()

	occurrences := make([]ExpenseOccurrence, 0)
	for rows.Next() {
		var o ExpenseOccurrence
		if err := rows.Scan(&o.Id, &o.ExpenseID, &o.DueOn, &o.Amount, &o.Paid, &o.PaidSource, &o.PaidTransactionID, &o.PaidAt); err != nil {
			return nil, fmt.Errorf("could not scan expense occurrence: %w", err)
		}
		occurrences = append(occurrences, o)
	}

	return occurrences, nil
}

/**************************************************************************************
* AutoPayOccurrencesByLabel marks every unpaid occurrence whose expense label
* case-insensitively equals a synced Plaid transaction's merchant_name or name. Only
* transactions on/after the occurrence's due date count (you can't pay a bill before
* it's due). Returns the ids of the occurrences that were auto-paid.
**************************************************************************************/
func AutoPayOccurrencesByLabel(userID uint) ([]int64, error) {
	query := `
		UPDATE expense_occurrences eo
		JOIN expenses e ON e.id = eo.expense_id
		JOIN plaid_transactions t
		  ON t.user_id = eo.user_id
		 AND t.date >= eo.due_on
		 AND (LOWER(TRIM(t.merchant_name)) = LOWER(TRIM(e.label))
		      OR LOWER(TRIM(t.name)) = LOWER(TRIM(e.label)))
		SET eo.paid = 1, eo.paid_source = 'auto', eo.paid_transaction_id = t.transaction_id, eo.paid_at = NOW()
		WHERE eo.user_id = ? AND eo.paid = 0
	`

	result, err := ModelsRepo.DB.Conn.Exec(query, userID)
	if err != nil {
		return nil, fmt.Errorf("could not auto-pay expense occurrences: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("could not verify auto-pay: %w", err)
	}

	if affected == 0 {
		return []int64{}, nil
	}

	// collect the ids of the rows we just auto-paid (paid_source = 'auto' and not yet reported)
	rows, err := ModelsRepo.DB.Conn.Query(
		`SELECT id FROM expense_occurrences WHERE user_id = ? AND paid = 1 AND paid_source = 'auto'`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("could not load auto-paid occurrences: %w", err)
	}
	defer rows.Close()

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("could not scan auto-paid occurrence: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

/**************************************************************************************
* SetExpenseOccurrencePaid marks one of the user's occurrences paid (source 'manual')
* or unpaid, clearing the paid metadata when un-marking. Returns whether a row changed.
**************************************************************************************/
func SetExpenseOccurrencePaid(userID uint, id int64, paid bool) (bool, error) {
	var query string
	if paid {
		query = `
			UPDATE expense_occurrences
			SET paid = 1, paid_source = 'manual', paid_transaction_id = NULL, paid_at = NOW()
			WHERE id = ? AND user_id = ?
		`
	} else {
		query = `
			UPDATE expense_occurrences
			SET paid = 0, paid_source = NULL, paid_transaction_id = NULL, paid_at = NULL
			WHERE id = ? AND user_id = ?
		`
	}

	result, err := ModelsRepo.DB.Conn.Exec(query, id, userID)
	if err != nil {
		return false, fmt.Errorf("could not update expense occurrence: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("could not verify occurrence update: %w", err)
	}

	return rowsAffected > 0, nil
}
