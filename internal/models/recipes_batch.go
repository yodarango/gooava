package models

import (
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
func (rb *RecipesBatch) MapFormToStruct(form url.Values) {
	isMaximizeIngredients, err := strconv.ParseBool(form.Get("isMaximizeIngredients"))
	if err != nil {
		rb.IsMaximizeIngredients = isMaximizeIngredients
	}

	isBudgetFriendly, err := strconv.ParseBool(form.Get("isBudgetFriendly"))
	if err != nil {
		rb.IsBudgetFriendly = isBudgetFriendly
	}

	recipeCount, err := strconv.ParseUint(form.Get("recipeCount"), 10, 16)
	if err != nil {
		rb.RecipeCount = uint16(recipeCount)
	}

	isHealthy, err := strconv.ParseBool(form.Get("isHealthy"))
	if err != nil {
		rb.IsHealthy = isHealthy
	}

	isQuick, err := strconv.ParseBool(form.Get("isHealthy"))
	if err != nil {
		rb.IsQuick = isQuick
	}

	rb.CuisineType = form.Get("cuisine_type")
	rb.Name = form.Get("name")
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

func (rb *RecipesBatch) Save() {

}

func (rb *RecipesBatch) Update() {

}

func (rb *RecipesBatch) Delete() {

}
