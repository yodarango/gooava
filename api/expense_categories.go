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

	category.Label = strings.TrimSpace(category.Label)
	if category.Label == "" {
		httpResponse.Error = "label is required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	category.Description = strings.TrimSpace(category.Description)

	category.Color = strings.TrimSpace(category.Color)
	if category.Color == "" {
		category.Color = defaultCategoryColor
	}
	if !hexColorPattern.MatchString(category.Color) {
		httpResponse.Error = "color must be a hex value like #f24c66"
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
