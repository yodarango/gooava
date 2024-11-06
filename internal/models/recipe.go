package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Recipe struct {
	Id                    uint   `json:id`
	UserId                uint   `json:user_id`
	Name                  string `json:name`
	IsHealthy             bool   `json:is_healthy`
	IsQuick               bool   `json:is_quick`
	IsMaximizeIngredients bool   `json:is_maximize_ingredients`
	IsBudgetFriendly      bool   `json:is_budget_friendly`
	CuisineType           string `json:cuisine_type`
	CreatedAt             string `json:created_at`
	BatchId               uint   `json:batch_id`
	Servings              uint16 `json:batch_id`
	Instructions          string `json:instructions`
}

/**
* I map the json fields to the struct fields
 */
func (r *Recipe) MapJsonToStruct(jsonBytes []byte) error {
	err := json.Unmarshal(jsonBytes, r)
	if err != nil {
		return fmt.Errorf("unable to unmarshall json %w", err)
	}

	return nil
}

/**
* I validate that all the required fields are set
 */
func (r *Recipe) Validate() []FieldValidationError {
	errors := make([]FieldValidationError, 0, 6)

	if !(r.UserId > 0) {
		errors = append(errors, FieldValidationError{Field: "UserId", Message: "Please provide the user Id associated with this recipe"})
	}

	if strings.TrimSpace(r.Name) == "" {
		errors = append(errors, FieldValidationError{Field: "Name", Message: "Recipes must have a name"})
	}

	if !(r.Servings > 0) {
		errors = append(errors, FieldValidationError{Field: "Servings", Message: "Servings count is necessary"})
	}

	if strings.TrimSpace(r.Instructions) == "" {
		errors = append(errors, FieldValidationError{Field: "Instructions", Message: "Users need to know how to prepare this recipe"})
	}

	return errors
}

func (rb *Recipe) Save() {

}

func (rb *Recipe) Update() {

}

func (rb *Recipe) Delete() {

}
