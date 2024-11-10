package apiv1

import (
	"fmt"
	"html/template"
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

// I get all the ingredients necessary for a specific batch
func (c *ApiConfiguration) GetSingleBatchIngredients(w http.ResponseWriter, r *http.Request) {
	templ := template.New("batches_ingredients.html").Funcs(template.FuncMap{
		"add": func(x, y int) int {
			return x + y
		},
	})
	temp, err := templ.ParseFiles("web/templates/batches_ingredients.html", "web/templates/partials/base.html", "web/templates/partials/header.html")

	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		log.Printf("Error parsing template %v", err)
		return
	}

	err = temp.Execute(w, map[string]interface{}{
		"Data": []map[string]interface{}{
			{"Name": "my recipe", "Qty": 2},
			{"Name": "my recipe", "Qty": 5},
			{"Name": "my recipe", "Qty": 22},
			{"Name": "my recipe", "Qty": 6},
			{"Name": "my recipe", "Qty": 42},
			{"Name": "my recipe", "Qty": 1},
			{"Name": "my recipe", "Qty": 68},
		},
		"MenuIcon": "restaurant",
	})

	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		log.Printf("Error executing template: %v", err)
		return
	}

}

// I get a single batch by its ID
func (c *ApiConfiguration) GetBatchById(w http.ResponseWriter, r *http.Request) {
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
	template := utils.TemplateRenderer{
		Name:  "batches",
		Title: "Batches",
		Data:  map[string]interface{}{},
	}

	// Check if the form provides valid values, otherwise return the errors
	// but also the form data to avoid resetting the form
	errors := batch.MapFormToStruct(r.Form)
	if errors != nil {
		template.Data = map[string]interface{}{
			"FormValidationError": errors,
			"Form":                batch,
		}
		template.Render(w)
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

	// if everything went well, save the data
	savingErr := batch.Save()
	if savingErr != nil {
		template.Data = map[string]interface{}{
			"Error": err,
		}
		template.Render(w)
		return
	}
}
