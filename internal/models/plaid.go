package models

import (
	"fmt"
	"gooava/internal/utils"
	"strings"
)

type PlaidItem struct {
	Id              int64  `json:"id"`
	UserID          uint   `json:"user_id"`
	ItemID          string `json:"item_id"`
	AccessToken     string `json:"-"`
	Cursor          string `json:"-"`
	InstitutionName string `json:"institution_name"`
}

type PlaidAccountRow struct {
	AccountID        string   `json:"account_id"`
	PlaidItemID      int64    `json:"plaid_item_id"`
	Name             string   `json:"name"`
	Mask             string   `json:"mask"`
	Type             string   `json:"type"`
	Subtype          string   `json:"subtype"`
	IsoCurrencyCode  string   `json:"iso_currency_code"`
	CurrentBalance   float64  `json:"current_balance"`
	AvailableBalance *float64 `json:"available_balance"`
}

type PlaidTransactionRow struct {
	TransactionID   string  `json:"transaction_id"`
	AccountID       string  `json:"account_id"`
	Name            string  `json:"name"`
	MerchantName    string  `json:"merchant_name"`
	Amount          float64 `json:"amount"`
	IsoCurrencyCode string  `json:"iso_currency_code"`
	Date            string  `json:"date"`
	Category        string  `json:"category"`
	Pending         bool    `json:"pending"`
}

// PlaidTransactionWithAccount is a transaction row joined with its account info,
// used by the list endpoint.
type PlaidTransactionWithAccount struct {
	PlaidTransactionRow
	AccountName string `json:"account_name"`
	AccountMask string `json:"account_mask"`
}

/**************************************************************************************
* Access tokens are stored AES-256-GCM-encrypted (see internal/utils/crypto.go).
* Legacy rows hold plaintext tokens, which always start with "access-" — on first
* read they are returned as-is and transparently re-encrypted in the database.
**************************************************************************************/
func storeAccessToken(plaintext string) (string, error) {
	encrypted, err := utils.EncryptString(plaintext)
	if err != nil {
		return "", fmt.Errorf("could not encrypt access token: %w", err)
	}
	return encrypted, nil
}

func loadAccessToken(itemDBID int64, stored string) (string, error) {
	if strings.HasPrefix(stored, "access-") {
		// legacy plaintext row — re-encrypt it now so the plaintext does not linger
		if encrypted, err := utils.EncryptString(stored); err == nil {
			if _, err := ModelsRepo.DB.Conn.Exec("UPDATE plaid_items SET access_token = ? WHERE id = ?", encrypted, itemDBID); err != nil {
				fmt.Printf("warning: could not re-encrypt legacy plaid token: %v\n", err)
			}
		}
		return stored, nil
	}

	token, err := utils.DecryptString(stored)
	if err != nil {
		return "", fmt.Errorf("could not decrypt access token: %w", err)
	}
	return token, nil
}

/**************************************************************************************
* SavePlaidItem inserts (or updates on duplicate item_id) a Plaid item for a user and
* returns the database row id. The access token is encrypted before it is stored.
*
* status: ✅
**************************************************************************************/
func SavePlaidItem(userID uint, itemID, accessToken, institutionName string) (int64, error) {
	// never write the raw token — it is a live credential for the user's bank
	encryptedToken, err := storeAccessToken(accessToken)
	if err != nil {
		return 0, err
	}

	query := `
		INSERT INTO plaid_items(user_id, item_id, access_token, institution_name)
		VALUES(?,?,?,?)
		ON DUPLICATE KEY UPDATE access_token = VALUES(access_token), institution_name = VALUES(institution_name)
	`

	_, err = ModelsRepo.DB.Conn.Exec(query, userID, itemID, encryptedToken, institutionName)
	if err != nil {
		return 0, fmt.Errorf("could not save plaid item: %w", err)
	}

	var id int64
	err = ModelsRepo.DB.Conn.QueryRow("SELECT id FROM plaid_items WHERE item_id = ?", itemID).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("could not fetch plaid item id: %w", err)
	}

	return id, nil
}

/**************************************************************************************
* GetPlaidItemByUser returns the user's first Plaid item, or sql.ErrNoRows when the
* user has not connected a bank yet.
*
* status: ✅
**************************************************************************************/
func GetPlaidItemByUser(userID uint) (*PlaidItem, error) {
	query := "SELECT id, user_id, item_id, access_token, COALESCE(`cursor`, ''), COALESCE(institution_name, '') " +
		"FROM plaid_items WHERE user_id = ? ORDER BY id LIMIT 1"

	var item PlaidItem
	row := ModelsRepo.DB.Conn.QueryRow(query, userID)
	err := row.Scan(&item.Id, &item.UserID, &item.ItemID, &item.AccessToken, &item.Cursor, &item.InstitutionName)
	if err != nil {
		return nil, err
	}

	// decrypt the stored token (legacy plaintext rows get re-encrypted on the fly)
	plaintextToken, err := loadAccessToken(item.Id, item.AccessToken)
	if err != nil {
		return nil, err
	}
	item.AccessToken = plaintextToken

	return &item, nil
}

