package models

import (
	"fmt"
)

// Setting keys in use (the user_settings table is key/value per user).
const SETTING_KEY_PAYCHECK_WINDOWS = "paycheck_windows"

/**************************************************************************************
* ListUserSettings returns every saved setting of the user as a key -> value map
* (missing keys are simply absent — callers apply their own defaults).
*
* status: ✅
**************************************************************************************/
func ListUserSettings(userID uint) (map[string]string, error) {
	rows, err := ModelsRepo.DB.Conn.Query(
		"SELECT setting_key, setting_value FROM user_settings WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("could not list user settings: %w", err)
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("could not scan user setting: %w", err)
		}
		settings[key] = value
	}

	return settings, nil
}

/**************************************************************************************
* UpsertUserSetting saves one setting for the user (one row per user+key).
*
* status: ✅
**************************************************************************************/
func UpsertUserSetting(userID uint, key, value string) error {
	_, err := ModelsRepo.DB.Conn.Exec(
		`INSERT INTO user_settings(user_id, setting_key, setting_value) VALUES(?,?,?)
		 ON DUPLICATE KEY UPDATE setting_value = VALUES(setting_value)`,
		userID, key, value,
	)
	if err != nil {
		return fmt.Errorf("could not save user setting: %w", err)
	}

	return nil
}
