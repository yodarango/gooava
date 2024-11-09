package models

import (
	"encoding/json"
	"fmt"
)

type Ingredient struct {
	*ModelConfiguration
	CreatedAt string `json:created_at`
	Name      string `json:name`
	Id        uint   `json:id`
}

/**
* I map the json fields to the struct fields
 */
func (i *Ingredient) MapJsonToStruct(jsonBytes []byte) error {

	err := json.Unmarshal(jsonBytes, i)

	if err != nil {
		return fmt.Errorf("error unmarshalig json %w", err)
	}

	return nil
}
