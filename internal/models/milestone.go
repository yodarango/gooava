package models

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type Milestone struct {
	Id           int64   `json:"id"`
	UserID       uint    `json:"user_id"`
	ExpenseID    int64   `json:"expense_id"`
	Title        string  `json:"title"`
	TargetAmount float64 `json:"target_amount"`
	TargetDate   string  `json:"target_date"` // YYYY-MM-DD (original date)
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type MilestoneWithProgress struct {
	Id                     int64   `json:"id"`
	UserID                 uint    `json:"user_id"`
	ExpenseID              int64   `json:"expense_id"`
	Title                  string  `json:"title"`
	TargetAmount           float64 `json:"target_amount"`
	TargetDate             string  `json:"target_date"` // original completion date
	AmountPerPaycheck      float64 `json:"amount_per_paycheck"`
	TotalSaved             float64 `json:"total_saved"`
	RemainingAmount        float64 `json:"remaining_amount"`
	Percentage             float64 `json:"percentage"`
	ExpectedCompletionDate string  `json:"expected_completion_date"`
	IsCompleted            bool    `json:"is_completed"`
	PaychecksRemaining     int     `json:"paychecks_remaining"`
	PaychecksBehind        int     `json:"paychecks_behind"`
	NextPayday             string  `json:"next_payday"`
	CreatedAt              string  `json:"created_at"`
	UpdatedAt              string  `json:"updated_at"`
}

/**************************************************************************************
* CreateMilestone creates a new savings goal, determines the amount per paycheck until
* the target date, creates a linked expense (category 'milestone'), and schedules its
* payments starting on upcoming paydays.
**************************************************************************************/
func CreateMilestone(userID uint, title string, targetAmount float64, targetDate string) (*MilestoneWithProgress, error) {
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if targetAmount <= 0 {
		return nil, fmt.Errorf("target amount must be greater than 0")
	}
	if _, err := time.Parse("2006-01-02", targetDate); err != nil {
		return nil, fmt.Errorf("target date must be YYYY-MM-DD")
	}

	now := time.Now()
	paydays, err := GetUpcomingPaydays(userID, now, 120)
	if err != nil {
		return nil, err
	}

	paychecksCount := 0
	for _, p := range paydays {
		if p <= targetDate {
			paychecksCount++
		}
	}
	if paychecksCount < 1 {
		paychecksCount = 1
	}

	amountPerPaycheck := round2(targetAmount / float64(paychecksCount))
	if amountPerPaycheck < 0.01 {
		amountPerPaycheck = 0.01
	}

	// create the linked expense
	expense := Expense{
		UserID:    userID,
		Category:  "milestone",
		Amount:    amountPerPaycheck,
		Label:     title,
		Recurring: "biweekly",
		Notes:     fmt.Sprintf("Savings milestone: %s (target %s)", title, targetDate),
	}
	if err := expense.Create(); err != nil {
		return nil, fmt.Errorf("could not create milestone expense: %w", err)
	}

	result, err := ModelsRepo.DB.Conn.Exec(
		`INSERT INTO milestones(user_id, expense_id, title, target_amount, target_date)
		 VALUES(?,?,?,?,?)`,
		userID, expense.Id, title, targetAmount, targetDate,
	)
	if err != nil {
		_ = DeleteExpense(userID, expense.Id)
		return nil, fmt.Errorf("could not create milestone: %w", err)
	}

	milestoneID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("could not get milestone id: %w", err)
	}

	if err := syncSingleMilestone(userID, milestoneID, expense.Id, targetAmount, amountPerPaycheck, targetDate); err != nil {
		return nil, err
	}

	return GetMilestone(userID, milestoneID)
}

/**************************************************************************************
* GetMilestone loads one milestone enriched with saved amounts, completion progress,
* and expected completion date.
**************************************************************************************/
func GetMilestone(userID uint, id int64) (*MilestoneWithProgress, error) {
	var m Milestone
	err := ModelsRepo.DB.Conn.QueryRow(
		`SELECT id, user_id, expense_id, title, target_amount, DATE_FORMAT(target_date, '%Y-%m-%d'),
		        DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at, '%Y-%m-%d %H:%i:%s')
		 FROM milestones WHERE id = ? AND user_id = ?`,
		id, userID,
	).Scan(&m.Id, &m.UserID, &m.ExpenseID, &m.Title, &m.TargetAmount, &m.TargetDate, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("milestone not found: %w", err)
	}

	return enrichMilestone(m)
}

