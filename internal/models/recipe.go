package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Recipe struct {
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

func (r *Recipe) Save() (*Recipe, error) {
	// set the query
	query := `
		INSERT INTO recipes (
			user_id, batch_id, name, is_healthy, is_quick, 
			is_maximize_ingredients, is_budget_friendly, cuisine_type, 
			servings, instructions
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 `

	// save the new recipe
	result, err := ModelConfig.AppRepo.DB.Connection.Exec(query,
		r.UserId,
		r.BatchId,
		r.Name,
		r.IsHealthy,
		r.IsQuick,
		r.IsMaximizeIngredients,
		r.IsBudgetFriendly,
		r.CuisineType,
		r.Servings,
		r.Instructions)

	if err != nil {
		return nil, fmt.Errorf("could not insert result into table %w", err)
	}

	lastInsertedId, err := result.LastInsertId()

	if err != nil {
		return nil, fmt.Errorf("failed to get last inserted ID: %w", err)
	}

	// there is no Id, la riga non fu inserita
	if !(uint(lastInsertedId) > 0) {
		return nil, fmt.Errorf("there was no errors found but the record does not appear to have been inserted")
	}

	r.Id = uint(lastInsertedId)

	return r, nil
}

func (r *Recipe) SaveMany(recipes []Recipe) (int64, error) {
	// set the query
	query := `
		INSERT INTO recipes (
			user_id, batch_id, name, is_healthy, is_quick, 
			is_maximize_ingredients, is_budget_friendly, cuisine_type, 
			servings, instructions
		) VALUES 
		 `
	queryValues := []interface{}{}
	queryValuePlaceholders := []string{}

	for _, recipe := range recipes {
		queryValuePlaceholders = append(queryValuePlaceholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

		queryValues = append(queryValues,
			recipe.UserId,
			recipe.BatchId,
			recipe.Name,
			recipe.IsHealthy,
			recipe.IsQuick,
			recipe.IsMaximizeIngredients,
			recipe.IsBudgetFriendly,
			recipe.CuisineType,
			recipe.Servings,
			recipe.Instructions)
	}

	unifiedQueryStrs := strings.Join(queryValuePlaceholders, ", ")
	query += unifiedQueryStrs

	// save the new recipe
	result, err := ModelConfig.AppRepo.DB.Connection.Exec(query, queryValues...)

	if err != nil {
		return 0, fmt.Errorf("could not insert result into table %w", err)
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return 0, fmt.Errorf("failed to get last inserted ID: %w", err)
	}

	// there is no Id, la riga non fu inserita
	if !(rowsAffected > 0) {
		return 0, fmt.Errorf("there was no errors found but the record does not appear to have been inserted")
	}

	return rowsAffected, nil
}
func (r *Recipe) Update() {

}

func (r *Recipe) Delete() {

}
