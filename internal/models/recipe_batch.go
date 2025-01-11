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
// LEFT OFF HERE. TRY TO UNDERSTAND THE DEPENDENCY INJECTION WHOLE THING BEFORE MOVING ONE. REDIRECT AFTER THE ID HAS BEEN INSERTED
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
* I get one batch by id for the logged in user
 */
func (rb *RecipesBatch) GetOneById(id uint) RecipesBatch {
	tempResults := RecipesBatch{
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
	}

	return tempResults
}

func (rb *RecipesBatch) GetBatchIngredientsById(id uint) {

}
