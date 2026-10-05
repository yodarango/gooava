-- Income sources: fixed (recurring schedule in recur_rule) or one-time (recurring NULL,
-- recur_rule {"date":"YYYY-MM-DD"}). Fixed-income occurrences must be approved before they
-- count toward the paycheck plan; every decision is recorded in income_entries.

CREATE TABLE incomes (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    label VARCHAR(512) NOT NULL,
    amount DECIMAL(14,2) NOT NULL,
    recurring VARCHAR(16) DEFAULT NULL, -- NULL = one-time; weekly, biweekly, monthly, quarterly, semiannually, annually
    recur_rule JSON DEFAULT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_incomes_user (user_id),
    CONSTRAINT fk_incomes_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Approval ledger: one row per income occurrence the user approved or rejected.
-- Rejected rows stay so the occurrence never comes back as pending; a re-approval of the
-- same date deletes them (see ApproveIncomeEntry).
CREATE TABLE income_entries (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    income_id INT NOT NULL,
    occurred_on DATE NOT NULL,
    amount DECIMAL(14,2) NOT NULL,
    status VARCHAR(8) NOT NULL DEFAULT 'approved', -- approved | rejected
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uniq_income_entry_approved (income_id, occurred_on, status),
    KEY idx_income_entries_user (user_id),
    CONSTRAINT fk_income_entries_income FOREIGN KEY (income_id) REFERENCES incomes(id) ON DELETE CASCADE,
    CONSTRAINT fk_income_entries_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
