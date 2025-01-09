package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

/**
* I map the json fields to the struct fields
 */
type PantryIngredient struct {
	Id           uint   `json:"id"`
	PantryId     uint   `json:"pantry_id"`
	IngredientId uint   `json:"ingredient_id"`
	Quantity     string `json:"quantity"`
	CreatedAt    string `json:"created_at"`
}

/**
* I map the json fields to the struct fields
 */
func (ri *PantryIngredient) MapJsonToStruct(jsonBytes []byte) error {
	err := json.Unmarshal(jsonBytes, ri)

	if err != nil {
		return fmt.Errorf("error unmarshaling struct: %w", err)
	}

	return nil
}

/**
* I validate that all the required fields are set
 */

func (ri *PantryIngredient) Validate() []map[string]interface{} {

	errors := []map[string]interface{}{}

	if !(ri.PantryId > 0) {
		errors = append(errors, map[string]interface{}{"Field": "PantryId", "Message": "Pantry items need to be associated with a recipe"})
	}

	if !(ri.IngredientId > 0) {
		errors = append(errors, map[string]interface{}{"Field": "IngredientId", "Message": "Pantry items need to be associated with a pantry"})
	}

	if strings.TrimSpace(ri.Quantity) == "" {
		errors = append(errors, map[string]interface{}{"Field": "Quantity", "Message": "Each ingredient must specify a quantity"})
	}

	return errors
}
