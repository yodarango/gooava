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
