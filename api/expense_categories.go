package api

import (
	"encoding/json"
	"fmt"
	"gooava/constants"
	"gooava/internal/models"
	"net/http"
	"regexp"
	"strings"
)

// categories store a hex color picked with the native color input
var hexColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

const defaultCategoryColor = "#f24c66"

/************************************************************************
* Creates an expense category for the authenticated user.
*
* status: ✅
************************************************************************/
func CreateExpenseCategory(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	var category models.ExpenseCategory
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		httpResponse.Error = "Invalid request format"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if err := validateExpenseCategoryPayload(&category); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	category.UserID = authUser.Id

	if err := category.Create(); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]string{
		"message": "Category added",
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Updates one of the authenticated user's categories — a label rename re-points
* expenses linked by the old label string.
*
* status: ✅
************************************************************************/
func UpdateExpenseCategory(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	var category models.ExpenseCategory
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		httpResponse.Error = "Invalid request format"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if category.Id <= 0 {
		httpResponse.Error = "id is required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if err := validateExpenseCategoryPayload(&category); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	category.UserID = authUser.Id

	if err := category.Update(); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]string{
		"message": "Category updated",
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Deletes one of the authenticated user's categories by id.
*
* status: ✅
************************************************************************/
func DeleteExpenseCategory(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	var requestBody struct {
		Id int64 `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil || requestBody.Id <= 0 {
		httpResponse.Error = "id is required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if err := models.DeleteExpenseCategory(authUser.Id, requestBody.Id); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]string{
		"message": "Category deleted",
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/**************************************************************************************
* Shared create/update normalization: trimmed label (required), trimmed description,
* hex color (defaults when blank).
**************************************************************************************/
func validateExpenseCategoryPayload(category *models.ExpenseCategory) error {
	category.Label = strings.TrimSpace(category.Label)
	if category.Label == "" {
		return fmt.Errorf("label is required")
	}

	category.Description = strings.TrimSpace(category.Description)

	category.Color = strings.TrimSpace(category.Color)
	if category.Color == "" {
		category.Color = defaultCategoryColor
	}
	if !hexColorPattern.MatchString(category.Color) {
		return fmt.Errorf("color must be a hex value like #f24c66")
	}

	return nil
}

/************************************************************************
* Lists the authenticated user's expense categories, alphabetically.
*
* status: ✅
************************************************************************/
func ListExpenseCategories(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	categories, err := models.ListExpenseCategories(authUser.Id)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = categories
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}
