package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Recipe struct {
	*ModelConfiguration
	Id                    uint   `json:"id"`
	UserId                uint   `json:"user_id"`
	Name                  string `json:"name"`
	IsHealthy             bool   `json:"is_healthy"`
	IsQuick               bool   `json:"is_quick"`
	IsMaximizeIngredients bool   `json:"is_maximize_ingredients"`
	IsBudgetFriendly      bool   `json:"is_budget_friendly"`
	CuisineType           string `json:"cuisine_type"`
	CreatedAt             string `json:"created_at"`
	BatchId               uint   `json:"batch_id"`
	Servings              uint16 `json:"servings"`
	Instructions          string `json:"instructions"`
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
func (r *Recipe) Validate() []map[string]interface{} {
	errors := []map[string]interface{}{}

	if !(r.UserId > 0) {
		errors = append(errors, map[string]interface{}{"Field": "UserId", "Message": "Please provide the user Id associated with this recipe"})
	}

	if strings.TrimSpace(r.Name) == "" {
		errors = append(errors, map[string]interface{}{"Field": "Name", "Message": "Recipes must have a name"})
	}

	if !(r.Servings > 0) {
		errors = append(errors, map[string]interface{}{"Field": "Servings", "Message": "Servings count is necessary"})
	}

	if strings.TrimSpace(r.Instructions) == "" {
		errors = append(errors, map[string]interface{}{"Field": "Instructions", "Message": "Users need to know how to prepare this recipe"})
	}

	return errors
}

/**
* I get all the recipes for a batch id
 */
func (r *Recipe) GetRecipesByBatchId(id uint) []Recipe {
	return []Recipe{
		{
			Id:                    212,
			UserId:                12,
			Name:                  "name",
			IsHealthy:             true,
			IsQuick:               false,
			IsMaximizeIngredients: false,
			IsBudgetFriendly:      true,
			CuisineType:           "italian",
			CreatedAt:             "12-12-63",
			BatchId:               123,
			Servings:              3,
			Instructions:          "This are the instructions",
		},
		{
			Id:                    212,
			UserId:                12,
			Name:                  "name",
			IsHealthy:             true,
			IsQuick:               false,
			IsMaximizeIngredients: false,
			IsBudgetFriendly:      true,
			CuisineType:           "italian",
			CreatedAt:             "12-12-63",
			BatchId:               123,
			Servings:              3,
			Instructions:          "This are the instructions",
		},
	}
}

/**
* I get all the recipes
 */
func (r *Recipe) GetAll() []Recipe {
	return []Recipe{
		{
			Id:                    212,
			UserId:                12,
			Name:                  "name",
			IsHealthy:             true,
			IsQuick:               false,
			IsMaximizeIngredients: false,
			IsBudgetFriendly:      true,
			CuisineType:           "italian",
			CreatedAt:             "12-12-63",
			BatchId:               123,
			Servings:              3,
			Instructions:          "This are the instructions",
		},
		{
			Id:                    212,
			UserId:                12,
			Name:                  "name",
			IsHealthy:             true,
			IsQuick:               false,
			IsMaximizeIngredients: false,
			IsBudgetFriendly:      true,
			CuisineType:           "italian",
			CreatedAt:             "12-12-63",
			BatchId:               123,
			Servings:              3,
			Instructions:          "This are the instructions",
		},
	}
}

func (r *Recipe) Save() {

}

func (r *Recipe) Update() {

}

func (r *Recipe) Delete() {

}
