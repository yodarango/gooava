package models

import (
	"database/sql"
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


/**************************************************************************************
* Update overwrites a category's fields. When the label changes, expenses linked by the
* old label string are re-pointed to the new one (expenses link to categories by label).
*
* status: ✅
**************************************************************************************/
func (c *ExpenseCategory) Update() error {
	var oldLabel string
	err := ModelsRepo.DB.Conn.QueryRow(
		"SELECT label FROM expense_categories WHERE id = ? AND user_id = ?",
		c.Id, c.UserID,
	).Scan(&oldLabel)
	if err == sql.ErrNoRows {
		return fmt.Errorf("category not found")
	}
	if err != nil {
		return fmt.Errorf("could not load expense category: %w", err)
	}

	if _, err := ModelsRepo.DB.Conn.Exec(
		"UPDATE expense_categories SET label = ?, description = ?, color = ? WHERE id = ? AND user_id = ?",
		c.Label, c.Description, c.Color, c.Id, c.UserID,
	); err != nil {
		return fmt.Errorf("could not update expense category: %w", err)
	}

	// re-point expenses linked by the old label string
	if oldLabel != c.Label {
		if _, err := ModelsRepo.DB.Conn.Exec(
			"UPDATE expenses SET category = ? WHERE user_id = ? AND category = ?",
			c.Label, c.UserID, oldLabel,
		); err != nil {
			return fmt.Errorf("could not re-link expenses to the renamed category: %w", err)
		}
	}

	return nil
}

/**************************************************************************************
* DeleteExpenseCategory removes a category by id. Expenses keep their category text —
* the link is by label string, so nothing else needs cleanup.
*
* status: ✅
**************************************************************************************/
func DeleteExpenseCategory(userID uint, id int64) error {
	result, err := ModelsRepo.DB.Conn.Exec(
		"DELETE FROM expense_categories WHERE id = ? AND user_id = ?",
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("could not delete expense category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not verify expense category deletion: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}
