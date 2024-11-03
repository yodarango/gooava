package apiv1

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func GetBathes(w http.ResponseWriter, r *http.Request) {
	templ := template.New("batches.html")

	temp, err := templ.ParseFiles("web/templates/batches.html", "web/templates/partials/base.html", "web/templates/partials/header.html")

	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing template %v", err), http.StatusInternalServerError)
		log.Printf("Error parsing template %v", err)
		return
	}

	err = temp.Execute(w, map[string]interface{}{
		"Data": []map[string]interface{}{{
			"Name":      "the name 1",
			"Date":      "The date 1",
			"Thumbnail": "https://picsum.photos/300/300",
			"Id":        "1",
		}, {
			"Name":      "the name 2",
			"Date":      "The date 2",
			"Thumbnail": "https://picsum.photos/300/300",
			"Id":        "2",
		}},
		"MenuIcon": "restaurant",
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Error executing the template %v", err), http.StatusInternalServerError)
		log.Printf("Error executing the template %v", err)
	}

}

func GetSingleBatchIngredients(w http.ResponseWriter, r *http.Request) {
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

func GetBatchById(w http.ResponseWriter, r *http.Request) {
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
