package api

import (
	"net/http"
	"strings"

	apiv1 "github.com/yodarango/gooava/api/v1"
)

func Routes() http.Handler {

	mux := http.NewServeMux()
	// serve the static files
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	/********************************
	 * Public pages
	 ********************************/
	// GET: I get the homepage
	mux.HandleFunc("/", apiv1.Public)

	/********************************
	 * Batches pages
	 ********************************/
	// GET: I get all the recipe batches with pagination
	mux.HandleFunc("/recipe-batch", apiv1.GetBathes)

	// POST: I creates a new recipe batch
	mux.HandleFunc("/recipe-batch/new", apiv1.PostNewBatch)

	// GET: I get the ingredients for a single batch
	// OR I get a single batch by id
	mux.HandleFunc("/recipe-batch/", func(w http.ResponseWriter, r *http.Request) {
		// get the path
		path := strings.TrimPrefix(r.URL.Path, "/recipe-batch/")
		pathParts := strings.Split(path, "/")

		if len(pathParts) > 1 {
			// check if the user is trying to get ingredients for a batch
			if pathParts[1] == "ingredients" {
				apiv1.GetSingleBatchIngredients(w, r)
				return
			}
		}

		apiv1.GetBatchById(w, r)
	})

	/********************************
	 * Recipes pages
	 ********************************/
	// I get all the recipes for the logged in user
	mux.HandleFunc("/recipes", apiv1.GetRecipes)

	return mux

}
