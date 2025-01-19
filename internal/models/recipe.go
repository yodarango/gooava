package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

type Recipe struct {
	Id                    uint   `json:"id"`
	UserId                uint   `json:"user_id"`
	Name                  string `json:"name"`
	IsHealthy             bool   `json:"is_healthy"`
	IsQuick               bool   `json:"is_quick"`
	IsMaximizeIngredients bool   `json:"is_maximize_ingredients"`
	IsBudgetFriendly      bool   `json:"is_budget_friendly"`
	CuisineType           string `json:"cuisine_type"`
	CreatedAt             string `json:"created_at"`
	BatchId               uint   `json:"batch_id"`
	Servings              uint16 `json:"servings"`
	Instructions          string `json:"instructions"`
}

/***************************************************************************************
 * Io imposto i campi da json a il struct
 ***************************************************************************************/
func (r *Recipe) MapJsonToStruct(jsonBytes []byte) error {
	err := json.Unmarshal(jsonBytes, r)
	if err != nil {
		return fmt.Errorf("unable to unmarshall json %w", err)
	}

	return nil
}

/***************************************************************************************
 * Io mi asicuro che tutti i campi necessari siano impost
 ***************************************************************************************/
func (r *Recipe) Validate() []map[string]interface{} {
	errors := []map[string]interface{}{}

	if !(r.UserId > 0) {
		errors = append(errors, map[string]interface{}{"Field": "UserId", "Message": "Please provide the user Id associated with this recipe"})
	}

	if strings.TrimSpace(r.Name) == "" {
		errors = append(errors, map[string]interface{}{"Field": "Name", "Message": "Recipes must have a name"})
	}

	if !(r.Servings > 0) {
		errors = append(errors, map[string]interface{}{"Field": "Servings", "Message": "Servings count is necessary"})
	}

	if strings.TrimSpace(r.Instructions) == "" {
		errors = append(errors, map[string]interface{}{"Field": "Instructions", "Message": "Users need to know how to prepare this recipe"})
	}

	return errors
}

/***************************************************************************************
 * Fornirò il nome e la id di ogni recetta per la id del RecipesBatch specificato. Dato
 * che il Batch già ha tutti i dati necesari non c'e bisogno di fornice tutte questi
 * per ogni recetta. Neanche c'e bisogno di fornire altra informazione, dao che
 * il usante e da solito chiedendo solo una lista delle reccette disponibile.
 ***************************************************************************************/
func (r *Recipe) GetRecipesByBatchId(id uint) ([]Recipe, error) {
	// Query con alias per evitare conflitti tra colonne
	query := `
		SELECT 
			IFNULL(id, 0), 
			IFNULL(name, '')
		FROM recipes
		WHERE batch_id = ?`

	rows, err := ModelConfig.AppRepo.DB.Connection.Query(query, id)

	if err != nil {
		return nil, fmt.Errorf("error querying for batch recipes %w", err)

	}

	// get the DTO
	var recipes []Recipe

	// this will match the properties of the DTO to the DTO and those of the recipe to the recipe
	for rows.Next() {
		var recipe Recipe

		err := rows.Scan(
			&recipe.Id, &recipe.Name,
		)

		if err != nil {
			return nil, fmt.Errorf("could not scan row %w", err)
		}

		recipes = append(recipes, recipe)
	}

	return recipes, nil
}

/***************************************************************************************
 * Io ritornero tuttal la informazione per una recetta. Se il usante sta rechiedendo una
 * recetta per id, loro vogliono sapare come cuccinarla da solito.
 ***************************************************************************************/
func (r *Recipe) GetRecipeById(id uint) (*Recipe, error) {

	// imposta la query
	query := `
		SELECT 
			IFNULL(id, 0), 
			IFNULL(user_id, 0),
			IFNULL(name, ''),
			IFNULL(is_healthy, false), 
			IFNULL(is_quick, false), 
			IFNULL(is_maximize_ingredients, false), 
			IFNULL(is_budget_friendly, false), 
			IFNULL(cuisine_type, ''), 
			IFNULL(created_at, ''), 
			IFNULL(batch_id, 0), 
			IFNULL(servings, 0), 
			IFNULL(instructions, '')
		FROM recipes
		WHERE batch_id = ?`

	rows := ModelConfig.AppRepo.DB.Connection.QueryRow(query, id)

	if rows.Err() != nil && rows.Err() != sql.ErrNoRows {
		return nil, fmt.Errorf("error querying for recipe with id of %d %w", id, rows.Err())

	}

	var recipe Recipe

	err := rows.Scan(
		&recipe.Id, &recipe.UserId, &recipe.Name, &recipe.IsHealthy, &recipe.IsQuick,
		&recipe.IsMaximizeIngredients, &recipe.IsBudgetFriendly, &recipe.CuisineType,
		&recipe.CreatedAt, &recipe.BatchId, &recipe.Servings, &recipe.Instructions,
	)

	if err != nil {
		return nil, fmt.Errorf("could not scan row %w", err)
	}

	return &recipe, nil
}

