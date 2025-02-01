package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

/**
* I map the json fields to the struct fields
 */

type RecipeIngredient struct {
	Id            uint   `json:"id"`
	RecipeId      uint   `json:"recipe_id"`
	BatchId       uint   `json:"batch_id"`
	IngredientId  uint   `json:"ingredient_id"`
	Quantity      int    `json:"quantity"`
	MeasuringUnit string `json:"measuring_unit"` // necessarrio per fare la mate con qty e dopo aggiungere la untia di misura
	CreatedAt     string `json:"created_at"`
}

/**
* I map the json fields to the struct fields
 */
func (ri *RecipeIngredient) MapJsonToStruct(jsonBytes []byte) error {
	err := json.Unmarshal(jsonBytes, ri)

	if err != nil {
		return fmt.Errorf("error unmarshaling struct: %w", err)
	}

	return nil
}

/**
* I validate that all the required fields are set
 */
func (ri *RecipeIngredient) Validate() []map[string]interface{} {

	errors := []map[string]interface{}{}

	if !(ri.RecipeId > 0) {
		errors = append(errors, map[string]interface{}{"Field": "RecipeId", "Message": "Recipe Ingredients need to be associated with a recipe"})
	}

	if !(ri.IngredientId > 0) {
		errors = append(errors, map[string]interface{}{"Field": "IngredientId", "Message": "Recipe ingredients need to be associated with an ingredient"})
	}

	if !(ri.Quantity > 0) {
		errors = append(errors, map[string]interface{}{"Field": "Quantity", "Message": "Quantity cannot be zero"})
	}

	if strings.TrimSpace(ri.MeasuringUnit) == "" {
		errors = append(errors, map[string]interface{}{"Field": "MeasuringUnit", "Message": "A unit of meassurement is needed"})
	}

	return errors
}

/***************************************************************************************
 * Dammi la id per qualsiasi BatchId e i o ti daro tutti i ingredenti necessari per
 * cucinare tuttle le recette nel batch. Faro una somma per averiguare il totale
 * necessario per ogni ingrediente.
 ***************************************************************************************/
func (ri *RecipeIngredient) GetIngredientsByBatchId(id uint) ([]DTORecipeIngredient, error) {
	// imposta la query
	query := `
		SELECT 
			IFNULL(i.id, 0) AS id, 
			IFNULL(i.name, '') AS name, 
			IFNULL(ir.measuring_unit, '') as measuring_unit,
			SUM(IFNULL(ir.quantity, 0)) AS total_quantity
		FROM recipe_ingredients ir
		JOIN ingredients i ON ir.ingredient_id = i.id
		WHERE ir.batch_id = ?
		GROUP BY i.id, i.name, ir.measuring_unit
		ORDER BY total_quantity DESC;
	`

	// chiama la db
	rows, err := ModelConfig.AppRepo.DB.Connection.Query(query, id)

	if err != nil {
		return nil, fmt.Errorf("could not get ingredients for this batch: %w", err)
	}

	// imposta le righe nel struct
	var dtoRecipeIngredients []DTORecipeIngredient

	for rows.Next() {

		var dtoRecipeIngredient DTORecipeIngredient

		err := rows.Scan(
			&dtoRecipeIngredient.Id,
			&dtoRecipeIngredient.Name,
			&dtoRecipeIngredient.MeasuringUnit,
			&dtoRecipeIngredient.TotalQuantity)

		if err != nil {
			return nil, fmt.Errorf("error scaning row: %w", err)
		}

		dtoRecipeIngredients = append(dtoRecipeIngredients, dtoRecipeIngredient)
	}

	return dtoRecipeIngredients, nil
}
