package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type RecipesBatch struct {
	*ModelConfiguration
	Id                    uint   `json:id`
	UserId                uint   `json:user_id`
	Name                  string `json:name`
	RecipeCount           uint16 `json:recipe_count`
	IsHealthy             bool   `json:is_healthy`
	IsQuick               bool   `json:is_quick`
	IsMaximizeIngredients bool   `json:is_maximize_ingredients`
	IsBudgetFriendly      bool   `json:is_budget_friendly`
	CuisineType           string `json:cuisine_type`
	PromptId              uint   `json:prompt_id`
	CreatedAt             string `json:created_at`
}

/**
 * I map the form fields to the struct fields
 */
func (rb *RecipesBatch) MapBodyToStruct(body io.ReadCloser) error {

	if body == nil {
		return errors.New("Body cannot be empty")
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
func (rb *RecipesBatch) Validate() []FieldValidationError {

	errors := make([]FieldValidationError, 0, 5)

	if !(rb.UserId > 0) {
		errors = append(errors, FieldValidationError{Field: "UserId", Message: "User Id cannot be empty"})
	}

	if strings.TrimSpace(rb.Name) == "" {
		errors = append(errors, FieldValidationError{Field: "Name", Message: "You need to provide a name for this batch"})
	}

	if !(rb.RecipeCount > 0) {
		errors = append(errors, FieldValidationError{Field: "RecipeCount", Message: "Please specify how many recipes this batch is for"})
	}

	if !(rb.PromptId > 0) {
		errors = append(errors, FieldValidationError{Field: "PromptId", Message: "The prompt used to created this batch must be specified"})
	}

	return errors
}

func (rb *RecipesBatch) Save() (*RecipesBatch, error) {
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
