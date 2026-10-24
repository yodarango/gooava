package models

import (
	"database/sql"
	"fmt"
	"strings"
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

// OccurrenceTransactionMatch is the Plaid transaction proposed (or recorded) as the
// payment for an expense occurrence.
type OccurrenceTransactionMatch struct {
	TransactionID string  `json:"transaction_id"`
	Name          string  `json:"name"`
	MerchantName  string  `json:"merchant_name"`
	Amount        float64 `json:"amount"`
	Date          string  `json:"date"`
}

// PastDueOccurrence is an unpaid past-due occurrence joined with its expense's labels and
// the transaction it string-matches (nil when nothing matched — those need manual handling).
type PastDueOccurrence struct {
	ExpenseOccurrence
	ExpenseLabel string                      `json:"expense_label"`
	LinkedLabel  string                      `json:"linked_label"`
	Match        *OccurrenceTransactionMatch `json:"match"`
}

// OccurrencePayment confirms one occurrence as paid by a specific transaction.
type OccurrencePayment struct {
	Id            int64  `json:"id"`
	TransactionID string `json:"transaction_id"`
}

// PaidExpenseOccurrence is a paid occurrence joined with its expense's label — one row
// of the billing history.
type PaidExpenseOccurrence struct {
	ExpenseOccurrence
	ExpenseLabel string `json:"expense_label"`
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
* AutoPayLinkedOccurrences marks every unpaid occurrence of an expense the user explicitly
* linked to a transaction (expenses.linked_label) as paid when a synced transaction matches
* that label — the link is the user's standing confirmation, so no review is needed. Each
* occurrence gets the earliest still-unused transaction on/after its due date. Returns the
* ids of the occurrences that were auto-paid.
**************************************************************************************/
func AutoPayLinkedOccurrences(userID uint) ([]int64, error) {
	occurrences, err := listUnpaidOccurrenceRows(userID, "", "", true)
	if err != nil {
		return nil, err
	}
	if len(occurrences) == 0 {
		return []int64{}, nil
	}

	transactions, err := listMatchCandidates(userID)
	if err != nil {
		return nil, err
	}
	used, err := listUsedTransactionIDs(userID)
	if err != nil {
		return nil, err
	}

	matches := matchTransactionsToOccurrences(occurrences, transactions, used, true)

	ids := make([]int64, 0, len(matches))
	for occurrenceID, txn := range matches {
		result, err := ModelsRepo.DB.Conn.Exec(
			`UPDATE expense_occurrences
			 SET paid = 1, paid_source = 'auto', paid_transaction_id = ?, paid_at = NOW()
			 WHERE id = ? AND user_id = ? AND paid = 0`,
			txn.TransactionID, occurrenceID, userID,
		)
		if err != nil {
			return nil, fmt.Errorf("could not auto-pay expense occurrence %v: %w", occurrenceID, err)
		}
		if changed, err := result.RowsAffected(); err == nil && changed > 0 {
			ids = append(ids, occurrenceID)
		}
	}

	return ids, nil
}

/**************************************************************************************
* ListPastDueOccurrences returns the user's unpaid occurrences with due_on in
* [fromDate, toDate] (the past-due review range), each joined with its expense's labels
* and the transaction it string-matches (nil when none). A transaction already recorded
* as paying another occurrence is never proposed twice.
**************************************************************************************/
func ListPastDueOccurrences(userID uint, fromDate, toDate string) ([]PastDueOccurrence, error) {
	rows, err := listUnpaidOccurrenceRows(userID, fromDate, toDate, false)
	if err != nil {
		return nil, err
	}

	transactions, err := listMatchCandidates(userID)
	if err != nil {
		return nil, err
	}
	used, err := listUsedTransactionIDs(userID)
	if err != nil {
		return nil, err
	}

	matches := matchTransactionsToOccurrences(rows, transactions, used, false)

	occurrences := make([]PastDueOccurrence, 0, len(rows))
	for _, row := range rows {
		occurrence := PastDueOccurrence{
			ExpenseOccurrence: ExpenseOccurrence{
				Id:        row.Id,
				ExpenseID: row.ExpenseID,
				DueOn:     row.DueOn,
				Amount:    row.Amount,
			},
			ExpenseLabel: row.ExpenseLabel,
			LinkedLabel:  row.LinkedLabel,
		}
		if txn, ok := matches[row.Id]; ok {
			matched := *txn
			occurrence.Match = &matched
		}
		occurrences = append(occurrences, occurrence)
	}

	return occurrences, nil
}

/**************************************************************************************
* SetExpenseOccurrencesPaidWithTransactions marks the user's occurrences paid (source
* 'auto') with the transaction that paid each one — the batch behind the review card's
* "mark them paid" confirmation. Each transaction must be one of the user's own synced
* transactions. Returns how many occurrences were marked.
**************************************************************************************/
func SetExpenseOccurrencesPaidWithTransactions(userID uint, payments []OccurrencePayment) (int, error) {
	marked := 0

	for _, payment := range payments {
		var exists int
		err := ModelsRepo.DB.Conn.QueryRow(
			"SELECT 1 FROM plaid_transactions WHERE user_id = ? AND transaction_id = ?",
			userID, payment.TransactionID,
		).Scan(&exists)
		if err != nil {
			return marked, fmt.Errorf("transaction %s is not one of the user's synced transactions", payment.TransactionID)
		}

		result, err := ModelsRepo.DB.Conn.Exec(
			`UPDATE expense_occurrences
			 SET paid = 1, paid_source = 'auto', paid_transaction_id = ?, paid_at = NOW()
			 WHERE id = ? AND user_id = ?`,
			payment.TransactionID, payment.Id, userID,
		)
		if err != nil {
			return marked, fmt.Errorf("could not mark occurrence %v paid: %w", payment.Id, err)
		}
		if changed, err := result.RowsAffected(); err == nil {
			marked += int(changed)
		}
	}

	return marked, nil
}

/**************************************************************************************
* ListPaidExpenseOccurrences returns the user's paid occurrences with due_on in
* [fromDate, toDate), newest due date first, joined with the expense label — one page
* (one month) of the billing history.
**************************************************************************************/
func ListPaidExpenseOccurrences(userID uint, fromDate, toDate string) ([]PaidExpenseOccurrence, error) {
	rows, err := ModelsRepo.DB.Conn.Query(
		`SELECT eo.id, eo.expense_id, DATE_FORMAT(eo.due_on, '%Y-%m-%d'), eo.amount,
		        COALESCE(eo.paid_source, ''), COALESCE(eo.paid_transaction_id, ''),
		        COALESCE(DATE_FORMAT(eo.paid_at, '%Y-%m-%d %H:%i:%s'), ''),
		        e.label
		 FROM expense_occurrences eo
		 JOIN expenses e ON e.id = eo.expense_id
		 WHERE eo.user_id = ? AND eo.paid = 1 AND eo.due_on >= ? AND eo.due_on < ?
		 ORDER BY eo.due_on DESC, eo.id DESC`,
		userID, fromDate, toDate,
	)
	if err != nil {
		return nil, fmt.Errorf("could not list paid expense occurrences: %w", err)
	}
	defer rows.Close()

	occurrences := make([]PaidExpenseOccurrence, 0)
	for rows.Next() {
		var o PaidExpenseOccurrence
		if err := rows.Scan(
			&o.Id, &o.ExpenseID, &o.DueOn, &o.Amount,
			&o.PaidSource, &o.PaidTransactionID, &o.PaidAt,
			&o.ExpenseLabel,
		); err != nil {
			return nil, fmt.Errorf("could not scan paid expense occurrence: %w", err)
		}
		occurrences = append(occurrences, o)
	}

	return occurrences, nil
}

/**************************************************************************************
* HasPaidExpenseOccurrencesBefore reports whether the user has any paid occurrence with
* a due date before the given one — i.e. whether the billing history has earlier pages.
**************************************************************************************/
func HasPaidExpenseOccurrencesBefore(userID uint, beforeDate string) (bool, error) {
	var exists int
	err := ModelsRepo.DB.Conn.QueryRow(
		`SELECT 1 FROM expense_occurrences
		 WHERE user_id = ? AND paid = 1 AND due_on < ? LIMIT 1`,
		userID, beforeDate,
	).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("could not check earlier paid occurrences: %w", err)
	}

	return true, nil
}

// unpaidOccurrenceRow is an unpaid occurrence joined with its expense's matching labels.
type unpaidOccurrenceRow struct {
	Id           int64
	ExpenseID    int64
	DueOn        string
	Amount       float64
	ExpenseLabel string
	LinkedLabel  string
}

/**************************************************************************************
* listUnpaidOccurrenceRows loads the user's unpaid occurrences (oldest due date first)
* joined with their expense's label and linked label. fromDate/toDate bound the due date
* ("" = unbounded); linkedOnly keeps only occurrences of explicitly linked expenses.
**************************************************************************************/
func listUnpaidOccurrenceRows(userID uint, fromDate, toDate string, linkedOnly bool) ([]unpaidOccurrenceRow, error) {
	query := `
		SELECT eo.id, eo.expense_id, DATE_FORMAT(eo.due_on, '%Y-%m-%d'), eo.amount,
		       e.label, COALESCE(e.linked_label, '')
		FROM expense_occurrences eo
		JOIN expenses e ON e.id = eo.expense_id
		WHERE eo.user_id = ? AND eo.paid = 0
	`
	args := []interface{}{userID}
	if linkedOnly {
		query += " AND e.linked_label IS NOT NULL AND e.linked_label <> ''"
	}
	if fromDate != "" {
		query += " AND eo.due_on >= ?"
		args = append(args, fromDate)
	}
	if toDate != "" {
		query += " AND eo.due_on <= ?"
		args = append(args, toDate)
	}
	query += " ORDER BY eo.due_on, eo.id"

	rows, err := ModelsRepo.DB.Conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("could not list unpaid expense occurrences: %w", err)
	}
	defer rows.Close()

	occurrences := make([]unpaidOccurrenceRow, 0)
	for rows.Next() {
		var row unpaidOccurrenceRow
		if err := rows.Scan(&row.Id, &row.ExpenseID, &row.DueOn, &row.Amount, &row.ExpenseLabel, &row.LinkedLabel); err != nil {
			return nil, fmt.Errorf("could not scan unpaid expense occurrence: %w", err)
		}
		occurrences = append(occurrences, row)
	}

	return occurrences, nil
}

