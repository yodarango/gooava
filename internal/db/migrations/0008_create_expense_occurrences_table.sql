-- One entry per expense occurrence. The `expenses` row is the "type" (schedule + defaults);
-- this table is the concrete bill that must be paid. Recurring expenses get one row per due date,
-- one-time expenses get a single row. Rows are created lazily by the paycheck plan's "ensure"
-- endpoint the first time an occurrence falls inside the displayed window, and are marked paid
-- either automatically (label string-matches a synced Plaid transaction) or manually. Paid rows
-- are kept for display (crossed off) but excluded from the plan's totals.

CREATE TABLE expense_occurrences (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    expense_id INT NOT NULL,
    due_on DATE NOT NULL,
    amount DECIMAL(14,2) NOT NULL,
    paid TINYINT(1) NOT NULL DEFAULT 0,
    paid_source VARCHAR(8) DEFAULT NULL, -- NULL = unpaid; 'auto' = matched a transaction; 'manual' = user marked it
    paid_transaction_id VARCHAR(255) DEFAULT NULL, -- plaid transaction that auto-paid it (when source = 'auto')
    paid_at DATETIME DEFAULT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uniq_expense_occurrence (expense_id, due_on),
    KEY idx_expense_occurrences_user (user_id),
    CONSTRAINT fk_expense_occurrences_expense FOREIGN KEY (expense_id) REFERENCES expenses(id) ON DELETE CASCADE,
    CONSTRAINT fk_expense_occurrences_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
