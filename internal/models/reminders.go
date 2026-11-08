package models

import (
	"fmt"
	"gooava/internal/lib"
	"strings"
	"time"
)

/**************************************************************************************
* StartBillReminderLoop runs the bill-reminder scheduler inside the app process: on boot
* and then hourly it emails the recipients of expenses with reminders on — 3 days before
* a bill is due, 1 day before, and every day while it's past due. The container is always
* up on the VPS, so reminders go out even when nobody loads the app.
*
* status: ✅
**************************************************************************************/
func StartBillReminderLoop() {
	go func() {
		runBillReminders()
		for range time.Tick(time.Hour) {
			runBillReminders()
		}
	}()
}

type reminderRow struct {
	OccurrenceID int64
	ExpenseID    int64
	UserID       uint
	DueOn        string
	Amount       float64
	Label        string
	Emails       string
	FirstName    string
}

/**************************************************************************************
* One sweep: every unpaid occurrence of a reminder-enabled expense whose due date is in
* the sending window (3 days out / 1 day out / any past day). The expense_reminders row
* (unique per occurrence + kind + day) gates duplicates — inserted before sending, and
* removed again if the send fails so the next tick retries.
**************************************************************************************/
func runBillReminders() {
	rows, err := ModelsRepo.DB.Conn.Query(`
		SELECT eo.id, eo.expense_id, eo.user_id, DATE_FORMAT(eo.due_on, '%Y-%m-%d'), eo.amount,
		       e.label, e.reminder_emails, u.first_name
		FROM expense_occurrences eo
		JOIN expenses e ON e.id = eo.expense_id
		JOIN users u ON u.id = eo.user_id
		WHERE eo.paid = 0
		  AND e.is_reminder_on = 1
		  AND e.reminder_emails IS NOT NULL AND e.reminder_emails <> ''
		  AND eo.due_on <= CURDATE() + INTERVAL 3 DAY
	`)
	if err != nil {
		fmt.Printf("bill reminders: could not load due occurrences: %v\n", err)
		return
	}
	defer rows.Close()

	reminders := make([]reminderRow, 0)
	for rows.Next() {
		var reminder reminderRow
		if err := rows.Scan(
			&reminder.OccurrenceID, &reminder.ExpenseID, &reminder.UserID,
			&reminder.DueOn, &reminder.Amount, &reminder.Label, &reminder.Emails,
			&reminder.FirstName,
		); err != nil {
			fmt.Printf("bill reminders: could not scan occurrence: %v\n", err)
			return
		}
		reminders = append(reminders, reminder)
	}

	today := time.Now().Truncate(24 * time.Hour)

	for _, reminder := range reminders {
		dueOn, err := time.Parse("2006-01-02", reminder.DueOn)
		if err != nil {
			continue
		}
		daysUntil := int(dueOn.Sub(today).Hours() / 24)

		var kind, timing string
		switch {
		case daysUntil == 3:
			kind, timing = "3d", "in 3 days"
		case daysUntil == 1:
			kind, timing = "1d", "tomorrow"
		case daysUntil < 0:
			kind, timing = "pastdue", "past due"
		default:
			continue // due today, tomorrow+1 edge, or 2 days out — no email
		}

		// claim the send slot first — the unique key makes this a no-op when today's
		// email of this kind for this occurrence already went out
		result, err := ModelsRepo.DB.Conn.Exec(
			`INSERT IGNORE INTO expense_reminders(user_id, expense_id, occurrence_id, kind, sent_for_date)
			 VALUES(?,?,?,?,CURDATE())`,
			reminder.UserID, reminder.ExpenseID, reminder.OccurrenceID, kind,
		)
		if err != nil {
			fmt.Printf("bill reminders: could not log reminder: %v\n", err)
			continue
		}
		claimed, err := result.RowsAffected()
		if err != nil || claimed == 0 {
			continue // already sent today
		}

		recipients := make([]string, 0)
		for _, email := range strings.Split(reminder.Emails, ",") {
			if trimmed := strings.TrimSpace(email); trimmed != "" {
				recipients = append(recipients, trimmed)
			}
		}
		if len(recipients) == 0 {
			continue
		}

		amount := fmt.Sprintf("$%.2f", reminder.Amount)
		dueDisplay := dueOn.Format("Jan 2")
		emailService := lib.NewEmailService()
		if err := emailService.SendBillReminderEmail(recipients, reminder.FirstName, reminder.Label, amount, dueDisplay, timing); err != nil {
			fmt.Printf("bill reminders: could not send %s reminder for %v: %v\n", kind, reminder.Label, err)
			// free the slot so the next tick retries instead of losing the reminder
			ModelsRepo.DB.Conn.Exec(
				"DELETE FROM expense_reminders WHERE occurrence_id = ? AND kind = ? AND sent_for_date = CURDATE()",
				reminder.OccurrenceID, kind,
			)
		}
	}
}
