package api

import (
	"net/http"

	apiv1 "github.com/yodarango/gooava/api/v1"
)

func Routes() http.Handler {

	mux := http.NewServeMux()
	// serve the static files
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/batches", apiv1.GetBathes)

	mux.HandleFunc("/batches/2/ingredients", apiv1.GetSingleBatchIngredients)

	mux.HandleFunc("/batches/new", apiv1.PostNewBatch)

	mux.HandleFunc("/batches/", apiv1.GetBatchById)

	mux.HandleFunc("/", apiv1.Public)

	return mux

}
