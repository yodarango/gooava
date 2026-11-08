-- Savings Milestones: financial goals with target amounts, target dates, and automated per-paycheck contributions.
-- Each milestone links to an expense record (category 'milestone') and its paid occurrences contribute to the goal.

CREATE TABLE milestones (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    expense_id INT NOT NULL,
    title VARCHAR(255) NOT NULL,
    target_amount DECIMAL(14,2) NOT NULL,
    target_date DATE NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_milestones_user (user_id),
    KEY idx_milestones_expense (expense_id),
    CONSTRAINT fk_milestones_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_milestones_expense FOREIGN KEY (expense_id) REFERENCES expenses(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
