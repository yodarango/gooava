package models

import "fmt"

// LEFT OFF.  I need to ge the units available to the html select list
type IngredientUnit struct {
	Id           uint   `json:"id"`
	Name         string `json:"name"`
	Category     string `json:"category"`
	Abbreviation string `json:"abbreviation"`
}

func (iu *IngredientUnit) GetAll() ([]IngredientUnit, error) {

	query := `SELECT id, name, category, abbreviation FROM ingredient_units ORDER BY name ASC;`

	row, err := ModelConfig.AppRepo.DB.Connection.Query(query)

	if err != nil {
		return nil, fmt.Errorf("could not query for ingredinet units: %w", err)
	}

	ingredientUnits := make([]IngredientUnit, 0)

	for row.Next() {
		var ingredientUnit IngredientUnit

		err := row.Scan(&ingredientUnit.Id, &ingredientUnit.Name, &ingredientUnit.Category, &ingredientUnit.Abbreviation)

		if err != nil {
			return nil, fmt.Errorf("could scan row: %w", err)
		}

		ingredientUnits = append(ingredientUnits, ingredientUnit)
	}

	return ingredientUnits, err

}
