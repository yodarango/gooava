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

func (rb *RecipesBatch) Save() (*RecipesBatch, error) {

	fmt.Println(rb)
	var newRecipeBatch RecipesBatch

	newRecipeBatch = RecipesBatch{
		Id: 123,
	}

	fmt.Print("....")
	return &newRecipeBatch, nil
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