/**************************************************************************************
* listMatchCandidates loads the user's settled transactions, oldest first, for matching.
* Pending transactions are excluded — they can still change or disappear.
**************************************************************************************/
func listMatchCandidates(userID uint) ([]OccurrenceTransactionMatch, error) {
	rows, err := ModelsRepo.DB.Conn.Query(
		`SELECT transaction_id, name, COALESCE(merchant_name, ''), amount, DATE_FORMAT(date, '%Y-%m-%d')
		 FROM plaid_transactions
		 WHERE user_id = ? AND pending = 0
		 ORDER BY date, id`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("could not list match candidate transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]OccurrenceTransactionMatch, 0)
	for rows.Next() {
		var txn OccurrenceTransactionMatch
		if err := rows.Scan(&txn.TransactionID, &txn.Name, &txn.MerchantName, &txn.Amount, &txn.Date); err != nil {
			return nil, fmt.Errorf("could not scan match candidate transaction: %w", err)
		}
		transactions = append(transactions, txn)
	}

	return transactions, nil
}

/**************************************************************************************
* listUsedTransactionIDs returns the ids of the transactions already recorded as paying
* one of the user's occurrences — a transaction must never pay two bills.
**************************************************************************************/
func listUsedTransactionIDs(userID uint) (map[string]bool, error) {
	rows, err := ModelsRepo.DB.Conn.Query(
		`SELECT DISTINCT paid_transaction_id FROM expense_occurrences
		 WHERE user_id = ? AND paid_transaction_id IS NOT NULL AND paid_transaction_id <> ''`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("could not list used transaction ids: %w", err)
	}
	defer rows.Close()

	used := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("could not scan used transaction id: %w", err)
		}
		used[id] = true
	}

	return used, nil
}

