package models

import (
	"encoding/json"
	"fmt"
)

type Prompt struct {
	*ModelConfiguration
	CreatedAt   string `json:"created_at"`
	UserId      uint   `json:"user_id"`
	Description string `json:"description"`
	Provider    string `json:"provider"` // gemini, gpt, etc.
	Model       string `json:"model"`    // model used to create the query
	Id          uint   `json:"id"`
}

/**
* I map the json fields to the struct fields
 */
func (i *Prompt) MapJsonToStruct(jsonBytes []byte) error {

	err := json.Unmarshal(jsonBytes, i)

	if err != nil {
		return fmt.Errorf("error unmarshalig json %w", err)
	}

	return nil
}