/**
* I get all the recipes
 */
func (r *Recipe) GetAll() []Recipe {
	return []Recipe{
		{
			Id:                    212,
			UserId:                12,
			Name:                  "name",
			IsHealthy:             true,
			IsQuick:               false,
			IsMaximizeIngredients: false,
			IsBudgetFriendly:      true,
			CuisineType:           "italian",
			CreatedAt:             "12-12-63",
			BatchId:               123,
			Servings:              3,
			Instructions:          "This are the instructions",
		},
		{
			Id:                    212,
			UserId:                12,
			Name:                  "name",
			IsHealthy:             true,
			IsQuick:               false,
			IsMaximizeIngredients: false,
			IsBudgetFriendly:      true,
			CuisineType:           "italian",
			CreatedAt:             "12-12-63",
			BatchId:               123,
			Servings:              3,
			Instructions:          "This are the instructions",
		},
	}
}

func (r *Recipe) Save() (*Recipe, error) {
	// set the query
	query := `
		INSERT INTO recipes (
			user_id, batch_id, name, is_healthy, is_quick, 
			is_maximize_ingredients, is_budget_friendly, cuisine_type, 
			servings, instructions
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 `

	// save the new recipe
	result, err := ModelConfig.AppRepo.DB.Connection.Exec(query,
		r.UserId,
		r.BatchId,
		r.Name,
		r.IsHealthy,
		r.IsQuick,
		r.IsMaximizeIngredients,
		r.IsBudgetFriendly,
		r.CuisineType,
		r.Servings,
		r.Instructions)

	if err != nil {
		return nil, fmt.Errorf("could not insert result into table %w", err)
	}

	lastInsertedId, err := result.LastInsertId()

	if err != nil {
		return nil, fmt.Errorf("failed to get last inserted ID: %w", err)
	}

	// there is no Id, la riga non fu inserita
	if !(uint(lastInsertedId) > 0) {
		return nil, fmt.Errorf("there was no errors found but the record does not appear to have been inserted")
	}

	r.Id = uint(lastInsertedId)

	return r, nil
}

func (r *Recipe) SaveMany(recipes []Recipe) (int64, error) {
	// set the query
	query := `
		INSERT INTO recipes (
			user_id, batch_id, name, is_healthy, is_quick, 
			is_maximize_ingredients, is_budget_friendly, cuisine_type, 
			servings, instructions
		) VALUES 
		 `
	queryValues := []interface{}{}
	queryValuePlaceholders := []string{}

	for _, recipe := range recipes {
		queryValuePlaceholders = append(queryValuePlaceholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

		queryValues = append(queryValues,
			recipe.UserId,
			recipe.BatchId,
			recipe.Name,
			recipe.IsHealthy,
			recipe.IsQuick,
			recipe.IsMaximizeIngredients,
			recipe.IsBudgetFriendly,
			recipe.CuisineType,
			recipe.Servings,
			recipe.Instructions)
	}

	unifiedQueryStrs := strings.Join(queryValuePlaceholders, ", ")
	query += unifiedQueryStrs

	// save the new recipe
	result, err := ModelConfig.AppRepo.DB.Connection.Exec(query, queryValues...)

	if err != nil {
		return 0, fmt.Errorf("could not insert result into table %w", err)
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return 0, fmt.Errorf("failed to get last inserted ID: %w", err)
	}

	// there is no Id, la riga non fu inserita
	if !(rowsAffected > 0) {
		return 0, fmt.Errorf("there was no errors found but the record does not appear to have been inserted")
	}

	return rowsAffected, nil
}
func (r *Recipe) Update() {

}

func (r *Recipe) Delete() {

}