/**************************************************************************************
* matchTransactionsToOccurrences greedily pairs each occurrence (oldest due date first)
* with the earliest still-unused transaction on/after its due date whose merchant or name
* equals one of the occurrence's labels (case-insensitive). linkedOnly restricts matching
* to the user-linked label — that is the only pairing allowed to happen without the
* user's confirmation.
**************************************************************************************/
func matchTransactionsToOccurrences(occurrences []unpaidOccurrenceRow, transactions []OccurrenceTransactionMatch, used map[string]bool, linkedOnly bool) map[int64]*OccurrenceTransactionMatch {
	taken := make(map[string]bool, len(used))
	for id := range used {
		taken[id] = true
	}

	matches := make(map[int64]*OccurrenceTransactionMatch)
	for _, occurrence := range occurrences {
		labels := map[string]bool{}
		if occurrence.LinkedLabel != "" {
			labels[normalizeLabel(occurrence.LinkedLabel)] = true
		}
		if !linkedOnly {
			labels[normalizeLabel(occurrence.ExpenseLabel)] = true
		}

		for i := range transactions {
			txn := &transactions[i]
			if taken[txn.TransactionID] || txn.Date < occurrence.DueOn {
				continue
			}
			if !labels[normalizeLabel(txn.MerchantName)] && !labels[normalizeLabel(txn.Name)] {
				continue
			}
			matches[occurrence.Id] = txn
			taken[txn.TransactionID] = true
			break
		}
	}

	return matches
}

func normalizeLabel(label string) string {
	return strings.ToLower(strings.TrimSpace(label))
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
