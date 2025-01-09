package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

/**
* I map the json fields to the struct fields
 */

type RecipeIngredient struct {
	Id           uint   `json:"id"`
	RecipeId     uint   `json:"recipe_id"`
	IngredientId uint   `json:"ingredient_id"`
	Quantity     string `json:"quantity"`
	CreatedAt    string `json:"created_at"`
}

/**
* I map the json fields to the struct fields
 */
func (ri *RecipeIngredient) MapJsonToStruct(jsonBytes []byte) error {
	err := json.Unmarshal(jsonBytes, ri)

	if err != nil {
		return fmt.Errorf("error unmarshaling struct: %w", err)
	}

	return nil
}

/**
* I validate that all the required fields are set
 */
func (ri *RecipeIngredient) Validate() []map[string]interface{} {

	errors := []map[string]interface{}{}

	if !(ri.RecipeId > 0) {
		errors = append(errors, map[string]interface{}{"Field": "RecipeId", "Message": "Recipe Ingredients need to be associated with a recipe"})
	}

	if !(ri.IngredientId > 0) {
		errors = append(errors, map[string]interface{}{"Field": "IngredientId", "Message": "Recipe ingredients need to be associated with an ingredient"})
	}

	if strings.TrimSpace(ri.Quantity) == "" {
		errors = append(errors, map[string]interface{}{"Field": "Quantity", "Message": "Each ingredient must specify a quantity"})
	}

	return errors
}
