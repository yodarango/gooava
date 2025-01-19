package models

type DTORecipesByBatchId struct {
	RecipesBatch
	Recipes []Recipe `json:"recipes"`
}
