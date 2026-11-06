package models

import (
	"database/sql"
	"fmt"
)

/**************************************************************************************
* GetBankBalance returns the user's manually entered "in the bank" balance, or nil
* when none has been saved yet (the frontend shows its "set amount" prompt then).
*
* status: ✅
**************************************************************************************/
func GetBankBalance(userID uint) (*float64, error) {
	var balance float64

	err := ModelsRepo.DB.Conn.QueryRow(
		"SELECT balance FROM bank_balances WHERE user_id = ?",
		userID,
	).Scan(&balance)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("could not get bank balance: %w", err)
	}

	return &balance, nil
}

/**************************************************************************************
* UpsertBankBalance saves the user's manually entered balance (one row per user).
*
* status: ✅
**************************************************************************************/
func UpsertBankBalance(userID uint, balance float64) error {
	_, err := ModelsRepo.DB.Conn.Exec(
		`INSERT INTO bank_balances(user_id, balance) VALUES(?,?)
		 ON DUPLICATE KEY UPDATE balance = VALUES(balance)`,
		userID, balance,
	)
	if err != nil {
		return fmt.Errorf("could not save bank balance: %w", err)
	}

	return nil
}
