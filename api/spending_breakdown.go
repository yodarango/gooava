package api

import (
	"fmt"
	"gooava/constants"
	"gooava/internal/models"
	"net/http"
)

/************************************************************************
* Returns the authenticated user's 36-month spending breakdown and
* category totals with percentage of total income.
************************************************************************/
func GetSpendingBreakdown(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	data, err := models.GetSpendingBreakdown(authUser.Id)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = data
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}
