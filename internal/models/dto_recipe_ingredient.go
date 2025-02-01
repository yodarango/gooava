package models

import "time"

type DTORecipeIngredient struct {
	RecipeIngredient
	IngredientCreatedAt time.Time `json:"ingredient_created_at"`
	IngredientId        uint      `json:"ingredient_id"`
	TotalQuantity       int       `json:"total_qty"` // Un batch potrebbe includere un ingredient piu da una volta, questo valore e la somma totale
	Name                string    `json:"name"`
}
