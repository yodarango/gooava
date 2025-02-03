package apiv1

import (
	"net/http"

	"github.com/yodarango/gooava/internal/models"
)

func PostNewPantryIngredients(w http.ResponseWriter, r *http.Request) {
	var pantryIngredient models.PantryIngredient
	var httpResponse models.HttpResponse
	var responseError models.ResponseError

	pantryIngredients, err := pantryIngredient.MapBodyToStruct(r.Body, true)

	if err != nil {

		responseError.Title = "Incomplete Data"
		responseError.Error = err.Error()
		responseError.Code = "Bad Data"

		httpResponse.Code = http.StatusBadRequest
		httpResponse.Data = responseError
		httpResponse.Success = false

		httpResponse.Respond(w)

	}

	validationErrors := make([]map[string]struct{}, 0)

	for _, pi := range pantryIngredients {
		errors := pi.Validate()
		validationErrors = append(validationErrors, errors...)
	}

	if len(validationErrors) > 0 {
		responseError.Title = "Incomplete Data"
		responseError.Error = validationErrors
		responseError.Code = "Bad Data"

		httpResponse.Code = http.StatusBadRequest
		httpResponse.Data = responseError
		httpResponse.Success = false

		httpResponse.Respond(w)
	}

	// LEFT OFF. Continue the post pantry request 02/03/25
	// httpResponse.Code = http.StatusOK
	// httpResponse.Data =
}
