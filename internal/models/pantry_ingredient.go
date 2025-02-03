package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

/**
* I map the json fields to the struct fields
 */
type PantryIngredient struct {
	Id           uint   `json:"id"`
	PantryId     uint   `json:"pantry_id"`
	IngredientId uint   `json:"ingredient_id"`
	Quantity     string `json:"quantity"`
	CreatedAt    string `json:"created_at"`
}

/**********************************************************************************************************************
* This maps either a single pantryIngredient or an arary of pantryIngredient from the body and return an array if the
* body is an array and an error if fails. If the body is not an array, the data is unmarshall unto the
* pantryIngredient.
*******************/
func (pi *PantryIngredient) MapBodyToStruct(body io.ReadCloser, isMany bool) ([]PantryIngredient, error) {

	if body == nil {
		return nil, errors.New("body cannot be empty")
	}
	defer body.Close()

	//read the body
	bodyBytes, err := io.ReadAll(body)

	if err != nil {
		return nil, err
	}

	var pantryIngredients = make([]PantryIngredient, 0)

	if isMany {
		err = json.Unmarshal(bodyBytes, &pantryIngredients)
		if err != nil {
			return nil, err
		}

		return pantryIngredients, nil
	}

	err = json.Unmarshal(bodyBytes, &pi)
	if err != nil {
		return nil, err
	}

	return nil, nil

}

/**
* I validate that all the required fields are set
 */

func (ri *PantryIngredient) Validate() []map[string]interface{} {

	errors := []map[string]interface{}{}

	if !(ri.PantryId > 0) {
		errors = append(errors, map[string]interface{}{"Field": "PantryId", "Message": "Pantry items need to be associated with a recipe"})
	}

	if !(ri.IngredientId > 0) {
		errors = append(errors, map[string]interface{}{"Field": "IngredientId", "Message": "Pantry items need to be associated with a pantry"})
	}

	if strings.TrimSpace(ri.Quantity) == "" {
		errors = append(errors, map[string]interface{}{"Field": "Quantity", "Message": "Each ingredient must specify a quantity"})
	}

	return errors
}

/***********************************************************************
* Takes the different type of ingredients sent by the client and saves
* them all to the pantry of the logged in user
**********************************************************************/
func (ri *PantryIngredient) SaveMany(pantryIngredients []PantryIngredient) error {
	query := `
	INSERT INTO pantry_ingredients (pantry_id, ingredient_id, quantity)
	VALUES
	`
	queryPlaceHolders := make([]string, 0)
	queryValues := make([]interface{}, 0)

	for _, pantrIngredient := range pantryIngredients {
		queryPlaceHolders = append(queryPlaceHolders, `(?, ?, ?)`)
		queryValues = append(queryValues, pantrIngredient.IngredientId, pantrIngredient.IngredientId, pantrIngredient.Quantity)
	}

	query = strings.Join(queryPlaceHolders, "\n")

	result, err := ModelConfig.AppRepo.DB.Connection.Exec(query, queryValues...)

	if err != nil {
		return fmt.Errorf("error insertings pantry ingredients %w", err)
	}

	affectedRows, _ := result.RowsAffected()

	if !(affectedRows > 0) {
		return fmt.Errorf("no ingredient rows were inserted")
	}

	return nil
}
