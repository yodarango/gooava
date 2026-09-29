-- Plaid bank integration tables (read-only: link, sync, list)

CREATE TABLE plaid_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    item_id VARCHAR(255) NOT NULL,
    access_token TEXT NOT NULL,
    `cursor` TEXT DEFAULT NULL,
    institution_name VARCHAR(255) DEFAULT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uniq_plaid_items_item_id (item_id),
    CONSTRAINT fk_plaid_items_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE plaid_accounts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    plaid_item_id INT NOT NULL,
    account_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    mask VARCHAR(16) DEFAULT NULL,
    type VARCHAR(64) DEFAULT NULL,
    subtype VARCHAR(64) DEFAULT NULL,
    current_balance DECIMAL(14,2) DEFAULT NULL,
    available_balance DECIMAL(14,2) DEFAULT NULL,
    iso_currency_code CHAR(3) DEFAULT 'USD',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uniq_plaid_accounts_account_id (account_id),
    CONSTRAINT fk_plaid_accounts_item FOREIGN KEY (plaid_item_id) REFERENCES plaid_items(id) ON DELETE CASCADE,
    CONSTRAINT fk_plaid_accounts_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE plaid_transactions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    plaid_account_id INT NOT NULL,
    transaction_id VARCHAR(255) NOT NULL,
    name VARCHAR(512) NOT NULL,
    merchant_name VARCHAR(255) DEFAULT NULL,
    amount DECIMAL(14,2) NOT NULL,
    iso_currency_code CHAR(3) DEFAULT 'USD',
    `date` DATE NOT NULL,
    category VARCHAR(255) DEFAULT NULL,
    pending TINYINT(1) NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uniq_plaid_transactions_transaction_id (transaction_id),
    KEY idx_plaid_transactions_user_date (user_id, `date`),
    CONSTRAINT fk_plaid_txn_account FOREIGN KEY (plaid_account_id) REFERENCES plaid_accounts(id) ON DELETE CASCADE,
    CONSTRAINT fk_plaid_txn_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
