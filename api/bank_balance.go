package api

import (
	"encoding/json"
	"fmt"
	"gooava/constants"
	"gooava/internal/models"
	"net/http"
)

/************************************************************************
* Returns the authenticated user's manually entered "in the bank" balance
* (null when none has been saved yet).
*
* status: ✅
************************************************************************/
func GetBankBalance(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	balance, err := models.GetBankBalance(authUser.Id)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]interface{}{
		"balance": balance,
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Saves the authenticated user's manually entered "in the bank" balance.
*
* status: ✅
************************************************************************/
func SaveBankBalance(w http.ResponseWriter, r *http.Request) {
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
		Balance *float64 `json:"balance"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil || requestBody.Balance == nil {
		httpResponse.Error = "balance is required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if *requestBody.Balance < 0 {
		httpResponse.Error = "balance must be 0 or more"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if err := models.UpsertBankBalance(authUser.Id, *requestBody.Balance); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]string{
		"message": "Balance updated",
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}
