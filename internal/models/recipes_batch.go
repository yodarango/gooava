package models

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type RecipesBatch struct {
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
func (rb *RecipesBatch) MapFormToStruct(form url.Values) []FieldValidationError {
	var errors []FieldValidationError

	isMaximizeIngredients, err := strconv.ParseBool(form.Get("isMaximizeIngredients"))
	if err != nil {
		errors = append(errors, FieldValidationError{Field: "isMaximizeIngredients", Message: "Error parsing this value. "})
		fmt.Println(fmt.Errorf("error decoding the isMaximizeIngredients field: %w", err).Error())
	} else {
		rb.IsMaximizeIngredients = isMaximizeIngredients
	}

	isBudgetFriendly, err := strconv.ParseBool(form.Get("isBudgetFriendly"))
	if err != nil {
		errors = append(errors, FieldValidationError{Field: "isBudgetFriendly", Message: "Error parsing this value. "})
		fmt.Println(fmt.Errorf("error decoding the isBudgetFriendly field: %w", err).Error())
	} else {
		rb.IsBudgetFriendly = isBudgetFriendly
	}

	recipeCount, err := strconv.ParseUint(form.Get("recipeCount"), 10, 16)
	if err != nil {
		errors = append(errors, FieldValidationError{Field: "recipeCount", Message: "Error parsing this value. "})
		fmt.Println(fmt.Errorf("error decoding the recipeCount field: %w", err).Error())
	} else {
		rb.RecipeCount = uint16(recipeCount)
	}

	isHealthy, err := strconv.ParseBool(form.Get("isHealthy"))
	if err != nil {
		errors = append(errors, FieldValidationError{Field: "isHealthy", Message: "Error parsing this value. "})
		fmt.Println(fmt.Errorf("error decoding the isHealthy field: %w", err).Error())
	} else {
		rb.IsHealthy = isHealthy
	}

	isQuick, err := strconv.ParseBool(form.Get("isQuick"))
	if err != nil {
		errors = append(errors, FieldValidationError{Field: "isQuick", Message: "Error parsing this value. "})
		fmt.Println(fmt.Errorf("error decoding the isQuick field: %w", err).Error())
	} else {
		rb.IsQuick = isQuick
	}

	rb.CuisineType = form.Get("cuisine_type")
	rb.Name = form.Get("name")

	return errors
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

func (rb *RecipesBatch) Save() error {
	return nil
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
