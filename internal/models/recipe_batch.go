package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type RecipesBatch struct {
	Id                    uint   `json:"id"`
	UserId                uint   `json:"user_id"`
	Name                  string `json:"name"`
	RecipeCount           uint16 `json:"recipe_count"`
	IsHealthy             bool   `json:"is_healthy"`
	IsQuick               bool   `json:"is_quick"`
	IsMaximizeIngredients bool   `json:"is_maximize_ingredients"`
	IsBudgetFriendly      bool   `json:"is_budget_friendly"`
	CuisineType           string `json:"cuisine_type"`
	PromptId              uint   `json:"prompt_id"`
	CreatedAt             string `json:"created_at"`
}

/**
 * I map the form fields to the struct fields
 */
func (rb *RecipesBatch) MapBodyToStruct(body io.ReadCloser) error {

	if body == nil {
		return errors.New("body cannot be empty")
	}
	defer body.Close()

	//read the body
	bodyBytes, err := io.ReadAll(body)

	if err != nil {
		return err
	}

	err = json.Unmarshal(bodyBytes, &rb)
	if err != nil {
		return err
	}

	return nil
}

/**
 * I validate that all the required fields are set
 */
func (rb *RecipesBatch) Validate() []map[string]interface{} {

	errors := []map[string]interface{}{}

	// TODO! :Delete once auth is set
	rb.UserId = 123
	rb.PromptId = 123

	if !(rb.UserId > 0) {
		errors = append(errors, map[string]interface{}{"field": "user_id", "message": "User Id cannot be empty"})
	}

	if !(rb.RecipeCount > 0) {
		errors = append(errors, map[string]interface{}{"field": "recipe_count", "message": "Please specify how many recipes this batch is for"})
	}

	if !(rb.PromptId > 0) {
		errors = append(errors, map[string]interface{}{"field": "prompt_id", "message": "The prompt used to created this batch must be specified"})
	}

	return errors
}

/**
 * I insert the current instance of RecipesBatch into the DB
 */
func (rb *RecipesBatch) Save() (*RecipesBatch, error) {
	// Query di inserimento
	query := `
		INSERT INTO recipe_batches 
		(user_id, name, recipe_count, is_healthy, is_quick, is_maximize_ingredients, is_budget_friendly, cuisine_type, prompt_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// esegui la query
	result, err := ModelConfig.AppRepo.DB.Connection.Exec(query, rb.UserId, rb.Name, rb.RecipeCount, rb.IsHealthy, rb.IsQuick,
		rb.IsMaximizeIngredients, rb.IsBudgetFriendly, rb.CuisineType, rb.PromptId)

	if err != nil {
		return nil, fmt.Errorf("failed to insert batch: %w", err)
	}

	// Recupera il ID della riga apena inserita
	insertedId, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last inserted ID: %w", err)
	}

	fmt.Println("==============", !(uint(insertedId) > 0))

	// there is no Id, la riga non fu inserita
	if !(uint(insertedId) > 0) {
		return nil, fmt.Errorf("there was no errors found but the record does not appear to have been inserted")
	}

	// aggiunge la ID al modelo
	rb.Id = uint(insertedId)

	return rb, nil
}

func (rb *RecipesBatch) Update() {

}

func (rb *RecipesBatch) Delete() {

}

/**
* I get all the batches for the logged in user
 */
func (rb *RecipesBatch) GetAll() []RecipesBatch {
	tempResults := []RecipesBatch{{
		Id:                    1,
		UserId:                2,
		Name:                  "name",
		RecipeCount:           4,
		IsHealthy:             false,
		IsQuick:               false,
		IsMaximizeIngredients: false,
		IsBudgetFriendly:      false,
		CuisineType:           "italian",
		PromptId:              1,
		CreatedAt:             "time",
	}, {
		Id:                    1,
		UserId:                2,
		Name:                  "name",
		RecipeCount:           4,
		IsHealthy:             false,
		IsQuick:               false,
		IsMaximizeIngredients: false,
		IsBudgetFriendly:      false,
		CuisineType:           "italian",
		PromptId:              1,
		CreatedAt:             "time",
	}}

	return tempResults
}

/**
* I get one batch by id without its belonging recipes
 */
func (rb *RecipesBatch) GetOneById(id uint) (*RecipesBatch, error) {
	query := `SELECT * FROM recipe_batches WHERE id = ?`
	row := ModelConfig.AppRepo.DB.Connection.QueryRow(query, id)

	err := row.Scan(rb)

	if err != nil {
		return nil, err
	}

	return rb, nil
}

/**
* I get one batch by id along with all its associated recipes
 */
func (rb *RecipesBatch) GetBatchIngredientsById(id uint) (*RecipesByBatchId, error) {
	// Query con alias per evitare conflitti tra colonne
	query := `
	SELECT 
	    rb.id AS rb_id, rb.user_id AS rb_user_id, rb.name AS rb_name, rb.recipe_count, rb.is_healthy AS rb_is_healthy, 
	    rb.is_quick AS rb_is_quick, rb.is_maximize_ingredients AS rb_is_maximize_ingredients, rb.is_budget_friendly AS rb_is_budget_friendly, 
	    rb.cuisine_type AS rb_cuisine_type, rb.prompt_id, rb.created_at AS rb_created_at,
	    r.id AS r_id, r.user_id AS r_user_id, r.name AS r_name, r.is_healthy AS r_is_healthy, r.is_quick AS r_is_quick, 
	    r.is_maximize_ingredients AS r_is_maximize_ingredients, r.is_budget_friendly AS r_is_budget_friendly, 
	    r.cuisine_type AS r_cuisine_type, r.created_at AS r_created_at, r.batch_id, r.servings, r.instructions
	FROM recipe_batches AS rb
	JOIN recipes AS r ON r.batch_id = rb.id
	WHERE rb.id = ?`

	rows, err := ModelConfig.AppRepo.DB.Connection.Query(query, id)

	if err != nil {
		return nil, fmt.Errorf("error querying for batch recipes %w", err)

	}

	// get the DTO
	var recipeByBatchId RecipesByBatchId
	var recipes []Recipe

	// this will match the properties of the DTO to the DTO and those of the recipe to the recipe
	for rows.Next() {
		var recipe Recipe

		err := rows.Scan(
			&recipeByBatchId.Id, &recipeByBatchId.UserId, &recipeByBatchId.Name, &recipeByBatchId.RecipeCount, &recipeByBatchId.IsHealthy,
			&recipeByBatchId.IsQuick, &recipeByBatchId.IsMaximizeIngredients, &recipeByBatchId.IsBudgetFriendly, &recipeByBatchId.CuisineType,
			&recipeByBatchId.PromptId, &recipeByBatchId.CreatedAt,
			&recipe.Id, &recipe.UserId, &recipe.Name, &recipe.IsHealthy, &recipe.IsQuick,
			&recipe.IsMaximizeIngredients, &recipe.IsBudgetFriendly, &recipe.CuisineType,
			&recipe.CreatedAt, &recipe.BatchId, &recipe.Servings, &recipe.Instructions,
		)

		if err != nil {
			return nil, fmt.Errorf("could not scan row %w", err)
		}

		recipes = append(recipes, recipe)
	}

	recipeByBatchId.Recipes = recipes

	return &recipeByBatchId, nil
}
