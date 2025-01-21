package apiv1

import (
	"fmt"
	"log"
	"net/http"

	"github.com/yodarango/gooava/internal/models"
	"github.com/yodarango/gooava/internal/utils"
)

// I get all the batches available for a specific user
func (c *ApiConfiguration) GetBathes(w http.ResponseWriter, r *http.Request) {

	var batches models.RecipesBatch

	templateRenderer := utils.TemplateRenderer{
		Title: "Batches",
		Name:  "batches",
		Data:  batches.GetAll(),
	}

	err := templateRenderer.Render(w)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing template %v", err), http.StatusInternalServerError)
		log.Printf("Error parsing template %v", err)
		return
	}
}

/***************************************************************************************
* Ottiene il totale di ingredienti necessarie per cucinare tuttle recette nel batchId
***************************************************************************************/
func (c *ApiConfiguration) GetSingleBatchIngredients(w http.ResponseWriter, r *http.Request, id uint) {
	var template utils.TemplateRenderer
	var recipeIngredients models.RecipeIngredient
	var responseError models.ResponseError

	// imposta le propieta dil template per default
	var templateUIMeta utils.TemplateUIMEta
	templateUIMeta.MenuIcon = "restaurant-outline"
	template.UiMeta = &templateUIMeta
	template.Title = fmt.Sprintf("Recipe %d", id)
	template.Name = "batches_id_ingredients"

	// coglie la recetta per id
	data, err := recipeIngredients.GetIngredientsByBatchId(id)

	if err != nil {
		log.Println(err)
		responseError.Title = "Could not find this Id"
		responseError.Code = "Internal Erro"
		responseError.Error = "I did not find any information pertaining to that Id"
		template.Error = &responseError
		template.Data = map[string]interface{}{}

		template.Render(w)
		return
	}

	template.Data = data
	template.Title = "Total ingredients needed for this batch"
	template.Error = nil

	err = template.Render(w)

	if err != nil {
		log.Println(err)
		errorMsg := fmt.Sprintf("%v", err)
		http.Error(w, errorMsg, http.StatusInternalServerError)
		return
	}

}

/***************************************************************************************
* Ottenere un batch tramite ID restituirà l'elenco di tutte le ricette incluse nel batch.
* Verranno fornite solo le informazioni di base su ciascuna ricetta; tutte le altre
* informazioni saranno disponibili solo se l'utente decide di visitare una specifica
* ricetta tramite ID, il che lo porterà a una pagina diversa.
***************************************************************************************/
func (c *ApiConfiguration) GetBatchById(w http.ResponseWriter, r *http.Request, id uint) {

	var dtoRecipesByBatchId models.DTORecipesByBatchId
	var responseError models.ResponseError
	var template utils.TemplateRenderer
	var recipeBatch models.RecipesBatch
	var recipe models.Recipe

	// set the defaults of the template. Regardless of what the outcome is, these will
	// remain the same
	template.UiMeta = &utils.TemplateUIMEta{}
	template.UiMeta.MenuIcon = "restaurant-outline"
	template.Name = "batches_id"
	template.Title = "Batch Recipes"

	batch, err := recipeBatch.GetBatchById(uint(id))

	if err != nil {
		log.Println(err)
		responseError.Title = "Could not retrieve batch"
		responseError.Code = "Internal Error"
		responseError.Error = err.Error()
		template.Error = &responseError
		template.Data = map[string]interface{}{}

		template.Render(w)
		return
	}

	recipes, err := recipe.GetRecipesByBatchId(uint(id))
	if err != nil {
		log.Println(err)
		responseError.Title = "Could not retrieve recipes"
		responseError.Code = "Internal Error"
		responseError.Error = err.Error()
		template.Error = &responseError
		template.Data = map[string]interface{}{}

		template.Render(w)
		return
	}

	dtoRecipesByBatchId.RecipesBatch = *batch
	dtoRecipesByBatchId.Recipes = recipes
	template.Data = dtoRecipesByBatchId
	template.Error = nil

	err = template.Render(w)

	if err != nil {
		log.Println(err)
		errorMsg := fmt.Sprintf("%v", err)
		http.Error(w, errorMsg, http.StatusInternalServerError)
		return
	}

}