/**************************************************************************************
* GetPlaidCursor returns the transactions/sync cursor stored on the user's item
* ("" when no sync has happened yet).
**************************************************************************************/
func GetPlaidCursor(userID uint) (string, error) {
	item, err := GetPlaidItemByUser(userID)
	if err != nil {
		return "", err
	}

	return item.Cursor, nil
}

/**************************************************************************************
* SavePlaidCursor stores the latest transactions/sync cursor on a Plaid item.
**************************************************************************************/
func SavePlaidCursor(itemDBID int64, cursor string) error {
	query := "UPDATE plaid_items SET `cursor` = ? WHERE id = ?"

	_, err := ModelsRepo.DB.Conn.Exec(query, cursor, itemDBID)
	if err != nil {
		return fmt.Errorf("could not save plaid cursor: %w", err)
	}

	return nil
}

/**************************************************************************************
* ListPlaidItems returns every Plaid item (bank connection) the user has, oldest first.
* Never includes the access token.
**************************************************************************************/
func ListPlaidItems(userID uint) ([]PlaidItem, error) {
	query := "SELECT id, user_id, item_id, COALESCE(institution_name, '') FROM plaid_items WHERE user_id = ? ORDER BY id"

	rows, err := ModelsRepo.DB.Conn.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("could not list plaid items: %w", err)
	}
	defer rows.Close()

	items := make([]PlaidItem, 0)
	for rows.Next() {
		var item PlaidItem
		if err := rows.Scan(&item.Id, &item.UserID, &item.ItemID, &item.InstitutionName); err != nil {
			return nil, fmt.Errorf("could not scan plaid item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

/**************************************************************************************
* GetPlaidItemById returns one of the user's Plaid items by its database id
* (including the access token and sync cursor), or sql.ErrNoRows when not found.
**************************************************************************************/
func GetPlaidItemById(userID uint, itemDBID int64) (*PlaidItem, error) {
	query := "SELECT id, user_id, item_id, access_token, COALESCE(`cursor`, ''), COALESCE(institution_name, '') " +
		"FROM plaid_items WHERE id = ? AND user_id = ?"

	var item PlaidItem
	row := ModelsRepo.DB.Conn.QueryRow(query, itemDBID, userID)
	err := row.Scan(&item.Id, &item.UserID, &item.ItemID, &item.AccessToken, &item.Cursor, &item.InstitutionName)
	if err != nil {
		return nil, err
	}

	// decrypt the stored token (legacy plaintext rows get re-encrypted on the fly)
	plaintextToken, err := loadAccessToken(item.Id, item.AccessToken)
	if err != nil {
		return nil, err
	}
	item.AccessToken = plaintextToken

	return &item, nil
}

/**************************************************************************************
* SavePlaidAccounts upserts the accounts belonging to a Plaid item (keyed on the
* Plaid account_id).
**************************************************************************************/
func SavePlaidAccounts(userID uint, itemDBID int64, accounts []PlaidAccountRow) error {
	query := `
		INSERT INTO plaid_accounts(
			user_id, plaid_item_id, account_id, name, mask, type, subtype,
			current_balance, available_balance, iso_currency_code
		)
		VALUES(?,?,?,?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name), mask = VALUES(mask), type = VALUES(type), subtype = VALUES(subtype),
			current_balance = VALUES(current_balance), available_balance = VALUES(available_balance),
			iso_currency_code = VALUES(iso_currency_code)
	`

	for _, account := range accounts {
		_, err := ModelsRepo.DB.Conn.Exec(query,
			userID, itemDBID, account.AccountID, account.Name, account.Mask, account.Type, account.Subtype,
			account.CurrentBalance, account.AvailableBalance, account.IsoCurrencyCode,
		)
		if err != nil {
			return fmt.Errorf("could not save plaid account %s: %w", account.AccountID, err)
		}
	}

	return nil
}

/**************************************************************************************
* UpsertPlaidTransactions inserts or updates transactions (keyed on the Plaid
* transaction_id), resolving each Plaid account_id to its database row id.
**************************************************************************************/
func UpsertPlaidTransactions(userID uint, txns []PlaidTransactionRow) error {
	if len(txns) == 0 {
		return nil
	}

	accountDBIDs := make(map[string]int64)
	rows, err := ModelsRepo.DB.Conn.Query("SELECT id, account_id FROM plaid_accounts WHERE user_id = ?", userID)
	if err != nil {
		return fmt.Errorf("could not load plaid accounts: %w", err)
	}

	for rows.Next() {
		var id int64
		var accountID string
		if err := rows.Scan(&id, &accountID); err != nil {
			rows.Close()
			return fmt.Errorf("could not scan plaid account: %w", err)
		}
		accountDBIDs[accountID] = id
	}
	rows.Close()

	query := `
		INSERT INTO plaid_transactions(
			user_id, plaid_account_id, transaction_id, name, merchant_name,
			amount, iso_currency_code, ` + "`date`" + `, category, pending
		)
		VALUES(?,?,?,?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name), merchant_name = VALUES(merchant_name), amount = VALUES(amount),
			category = VALUES(category), pending = VALUES(pending)
	`

	for _, txn := range txns {
		accountDBID, ok := accountDBIDs[txn.AccountID]
		if !ok {
			continue // transaction for an account we have not stored — skip
		}

		_, err := ModelsRepo.DB.Conn.Exec(query,
			userID, accountDBID, txn.TransactionID, txn.Name, txn.MerchantName,
			txn.Amount, txn.IsoCurrencyCode, txn.Date, txn.Category, txn.Pending,
		)
		if err != nil {
			return fmt.Errorf("could not upsert plaid transaction %s: %w", txn.TransactionID, err)
		}
	}

	return nil
}

/**************************************************************************************
* DeletePlaidTransactions removes transactions that Plaid reported as removed.
**************************************************************************************/
func DeletePlaidTransactions(userID uint, transactionIDs []string) error {
	if len(transactionIDs) == 0 {
		return nil
	}

	placeholders := make([]string, len(transactionIDs))
	args := make([]interface{}, 0, len(transactionIDs)+1)
	args = append(args, userID)
	for i, id := range transactionIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := fmt.Sprintf(
		"DELETE FROM plaid_transactions WHERE user_id = ? AND transaction_id IN (%s)",
		strings.Join(placeholders, ","),
	)

	_, err := ModelsRepo.DB.Conn.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("could not delete plaid transactions: %w", err)
	}

	return nil
}

/**************************************************************************************
* GetPlaidTransactionLabel returns the display label of one of the user's transactions —
* the merchant name when present, the transaction name otherwise. Used when linking an
* expense to a transaction so future occurrences can be matched by that label.
**************************************************************************************/
func GetPlaidTransactionLabel(userID uint, transactionID string) (string, error) {
	var label string
	err := ModelsRepo.DB.Conn.QueryRow(
		`SELECT COALESCE(NULLIF(TRIM(COALESCE(merchant_name, '')), ''), TRIM(name))
		 FROM plaid_transactions
		 WHERE user_id = ? AND transaction_id = ?`,
		userID, transactionID,
	).Scan(&label)
	if err != nil {
		return "", fmt.Errorf("could not find plaid transaction %s: %w", transactionID, err)
	}

	return label, nil
}

/**************************************************************************************
* ListPlaidTransactions returns the user's most recent transactions, newest first,
* joined with the account name/mask for display.
**************************************************************************************/
func ListPlaidTransactions(userID uint, limit int) ([]PlaidTransactionWithAccount, error) {
	query := `
		SELECT
			t.transaction_id, a.account_id, t.name, COALESCE(t.merchant_name, ''),
			t.amount, COALESCE(t.iso_currency_code, 'USD'), DATE_FORMAT(t.date, '%Y-%m-%d'),
			COALESCE(t.category, ''), t.pending, a.name, COALESCE(a.mask, '')
		FROM plaid_transactions t
		JOIN plaid_accounts a ON a.id = t.plaid_account_id
		WHERE t.user_id = ?
		ORDER BY t.date DESC, t.id DESC
		LIMIT ?
	`

	rows, err := ModelsRepo.DB.Conn.Query(query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("could not list plaid transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]PlaidTransactionWithAccount, 0)
	for rows.Next() {
		var txn PlaidTransactionWithAccount
		err := rows.Scan(
			&txn.TransactionID, &txn.AccountID, &txn.Name, &txn.MerchantName,
			&txn.Amount, &txn.IsoCurrencyCode, &txn.Date,
			&txn.Category, &txn.Pending, &txn.AccountName, &txn.AccountMask,
		)
		if err != nil {
			return nil, fmt.Errorf("could not scan plaid transaction: %w", err)
		}
		transactions = append(transactions, txn)
	}

	return transactions, nil
}

/**************************************************************************************
* ListPlaidAccounts returns every account the user has across their connected Plaid
* items, oldest connection first.
**************************************************************************************/
func ListPlaidAccounts(userID uint) ([]PlaidAccountRow, error) {
	query := `
		SELECT account_id, plaid_item_id, name, COALESCE(mask, ''), COALESCE(type, ''), COALESCE(subtype, ''),
			COALESCE(current_balance, 0), available_balance, COALESCE(iso_currency_code, 'USD')
		FROM plaid_accounts
		WHERE user_id = ?
		ORDER BY id
	`

	rows, err := ModelsRepo.DB.Conn.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("could not list plaid accounts: %w", err)
	}
	defer rows.Close()

	accounts := make([]PlaidAccountRow, 0)
	for rows.Next() {
		var account PlaidAccountRow
		err := rows.Scan(
			&account.AccountID, &account.PlaidItemID, &account.Name, &account.Mask, &account.Type,
			&account.Subtype, &account.CurrentBalance, &account.AvailableBalance, &account.IsoCurrencyCode,
		)
		if err != nil {
			return nil, fmt.Errorf("could not scan plaid account: %w", err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}
