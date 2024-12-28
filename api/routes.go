package api

import (
	"net/http"
	"strconv"
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
	mux.HandleFunc("/", apiv1.ApiConfig.Home)

	/********************************
	 * Batches pages
	 ********************************/
	// GET: I get all the recipe batches with pagination
	mux.HandleFunc("/batch", apiv1.ApiConfig.GetBathes)

	// POST: I create a new recipe batch
	mux.HandleFunc("/batch/new", apiv1.ApiConfig.PostNewBatch)

	// GET: I get the ingredients for a single batch
	// OR I get a single batch by id
	mux.HandleFunc("/batch/", func(w http.ResponseWriter, r *http.Request) {
		// get the path
		path := strings.TrimPrefix(r.URL.Path, "/batch/")
		pathParts := strings.Split(path, "/")

		if len(pathParts) > 1 {
			// check if the user is trying to get ingredients for a batch
			if pathParts[1] == "ingredients" {
				apiv1.ApiConfig.GetSingleBatchIngredients(w, r)
				return
			}
		}

		if len(pathParts) > 2 {
			// Check if the use is trying to get the recipes for a batch Id.
			// The ID must be present
			i, err := strconv.Atoi(pathParts[1])
			batchId := uint(i)

			if pathParts[2] == "ingredients" && err == nil {
				apiv1.ApiConfig.GetRecipesByBatchId(w, r, batchId)
				return
			}
		}

		apiv1.ApiConfig.GetBatchById(w, r)
	})

	/********************************
	 * Recipes pages
	 ********************************/
	// I get all the recipes for the logged in user
	mux.HandleFunc("/recipes", apiv1.ApiConfig.GetAllRecipes)

	return mux

}
