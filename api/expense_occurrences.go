package api

import (
	"encoding/json"
	"fmt"
	"gooava/constants"
	"gooava/internal/models"
	"net/http"
	"time"
)

/************************************************************************
* Ensures the given expense occurrences exist (created lazily the first time the
* paycheck plan's window includes them), auto-pays any whose label string-matches a
* synced Plaid transaction, then returns every occurrence in the requested range so
* the plan can render paid/unpaid state.
*
* status: ✅
************************************************************************/
func EnsureExpenseOccurrences(w http.ResponseWriter, r *http.Request) {
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
		From  string `json:"from"` // YYYY-MM-DD
		To    string `json:"to"`   // YYYY-MM-DD
		Items []struct {
			ExpenseID int64  `json:"expense_id"`
			DueOn     string `json:"due_on"`
		} `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		httpResponse.Error = "Invalid request format"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if !incomeDatePattern.MatchString(requestBody.From) || !incomeDatePattern.MatchString(requestBody.To) {
		httpResponse.Error = "from and to must be YYYY-MM-DD"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}
	if _, err := time.Parse("2006-01-02", requestBody.From); err != nil {
		httpResponse.Error = "from must be a valid date"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}
	if _, err := time.Parse("2006-01-02", requestBody.To); err != nil {
		httpResponse.Error = "to must be a valid date"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	// validate each item before touching the db
	ensureItems := make([]struct {
		ExpenseID int64
		DueOn     string
		Amount    float64
	}, 0, len(requestBody.Items))
	for _, item := range requestBody.Items {
		if item.ExpenseID <= 0 || !incomeDatePattern.MatchString(item.DueOn) {
			httpResponse.Error = "each item needs a valid expense_id and due_on (YYYY-MM-DD)"
			httpResponse.Success = false
			httpResponse.Data = nil
			httpResponse.Send(w)
			return
		}
		if _, err := time.Parse("2006-01-02", item.DueOn); err != nil {
			httpResponse.Error = "each item's due_on must be a valid date"
			httpResponse.Success = false
			httpResponse.Data = nil
			httpResponse.Send(w)
			return
		}
		ensureItems = append(ensureItems, struct {
			ExpenseID int64
			DueOn     string
			Amount    float64
		}{ExpenseID: item.ExpenseID, DueOn: item.DueOn})
	}

	if err := models.EnsureExpenseOccurrences(authUser.Id, ensureItems); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	// auto-mark any occurrence that matches a synced transaction by label
	if _, err := models.AutoPayOccurrencesByLabel(authUser.Id); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	occurrences, err := models.ListExpenseOccurrences(authUser.Id, requestBody.From, requestBody.To)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = occurrences
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Marks one of the authenticated user's expense occurrences paid or unpaid.
*
* status: ✅
************************************************************************/
func SetExpenseOccurrencePaid(w http.ResponseWriter, r *http.Request) {
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
		Id   int64 `json:"id"`
		Paid bool  `json:"paid"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil || requestBody.Id <= 0 {
		httpResponse.Error = "a valid id is required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	changed, err := models.SetExpenseOccurrencePaid(authUser.Id, requestBody.Id, requestBody.Paid)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if !changed {
		httpResponse.Error = "expense occurrence not found"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]string{
		"message": map[bool]string{true: "Marked as paid", false: "Marked as unpaid"}[requestBody.Paid],
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}
