package api

import (
	"encoding/json"
	"fmt"
	"gooava/constants"
	"gooava/internal/models"
	"net/http"
	"strconv"
)

// the paycheck plan shows this many windows unless the user picks another allowed value
const defaultPaycheckWindows = 6

var allowedPaycheckWindows = map[int]bool{3: true, 6: true, 9: true, 12: true, 24: true, 36: true}

/************************************************************************
* Returns the authenticated user's finances preferences, with defaults for
* anything never saved.
*
* status: ✅
************************************************************************/
func GetFinancePreferences(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	settings, err := models.ListUserSettings(authUser.Id)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	paycheckWindows := defaultPaycheckWindows
	if saved, exists := settings[models.SETTING_KEY_PAYCHECK_WINDOWS]; exists {
		if parsed, err := strconv.Atoi(saved); err == nil && allowedPaycheckWindows[parsed] {
			paycheckWindows = parsed
		}
	}

	httpResponse.Data = map[string]interface{}{
		"paycheck_windows": paycheckWindows,
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Saves the authenticated user's finances preferences.
*
* status: ✅
************************************************************************/
func SaveFinancePreferences(w http.ResponseWriter, r *http.Request) {
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
		PaycheckWindows *int `json:"paycheck_windows"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil || requestBody.PaycheckWindows == nil {
		httpResponse.Error = "paycheck_windows is required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if !allowedPaycheckWindows[*requestBody.PaycheckWindows] {
		httpResponse.Error = "paycheck_windows must be one of 3, 6, 9, 12, 24, 36"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if err := models.UpsertUserSetting(authUser.Id, models.SETTING_KEY_PAYCHECK_WINDOWS, strconv.Itoa(*requestBody.PaycheckWindows)); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]string{
		"message": "Preferences updated",
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}
