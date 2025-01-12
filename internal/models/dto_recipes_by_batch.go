package models

type RecipesByBatchId struct {
	RecipesBatch
	Recipes []Recipe `json:"recipes"`
}
