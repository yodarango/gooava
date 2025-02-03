package apiv1

import (
	"net/http"

	"github.com/yodarango/gooava/internal/models"
	"github.com/yodarango/gooava/internal/utils"
)

func GetRecipesByBatchId(w http.ResponseWriter, r *http.Request, id uint) {

	// get the ingredients for this batch
	templateRenderer := utils.TemplateRenderer{
		Title: "batches_id",
		Name:  "batches_id",
		Data:  "niente", // recipes.GetRecipesByBatchId(id),
	}

	templateRenderer.Render(w)
}

func GetAllRecipes(w http.ResponseWriter, r *http.Request) {

	var recipes models.Recipe
	// get the ingredients for this batch
	templateRenderer := utils.TemplateRenderer{
		Title: "batches_id",
		Name:  "batches_id",
		Data:  recipes.GetAll(),
	}

	templateRenderer.Render(w)
}

/***************************************************************************************
* Ottiene tutta la informazione relazionata con una recetta specifica per la Id. Non
* fornice nessuna altra informazione che non sia relazionata con la recetta.
***************************************************************************************/
func GetRecipeById(w http.ResponseWriter, r *http.Request) {

}
