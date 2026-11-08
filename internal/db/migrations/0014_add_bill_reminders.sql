-- Bill reminders: per-expense switch + comma-separated recipient list, and a send log
-- so the scheduler never double-sends (3 days out, 1 day out, then daily while past due —
-- the unique key gates one email per occurrence per kind per day).

ALTER TABLE expenses
    ADD COLUMN is_reminder_on TINYINT(1) NOT NULL DEFAULT 0 AFTER notes,
    ADD COLUMN reminder_emails TEXT DEFAULT NULL AFTER is_reminder_on;

CREATE TABLE expense_reminders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    expense_id INT NOT NULL,
    occurrence_id INT NOT NULL,
    kind VARCHAR(8) NOT NULL,      -- '3d' | '1d' | 'pastdue'
    sent_for_date DATE NOT NULL,   -- the day this send covers (pastdue sends once per day)
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uniq_expense_reminders_once (occurrence_id, kind, sent_for_date),
    KEY idx_expense_reminders_user (user_id),
    CONSTRAINT fk_expense_reminders_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_expense_reminders_expense FOREIGN KEY (expense_id) REFERENCES expenses(id) ON DELETE CASCADE,
    CONSTRAINT fk_expense_reminders_occurrence FOREIGN KEY (occurrence_id) REFERENCES expense_occurrences(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
