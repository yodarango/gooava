package models

import (
	"fmt"
)

type ExpenseCategory struct {
	Id          int64  `json:"id"`
	UserID      uint   `json:"user_id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Color       string `json:"color"`
	CreatedAt   string `json:"created_at"`
}

/**************************************************************************************
* Creates a category for the user. Labels are unique per user since expenses link to
* categories by the label string.
*
* status: ✅
**************************************************************************************/
func (c *ExpenseCategory) Create() error {
	query := `
		INSERT INTO expense_categories(user_id, label, description, color)
		VALUES(?,?,?,?)
	`

	result, err := ModelsRepo.DB.Conn.Exec(query, c.UserID, c.Label, c.Description, c.Color)
	if err != nil {
		return fmt.Errorf("could not create expense category: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("could not get last expense category id: %w", err)
	}

	c.Id = id

	return nil
}

/**************************************************************************************
* ListExpenseCategories returns the user's categories, alphabetically by label.
**************************************************************************************/
func ListExpenseCategories(userID uint) ([]ExpenseCategory, error) {
	query := `
		SELECT id, label, COALESCE(description, ''), color
		FROM expense_categories
		WHERE user_id = ?
		ORDER BY label
	`

	rows, err := ModelsRepo.DB.Conn.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("could not list expense categories: %w", err)
	}
	defer rows.Close()

	categories := make([]ExpenseCategory, 0)
	for rows.Next() {
		var category ExpenseCategory
		err := rows.Scan(&category.Id, &category.Label, &category.Description, &category.Color)
		if err != nil {
			return nil, fmt.Errorf("could not scan expense category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}
