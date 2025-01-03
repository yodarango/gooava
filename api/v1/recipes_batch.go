package apiv1

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/yodarango/gooava/internal/dto"
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

	var recipientIngredients dto.RecipeIngredientDetails

	templateRenderer := utils.TemplateRenderer{
		Title: "Batch name",
		Name:  "batches_ingredients",
		Data:  recipientIngredients.GetIngredientsByBatchId(213),
	}

	err := templateRenderer.Render(w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error executing template: %v", err.Error())
		return
	}

}

// I get a single batch by its ID
func (c *ApiConfiguration) GetBatchById(w http.ResponseWriter, r *http.Request, id uint) {
	templ := template.New("batches_recipes.html").Funcs(template.FuncMap{
		"add": func(x, y int) int {
			return x + y
		},
	})
	temp, err := templ.ParseFiles("web/templates/batches_recipes.html", "web/templates/partials/base.html", "web/templates/partials/header.html")

	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		log.Printf("Error parsing template %v", err)
		return
	}

	err = temp.Execute(w, map[string]interface{}{
		"Data": []map[string]interface{}{
			{"Name": "my recipe"},
			{"Name": "my second recipe"},
			{"Name": "my third recipe"},
			{"Name": "my fourth recipe"},
			{"Name": "my fifth recipe"},
			{"Name": "my sixth recipe"},
			{"Name": "my seventh recipe"},
		},
		"MenuIcon": "restaurant",
	})

	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		log.Printf("Error executing template: %v", err)
		return
	}

}

// I create a brand new batch
func (c *ApiConfiguration) PostNewBatch(w http.ResponseWriter, r *http.Request) {

	var batch models.RecipesBatch
	var template utils.TemplateRenderer

	// Check if the form provides valid values, otherwise return the errors
	// but also the form data to avoid resetting the form
	errors := batch.MapFormToStruct(r.Form)

	if len(errors) > 0 {
		template.Name = "index"
		template.Title = "Home"
		template.Data = map[string]interface{}{
			"FormValidationError": errors,
			"Form":                batch,
		}

		err := template.Render(w)

		if err != nil {
			fmt.Println(err)
			http.Error(w, "Sorry, there was an issue!", http.StatusInternalServerError)
			return
		}
		return
	}

	// Check if the form is missing values, otherwise return the errors
	// but also the form data to avoid resetting the form
	err := batch.Validate()
	if err != nil {
		template.Data = map[string]interface{}{
			"FormValidationError": errors,
			"Form":                batch,
		}
		template.Render(w)
		return
	}

	// --------- AI STUFF ----------- //

	// if everything went well, save the data
	recipeBatch, savingErr := batch.Save()
	if savingErr != nil {
		template.Data = map[string]interface{}{
			"Error": err,
		}
		template.Render(w)
		return
	}

	recipeBatchId := strconv.FormatUint(uint64(recipeBatch.Id), 10)
	redirectToRoute := utils.MakeRouteFromPath(r.URL.Path, recipeBatchId)

	// if everything went well redirect to the batch page
	http.Redirect(w, r, redirectToRoute, http.StatusSeeOther)

}
