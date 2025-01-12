package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	apiv1 "github.com/yodarango/gooava/api/v1"
	"github.com/yodarango/gooava/internal/constants"
	"github.com/yodarango/gooava/internal/utils"
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
	mux.HandleFunc(constants.ROUTE_ROOT, apiv1.ApiConfig.Home)

	/********************************
	 * Batches pages
	 ********************************/
	// GET: I get all the recipe batches with pagination
	mux.HandleFunc(constants.ROUTE_RECIPEBATCHES, apiv1.ApiConfig.GetBathes)

	// POST: I create a new recipe batch
	mux.HandleFunc(constants.ROUTE_RECIPE_BATCHES_NEW, apiv1.ApiConfig.PostNewBatch)

	// GET: I get the ingredients for a single batch
	// OR I get a single batch by id
	mux.HandleFunc(constants.ROUTE_RECIPEBATCHES+"/", func(w http.ResponseWriter, r *http.Request) {
		// get the path and remove the router root path

		// check if the user is trying to get ingredients for a batch
		_, isIngredientsPath := utils.MakePathFromRoute(r.URL.Path, constants.ROUTE_RECIPEBATCHES_ID_INGREDIENTS)
		if isIngredientsPath {
			apiv1.ApiConfig.GetSingleBatchIngredients(w, r)
			return

		}

		// Check if the use is trying to get a batch by ID
		_, isBatchByIdPath := utils.MakePathFromRoute(r.URL.Path, constants.ROUTE_RECIPEBATCHES_ID)
		fmt.Println(isBatchByIdPath)
		if isBatchByIdPath {

			// I need to get the last part of the path which MUST be the ID or I will fail
			pathParts := strings.Split(r.URL.Path, "/")
			batchId := pathParts[len(pathParts)-1]
			i, err := strconv.ParseUint(batchId, 10, 32)

			if err != nil {
				http.Error(w, "Invalid batch Id", http.StatusNotAcceptable)
				return
			}

			// Now convert 64 to 32
			id := uint(i)

			apiv1.ApiConfig.GetBatchById(w, r, id)
			return

		}

		// If none of the above conditions where met, there must be an error
		// TODO: Build an error template
		http.Error(w, "Invalid batch Id", http.StatusNotAcceptable)
	})

	/********************************
	 * Recipes pages
	 ********************************/
	// I get all the recipes for the logged in user
	mux.HandleFunc(constants.ROUTE_RECIPES, apiv1.ApiConfig.GetAllRecipes)

	return mux

}