/**************************************************************************************
* ListMilestones returns all milestones for the user, synced and enriched with progress.
**************************************************************************************/
func ListMilestones(userID uint) ([]MilestoneWithProgress, error) {
	if err := SyncMilestones(userID); err != nil {
		fmt.Printf("warning: sync milestones failed: %v\n", err)
	}

	rows, err := ModelsRepo.DB.Conn.Query(
		`SELECT id, user_id, expense_id, title, target_amount, DATE_FORMAT(target_date, '%Y-%m-%d'),
		        DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s'), DATE_FORMAT(updated_at, '%Y-%m-%d %H:%i:%s')
		 FROM milestones WHERE user_id = ? ORDER BY id DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("could not list milestones: %w", err)
	}
	defer rows.Close()

	list := make([]MilestoneWithProgress, 0)
	for rows.Next() {
		var m Milestone
		if err := rows.Scan(&m.Id, &m.UserID, &m.ExpenseID, &m.Title, &m.TargetAmount, &m.TargetDate, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("could not scan milestone: %w", err)
		}
		enriched, err := enrichMilestone(m)
		if err != nil {
			return nil, err
		}
		list = append(list, *enriched)
	}

	return list, nil
}

/**************************************************************************************
* UpdateMilestone updates title, target amount, or target date, recomputing the
* contribution per paycheck and resyncing future occurrences.
**************************************************************************************/
func UpdateMilestone(userID uint, id int64, title string, targetAmount float64, targetDate string) (*MilestoneWithProgress, error) {
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if targetAmount <= 0 {
		return nil, fmt.Errorf("target amount must be greater than 0")
	}
	if _, err := time.Parse("2006-01-02", targetDate); err != nil {
		return nil, fmt.Errorf("target date must be YYYY-MM-DD")
	}

	var expenseID int64
	err := ModelsRepo.DB.Conn.QueryRow(
		"SELECT expense_id FROM milestones WHERE id = ? AND user_id = ?",
		id, userID,
	).Scan(&expenseID)
	if err != nil {
		return nil, fmt.Errorf("milestone not found: %w", err)
	}

	now := time.Now()
	paydays, err := GetUpcomingPaydays(userID, now, 120)
	if err != nil {
		return nil, err
	}

	// calculate how many remaining paychecks until targetDate
	paychecksCount := 0
	for _, p := range paydays {
		if p <= targetDate {
			paychecksCount++
		}
	}
	if paychecksCount < 1 {
		paychecksCount = 1
	}

	var totalSaved float64
	_ = ModelsRepo.DB.Conn.QueryRow(
		"SELECT COALESCE(SUM(amount), 0) FROM expense_occurrences WHERE user_id = ? AND expense_id = ? AND paid = 1",
		userID, expenseID,
	).Scan(&totalSaved)

	remaining := targetAmount - totalSaved
	if remaining < 0 {
		remaining = 0
	}
	amountPerPaycheck := round2(remaining / float64(paychecksCount))
	if amountPerPaycheck < 0.01 {
		amountPerPaycheck = 0.01
	}

	// update milestone row
	_, err = ModelsRepo.DB.Conn.Exec(
		`UPDATE milestones SET title = ?, target_amount = ?, target_date = ? WHERE id = ? AND user_id = ?`,
		title, targetAmount, targetDate, id, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("could not update milestone: %w", err)
	}

	// update linked expense
	_, _ = ModelsRepo.DB.Conn.Exec(
		`UPDATE expenses SET label = ?, amount = ? WHERE id = ? AND user_id = ?`,
		title, amountPerPaycheck, expenseID, userID,
	)

	if err := syncSingleMilestone(userID, id, expenseID, targetAmount, amountPerPaycheck, targetDate); err != nil {
		return nil, err
	}

	return GetMilestone(userID, id)
}

/**************************************************************************************
* DeleteMilestone deletes a milestone and its associated expense record (which cascades
* all occurrence rows in expense_occurrences).
**************************************************************************************/
func DeleteMilestone(userID uint, id int64) error {
	var expenseID int64
	err := ModelsRepo.DB.Conn.QueryRow(
		"SELECT expense_id FROM milestones WHERE id = ? AND user_id = ?",
		id, userID,
	).Scan(&expenseID)
	if err != nil {
		return fmt.Errorf("milestone not found: %w", err)
	}

	if _, err := ModelsRepo.DB.Conn.Exec("DELETE FROM milestones WHERE id = ? AND user_id = ?", id, userID); err != nil {
		return fmt.Errorf("could not delete milestone: %w", err)
	}

	// deleting the expense cascades occurrences
	_, _ = ModelsRepo.DB.Conn.Exec("DELETE FROM expenses WHERE id = ? AND user_id = ?", expenseID, userID)

	return nil
}

/**************************************************************************************
* enrichMilestone loads the total saved from paid occurrences, computes the remaining
* balance, and calculates the expected completion date vs the original target date.
**************************************************************************************/
func enrichMilestone(m Milestone) (*MilestoneWithProgress, error) {
	var amountPerPaycheck float64
	_ = ModelsRepo.DB.Conn.QueryRow(
		"SELECT COALESCE(amount, 0) FROM expenses WHERE id = ? AND user_id = ?",
		m.ExpenseID, m.UserID,
	).Scan(&amountPerPaycheck)

	var totalSaved float64
	_ = ModelsRepo.DB.Conn.QueryRow(
		"SELECT COALESCE(SUM(amount), 0) FROM expense_occurrences WHERE user_id = ? AND expense_id = ? AND paid = 1",
		m.UserID, m.ExpenseID,
	).Scan(&totalSaved)

	totalSaved = round2(totalSaved)
	remaining := round2(m.TargetAmount - totalSaved)
	if remaining < 0 {
		remaining = 0
	}

	pct := 0.0
	if m.TargetAmount > 0 {
		pct = math.Round((totalSaved/m.TargetAmount)*1000) / 10
		if pct > 100 {
			pct = 100
		}
	}

	isCompleted := totalSaved >= m.TargetAmount

	now := time.Now()
	paydays, err := GetUpcomingPaydays(m.UserID, now, 120)
	if err != nil {
		return nil, err
	}

	nextPayday := ""
	if len(paydays) > 0 {
		nextPayday = paydays[0]
	}

	expectedDate := m.TargetDate
	paychecksRemaining := 0
	paychecksBehind := 0

	if isCompleted {
		// find date of the latest paid occurrence
		var lastPaid string
		err := ModelsRepo.DB.Conn.QueryRow(
			`SELECT COALESCE(DATE_FORMAT(MAX(due_on), '%Y-%m-%d'), ?) FROM expense_occurrences WHERE user_id = ? AND expense_id = ? AND paid = 1`,
			m.TargetDate, m.UserID, m.ExpenseID,
		).Scan(&lastPaid)
		if err == nil && lastPaid != "" {
			expectedDate = lastPaid
		}
	} else if amountPerPaycheck > 0 {
		paychecksRemaining = int(math.Ceil(remaining / amountPerPaycheck))
		if paychecksRemaining < 1 {
			paychecksRemaining = 1
		}

		if paychecksRemaining <= len(paydays) {
			expectedDate = paydays[paychecksRemaining-1]
		} else {
			// project further if needed
			extended, _ := GetUpcomingPaydays(m.UserID, now, paychecksRemaining+5)
			if paychecksRemaining <= len(extended) {
				expectedDate = extended[paychecksRemaining-1]
			}
		}

		// compare expected completion date vs original target date
		if expectedDate > m.TargetDate {
			// count how many paydays past original target date
			for _, p := range paydays {
				if p > m.TargetDate && p <= expectedDate {
					paychecksBehind++
				}
			}
		}
	}

	return &MilestoneWithProgress{
		Id:                     m.Id,
		UserID:                 m.UserID,
		ExpenseID:              m.ExpenseID,
		Title:                  m.Title,
		TargetAmount:           m.TargetAmount,
		TargetDate:             m.TargetDate,
		AmountPerPaycheck:      amountPerPaycheck,
		TotalSaved:             totalSaved,
		RemainingAmount:        remaining,
		Percentage:             pct,
		ExpectedCompletionDate: expectedDate,
		IsCompleted:            isCompleted,
		PaychecksRemaining:     paychecksRemaining,
		PaychecksBehind:        paychecksBehind,
		NextPayday:             nextPayday,
		CreatedAt:              m.CreatedAt,
		UpdatedAt:              m.UpdatedAt,
	}, nil
}

/**************************************************************************************
* syncSingleMilestone ensures milestone occurrences are kept in sync with reality:
* - Unpaid occurrences in the past (skipped paychecks) are deleted so they NEVER show
*   as late or overdue.
* - Remaining required payments are scheduled starting on current/upcoming paydays.
* - When the target amount has been reached, all future unpaid rows are removed.
**************************************************************************************/
func syncSingleMilestone(userID uint, milestoneID, expenseID int64, targetAmount, amountPerPaycheck float64, targetDate string) error {
	today := time.Now().Format("2006-01-02")

	// 1. Delete any unpaid occurrences in the past (skipped payments are not overdue; they roll forward)
	_, err := ModelsRepo.DB.Conn.Exec(
		`DELETE FROM expense_occurrences
		 WHERE user_id = ? AND expense_id = ? AND paid = 0 AND due_on < ?`,
		userID, expenseID, today,
	)
	if err != nil {
		return fmt.Errorf("could not clear skipped milestone occurrences: %w", err)
	}

	// 2. Check total amount already saved (paid occurrences)
	var totalSaved float64
	_ = ModelsRepo.DB.Conn.QueryRow(
		`SELECT COALESCE(SUM(amount), 0) FROM expense_occurrences
		 WHERE user_id = ? AND expense_id = ? AND paid = 1`,
		userID, expenseID,
	).Scan(&totalSaved)

	remaining := round2(targetAmount - totalSaved)
	if remaining <= 0 {
		// Milestone already reached! Clean up any remaining future unpaid rows
		_, _ = ModelsRepo.DB.Conn.Exec(
			`DELETE FROM expense_occurrences WHERE user_id = ? AND expense_id = ? AND paid = 0`,
			userID, expenseID,
		)
		return nil
	}

	if amountPerPaycheck <= 0 {
		return nil
	}

	// 3. Get upcoming paydays starting from today
	now := time.Now()
	paydays, err := GetUpcomingPaydays(userID, now, 120)
	if err != nil {
		return err
	}

	// 4. Ensure unpaid occurrences for upcoming paydays until remaining is satisfied
	accumulated := 0.0
	scheduledDueDates := make(map[string]bool)
	var lastDueDate string

	for _, p := range paydays {
		if accumulated >= remaining {
			break
		}
		toPay := round2(math.Min(amountPerPaycheck, remaining-accumulated))
		if toPay <= 0 {
			break
		}

		// Insert or update unpaid row for this payday
		_, err := ModelsRepo.DB.Conn.Exec(
			`INSERT INTO expense_occurrences(user_id, expense_id, due_on, amount, paid)
			 VALUES(?, ?, ?, ?, 0)
			 ON DUPLICATE KEY UPDATE
			   amount = IF(paid = 0, VALUES(amount), amount)`,
			userID, expenseID, p, toPay,
		)
		if err != nil {
			return fmt.Errorf("could not schedule milestone occurrence on %s: %w", p, err)
		}

		scheduledDueDates[p] = true
		lastDueDate = p
		accumulated += toPay
	}

	// 5. Delete any future unpaid occurrences beyond what is needed to reach the goal
	if lastDueDate != "" {
		_, _ = ModelsRepo.DB.Conn.Exec(
			`DELETE FROM expense_occurrences
			 WHERE user_id = ? AND expense_id = ? AND paid = 0 AND due_on > ?`,
			userID, expenseID, lastDueDate,
		)
	}

	return nil
}

/**************************************************************************************
* SyncMilestones runs syncSingleMilestone for every milestone owned by the user.
**************************************************************************************/
func SyncMilestones(userID uint) error {
	rows, err := ModelsRepo.DB.Conn.Query(
		`SELECT m.id, m.expense_id, m.target_amount, COALESCE(e.amount, 0), DATE_FORMAT(m.target_date, '%Y-%m-%d')
		 FROM milestones m
		 JOIN expenses e ON e.id = m.expense_id
		 WHERE m.user_id = ?`,
		userID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	type mItem struct {
		id                int64
		expenseID         int64
		targetAmount      float64
		amountPerPaycheck float64
		targetDate        string
	}
	items := make([]mItem, 0)
	for rows.Next() {
		var item mItem
		if err := rows.Scan(&item.id, &item.expenseID, &item.targetAmount, &item.amountPerPaycheck, &item.targetDate); err == nil {
			items = append(items, item)
		}
	}

	for _, item := range items {
		_ = syncSingleMilestone(userID, item.id, item.expenseID, item.targetAmount, item.amountPerPaycheck, item.targetDate)
	}

	return nil
}

/**************************************************************************************
* GetUpcomingPaydays returns the user's upcoming paydays sorted chronologically on or
* after the given date, derived from active recurring income schedules or fallback anchor.
**************************************************************************************/
func GetUpcomingPaydays(userID uint, from time.Time, count int) ([]string, error) {
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.Local)
	incomes, err := ListIncomes(userID)
	if err != nil {
		return nil, err
	}

	dateSet := make(map[string]bool)
	hasRecurring := false
	for _, inc := range incomes {
		if inc.Recurring != "" {
			hasRecurring = true
			dates := nextIncomeDates(inc, from, count*2)
			for _, d := range dates {
				dateSet[d] = true
			}
		}
	}

	if !hasRecurring {
		// Fallback: every other Friday, anchor 2026-10-02
		anchor := time.Date(2026, 10, 2, 0, 0, 0, 0, time.Local)
		diffDays := int(math.Round(from.Sub(anchor).Hours() / 24))
		sinceLast := ((diffDays % 14) + 14) % 14
		daysUntil := (14 - sinceLast) % 14
		first := from.AddDate(0, 0, daysUntil)
		for i := 0; i < count*2; i++ {
			dateSet[first.AddDate(0, 0, i*14).Format("2006-01-02")] = true
		}
	}

	allDates := make([]string, 0, len(dateSet))
	for d := range dateSet {
		allDates = append(allDates, d)
	}
	sort.Strings(allDates)
	if len(allDates) > count {
		allDates = allDates[:count]
	}
	return allDates, nil
}
