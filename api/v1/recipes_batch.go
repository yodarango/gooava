package apiv1

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

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

// I get all the ingredients necessary for a specific batch
func (c *ApiConfiguration) GetSingleBatchIngredients(w http.ResponseWriter, r *http.Request) {

	// var recipientIngredients models.RecipeIngredientDetails

	// templateRenderer := utils.TemplateRenderer{
	// 	Title: "Batch name",
	// 	Name:  "batches_ingredients",
	// 	Data:  recipientIngredients.GetIngredientsByBatchId(213),
	// }

	// err := templateRenderer.Render(w)

	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	log.Printf("Error executing template: %v", err.Error())
	// 	return
	// }

}

// I get a single batch by its ID

// LEFT OFF HERE: study dep injection. La megliore forma da gestire DTOs. renomina DTO modeli con prefisso DTO
func (c *ApiConfiguration) GetBatchById(w http.ResponseWriter, r *http.Request, id uint) {

	var responseError models.ResponseError
	var template utils.TemplateRenderer
	var recipeBatch models.RecipesBatch

	// set the defaults of the template. Regardless of what the outcome is, these will
	// remain the same
	template.Name = "batches_id_recipes"
	template.Title = "Batch Recipes"

	// get the batch id
	var pathParts []string = strings.Split(r.URL.Path, "/")

	// if the id is not greater than 0 is not valid
	if !(len(pathParts) > 0) {
		responseError.Title = "Invalid Batch Id"
		responseError.Code = "Bad Request"
		responseError.Error = "The id provided is not valid"

		template.Data = responseError
		template.Render(w)
		return

	}

	lastPart := pathParts[len(pathParts)-1]
	batchIdUint, err := strconv.ParseUint(lastPart, 10, 32)

	if err != nil {
		responseError.Title = "Invalid Batch Id"
		responseError.Code = "Bad Request"
		responseError.Error = err.Error()

		template.Data = responseError
		template.Render(w)
		return
	}

	data, err := recipeBatch.GetBatchIngredientsById(uint(batchIdUint))

	if err != nil {
		responseError.Title = "Could not retrieve batch"
		responseError.Code = "Internal Error"
		responseError.Error = err.Error()

		template.Data = err
		template.Render(w)
		return
	}

	template.Data = data

	template.UiMeta = &utils.TemplateUIMEta{}
	template.UiMeta.MenuIcon = "restaurant-outline"

	err = template.Render(w)

	if err != nil {
		log.Println(err)
		errorMsg := fmt.Sprintf("%v", err)
		http.Error(w, errorMsg, http.StatusInternalServerError)
		return
	}

}

// I create a brand new batch
func (c *ApiConfiguration) PostNewBatch(w http.ResponseWriter, r *http.Request) {

	var responseError models.ResponseError
	var response models.HttpResponse
	var batchRecipe models.RecipesBatch

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

	// if everything went well, save the data

	recipeBatch, savingErr := batchRecipe.Save()

	if savingErr != nil {
		log.Println(savingErr)

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

	response.Code = http.StatusOK
	response.Data = recipeBatch
	response.Success = true

	err = response.Respond(w)
	if err != nil {
		log.Println(err)
	}
}