/***************************************************************************************
* Crea un nuovo batch con i parametri forniti. Una volta creato il batch, è necessario
* chiamare l'IA per generare le ricette in base a tali parametri. Dopo aver ottenuto
* le ricette dall'IA, posso salvarle nella tabella delle ricette e associarle al
* batch appena creato. Questa operazione potrebbe richiedere molto tempo, poiché
* è necessario attendere che l'IA elabori i dati.
*
* TODO? In futuro potrebbe essere utile rendere questa richiesta di tipo
* "fire-and-forget" per poi rispondere all'utente tramite notifica or
* email.
***************************************************************************************/
func (c *ApiConfiguration) PostNewBatch(w http.ResponseWriter, r *http.Request) {

	var responseError models.ResponseError
	var response models.HttpResponse
	var batchRecipe models.RecipesBatch
	var recipe models.Recipe

	// TODO! :Delete once auth is set
	batchRecipe.UserId = 1
	batchRecipe.PromptId = 1

	// Check if the form provides valid values, otherwise return the errors
	// but also the form data to avoid resetting the form
	err := batchRecipe.MapBodyToStruct(r.Body)
	if err != nil {
		log.Printf("Error mapping body to struct: %v \n", err)

		responseError.Error = fmt.Sprintf("%v", err)
		responseError.Title = "Failed data validation"
		responseError.Code = "dataValidation"

		response.Data = responseError
		response.Code = http.StatusBadRequest

		err = response.Respond(w)

		if err != nil {
			log.Println(err)
		}
		return
	}

	// Check if the form is missing values, otherwise return the errors
	// but also the form data to avoid resetting the form
	errors := batchRecipe.Validate()
	if len(errors) > 0 {
		responseError.Title = "Incomplete data"
		responseError.Code = "dataValidation"
		responseError.Error = errors

		response.Data = responseError
		response.Code = http.StatusBadRequest

		err = response.Respond(w)
		if err != nil {
			log.Println(err)
		}
		return
	}

	// --------- AI STUFF ----------- //

	// if everything went well, save the recipe batch

	recipeBatch, savingErr := batchRecipe.Save()

	if savingErr != nil {
		responseError.Title = "Could not save"
		responseError.Code = "savingData"
		responseError.Error = savingErr.Error()

		response.Code = http.StatusInternalServerError
		response.Data = responseError

		err = response.Respond(w)

		if err != nil {
			log.Println(err)
		}
		return
	}

	// once the batch is saved, save each recipe under it
	//**************************************************************************
	//* This is a momentary loop to create fake recipes that will eventually be
	//* created by AI.
	//****************
	var srtIndex uint16 = 0
	limit := recipeBatch.RecipeCount

	recipes := make([]models.Recipe, 0, limit)

	for srtIndex < limit {
		recipe.UserId = 1
		recipe.Name = "Test"
		recipe.IsHealthy = (srtIndex % 2) == 0
		recipe.IsQuick = (srtIndex % 2) == 0
		recipe.IsMaximizeIngredients = (srtIndex % 2) == 0
		recipe.IsBudgetFriendly = (srtIndex % 2) == 0
		recipe.CuisineType = "Test"
		recipe.BatchId = 1
		recipe.Servings = uint16(srtIndex)
		recipe.Instructions = "Test...."

		recipes = append(recipes, recipe)

		srtIndex++
	}

	_, err = recipe.SaveMany(recipes)

	if err != nil {
		log.Println(err)

		responseError.Title = "Could not save recipes in batch"
		responseError.Code = "savingData"
		responseError.Error = err.Error()

		response.Code = http.StatusInternalServerError
		response.Data = err

		err = response.Respond(w)

		if err != nil {
			log.Println(err)
		}
		return
	}

	response.Code = http.StatusOK
	response.Data = recipeBatch
	response.Success = true

	err = response.Respond(w)
	if err != nil {
		log.Println(err)
	}
}
