package dto

type RecipeIngredientDetails struct {
	*DtoConfiguration
	Id             uint   `json:id`
	RecipeId       uint   `json:recipe_id`
	IngredientId   uint   `json:ingredient_id`
	Quantity       string `json:quantity`
	CreatedAt      string `json:created_at`
	IngredientName string `json:ingredient_name`
}

/**
* I get all the ingredients for each recipe in a batch. I add these ingredients and
* return a final sum representing the ingredients needed for the entire batch.
 */
func (rid *RecipeIngredientDetails) GetIngredientsByBatchId(id uint) []RecipeIngredientDetails {
	// get all the recipes within this batch.

	// Get the ingredients for each

	// sum them up

	// send them
	return []RecipeIngredientDetails{{

		Id:             12,
		RecipeId:       232,
		IngredientId:   12,
		Quantity:       "12lb",
		CreatedAt:      "now",
		IngredientName: "some",
	}, {

		Id:             12,
		RecipeId:       232,
		IngredientId:   12,
		Quantity:       "12lb",
		CreatedAt:      "now",
		IngredientName: "some",
	}, {

		Id:             12,
		RecipeId:       232,
		IngredientId:   12,
		Quantity:       "12lb",
		CreatedAt:      "now",
		IngredientName: "some",
	}}
}
