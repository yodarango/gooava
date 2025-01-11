package dbinterfacerepo

// import (
// 	"fmt"

// 	"github.com/yodarango/gooava/internal/models"
// )

// func (db *DbInterfaceRepositoryConfig) CreateBatchRecipe(batchRecipe *models.RecipesBatch) (*models.RecipesBatch, error) {

// 	// Query di inserimento
// 	query := `
// 		INSERT INTO recipe_batches
// 		(user_id, name, recipe_count, is_healthy, is_quick, is_maximize_ingredients, is_budget_friendly, cuisine_type, prompt_id, created_at)
// 		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
// 	`

// 	result, err := db.AppRepo.DB.Connection.Exec(query, batchRecipe.UserId, batchRecipe.Name, batchRecipe.RecipeCount, batchRecipe.IsHealthy, batchRecipe.IsQuick,
// 		batchRecipe.IsMaximizeIngredients, batchRecipe.IsBudgetFriendly, batchRecipe.CuisineType, batchRecipe.PromptId, batchRecipe.CreatedAt)

// 	if err != nil {
// 		return nil, fmt.Errorf("failed to insert batch: %w", err)
// 	}

// 	// Recupera il ID della riga apena inserita
// 	insertedId, err := result.LastInsertId()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get last inserted ID: %w", err)
// 	}

// 	// aggiunge la ID al modelo
// 	batchRecipe.Id = uint(insertedId)

// 	return batchRecipe, nil

// }
