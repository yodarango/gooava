package apiv1

import (
	"net/http"

	"github.com/yodarango/gooava/internal/models"
	"github.com/yodarango/gooava/internal/utils"
)

func GetRecipesByBatchId(w http.ResponseWriter, r *http.Request, id uint) {

	var recipes models.Recipe
	// get the ingredients for this batch
	templateRenderer := utils.TemplateRenderer{
		Title: "batches_id_recipes",
		Name:  "batches_id_recipes",
		Data:  recipes.GetRecipesByBatchId(id),
	}

	templateRenderer.Render(w)
}

func GetAllRecipes(w http.ResponseWriter, r *http.Request) {

	var recipes models.Recipe
	// get the ingredients for this batch
	templateRenderer := utils.TemplateRenderer{
		Title: "batches_id_recipes",
		Name:  "batches_id_recipes",
		Data:  recipes.GetAll(),
	}

	templateRenderer.Render(w)
}
