package models

/**
* Ingredients for each recipe in a batch. I add these ingredients and
* return a final sum representing the ingredients needed for the entire batch.
 */
type RecipeIngredientDetails struct {
	Id             uint   `json:"id"`
	RecipeId       uint   `json:"recipe_id"`
	IngredientId   uint   `json:"ingredient_id"`
	Quantity       string `json:"quantity"`
	CreatedAt      string `json:"created_at"`
	IngredientName string `json:"ingredient_name"`
}
