package api

import (
	"net/http"
	"strconv"
	"strings"

	apiv1 "github.com/yodarango/gooava/api/v1"
	"github.com/yodarango/gooava/internal/constants"
	"github.com/yodarango/gooava/internal/utils"
)

func Routes() http.Handler {

	mux := http.NewServeMux()
	// I servo tutti i file statici
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	/***************************************************************************************
	 * GET: Fornisco la pagina di inizio
	 ***************************************************************************************/
	mux.HandleFunc(constants.ROUTE_ROOT, apiv1.ApiConfig.Home)

	/***************************************************************************************
	 * GET: Fornisco tutti i RecipeBatches con impaginazione
	 ***************************************************************************************/
	mux.HandleFunc(constants.ROUTE_RECIPEBATCHES, apiv1.ApiConfig.GetBathes)

	/***************************************************************************************
	 * POST: Creero una nuovo batch
	 ***************************************************************************************/
	mux.HandleFunc(constants.ROUTE_RECIPE_BATCHES_NEW, apiv1.ApiConfig.PostNewBatch)

	/***************************************************************************************
	 * GET: Tratta tutte le richieste sparati a /recipe-batches/, quindi devve assicurse di
	 * analizare il percorso della richiesta e chiamare il handler corretto. Per
	 * queto utiliza vari funzione di utilità.
	 *
	 * TODO? bisogna di avere un framework piu strutturato per riaggungere questo compito
	 ***************************************************************************************/
	mux.HandleFunc(constants.ROUTE_RECIPEBATCHES+"/", func(w http.ResponseWriter, r *http.Request) {

		// queto significa che loro volgliono vedere colo i ingredenti dentro da un batch
		_, isIngredientsPath := utils.MakePathFromRoute(r.URL.Path, constants.ROUTE_RECIPEBATCHES_ID_INGREDIENTS)
		if isIngredientsPath {

			// Prima di chiamare il batch, devo assicurarmi che il id e veramente valido
			pathParts := strings.Split(r.URL.Path, "/")
			batchId := pathParts[len(pathParts)-2]
			i, err := strconv.ParseUint(batchId, 10, 32)

			if err == nil {
				id := uint(i)
				apiv1.ApiConfig.GetSingleBatchIngredients(w, r, id)
				return
			}

		}

		// se non vogliono i ingredenti, questo significa che vogliono vedere il batch compieto
		_, isBatchByIdPath := utils.MakePathFromRoute(r.URL.Path, constants.ROUTE_RECIPEBATCHES_ID)

		if isBatchByIdPath {

			// Prima di chiamare il batch, devo assicurarmi che il id e veramente valido
			pathParts := strings.Split(r.URL.Path, "/")
			batchId := pathParts[len(pathParts)-1]
			i, err := strconv.ParseUint(batchId, 10, 32)

			if err == nil {
				// Now convert 64 to 32
				id := uint(i)

				apiv1.ApiConfig.GetBatchById(w, r, id)
				return
			}

		}

		// Se non abbina nessunre delle condizzine sopra, quindi responde con un errore perche
		// sono probando a chiamare una pagina che non essiste.
		// TODO: construie una template per 404

		http.Error(w, "Invalid batch Id", http.StatusNotAcceptable)
	})

	/***************************************************************************************
	 * GET: Fornisco tutte le recete per il usuario indicato
	 ***************************************************************************************/
	mux.HandleFunc(constants.ROUTE_RECIPES, apiv1.ApiConfig.GetAllRecipes)

	return mux

}
