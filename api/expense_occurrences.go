package api

import (
	"encoding/json"
	"fmt"
	"gooava/constants"
	"gooava/internal/models"
	"net/http"
	"strings"
	"time"
)

// occurrenceEnsureRequest is the shared body of the ensure and review endpoints: the
// date range to read back plus the (expense_id, due_on) pairs to lazily create.
type occurrenceEnsureRequest struct {
	From  string `json:"from"` // YYYY-MM-DD
	To    string `json:"to"`   // YYYY-MM-DD
	Items []struct {
		ExpenseID int64  `json:"expense_id"`
		DueOn     string `json:"due_on"`
	} `json:"items"`
}

/************************************************************************
* Parses the ensure/review request, validates every date, lazily creates the given
* occurrences, and returns the requested range. Writes the error response and returns
* ok=false on any failure.
************************************************************************/
func ensureOccurrencesFromRequest(w http.ResponseWriter, r *http.Request, userID uint) (from, to string, ok bool) {
	var httpResponse models.HttpResponse
	fail := func(message string) (string, string, bool) {
		httpResponse.Error = message
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return "", "", false
	}

	var requestBody occurrenceEnsureRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		return fail("Invalid request format")
	}

	if !incomeDatePattern.MatchString(requestBody.From) || !incomeDatePattern.MatchString(requestBody.To) {
		return fail("from and to must be YYYY-MM-DD")
	}
	if _, err := time.Parse("2006-01-02", requestBody.From); err != nil {
		return fail("from must be a valid date")
	}
	if _, err := time.Parse("2006-01-02", requestBody.To); err != nil {
		return fail("to must be a valid date")
	}

	// validate each item before touching the db
	ensureItems := make([]struct {
		ExpenseID int64
		DueOn     string
		Amount    float64
	}, 0, len(requestBody.Items))
	for _, item := range requestBody.Items {
		if item.ExpenseID <= 0 || !incomeDatePattern.MatchString(item.DueOn) {
			return fail("each item needs a valid expense_id and due_on (YYYY-MM-DD)")
		}
		if _, err := time.Parse("2006-01-02", item.DueOn); err != nil {
			return fail("each item's due_on must be a valid date")
		}
		ensureItems = append(ensureItems, struct {
			ExpenseID int64
			DueOn     string
			Amount    float64
		}{ExpenseID: item.ExpenseID, DueOn: item.DueOn})
	}

	if err := models.EnsureExpenseOccurrences(userID, ensureItems); err != nil {
		return fail(fmt.Sprintf("%v", err))
	}

	return requestBody.From, requestBody.To, true
}

/************************************************************************
* Ensures the given expense occurrences exist (created lazily the first time the
* paycheck plan's window includes them), auto-pays the ones whose expense the user
* explicitly linked to a transaction, then returns every occurrence in the requested
* range so the plan can render paid/unpaid state.
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

	from, to, valid := ensureOccurrencesFromRequest(w, r, authUser.Id)
	if !valid {
		return
	}

	// auto-mark occurrences of expenses the user linked to a transaction — a coincidental
	// label match alone is never enough to mark a bill paid without confirmation
	if _, err := models.AutoPayLinkedOccurrences(authUser.Id); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	occurrences, err := models.ListExpenseOccurrences(authUser.Id, from, to)
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
* The lazy past-due review run each time the app opens: ensures the past occurrences of
* the requested range exist, lets linked expenses pay themselves, then returns every
* still-unpaid occurrence in the range with the transaction it string-matches (nil when
* none). The frontend splits the result into the "found matching transactions" card and
* the manual past-due card.
*
* status: ✅
************************************************************************/
func ReviewPastDueOccurrences(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	from, to, valid := ensureOccurrencesFromRequest(w, r, authUser.Id)
	if !valid {
		return
	}

	// linked expenses pay themselves first — only unlinked bills need the user's review
	if _, err := models.AutoPayLinkedOccurrences(authUser.Id); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	occurrences, err := models.ListPastDueOccurrences(authUser.Id, from, to)
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
* Marks the given occurrences paid with the transaction that paid each one — the
* confirmation behind the "found matching transactions" card (and the link action on
* the manual card). Each transaction must belong to the user.
*
* status: ✅
************************************************************************/
func ConfirmExpenseOccurrences(w http.ResponseWriter, r *http.Request) {
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
		Items []models.OccurrencePayment `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		httpResponse.Error = "Invalid request format"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if len(requestBody.Items) == 0 || len(requestBody.Items) > 200 {
		httpResponse.Error = "items must contain between 1 and 200 entries"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}
	for _, item := range requestBody.Items {
		if item.Id <= 0 || strings.TrimSpace(item.TransactionID) == "" {
			httpResponse.Error = "each item needs a valid id and transaction_id"
			httpResponse.Success = false
			httpResponse.Data = nil
			httpResponse.Send(w)
			return
		}
	}

	marked, err := models.SetExpenseOccurrencesPaidWithTransactions(authUser.Id, requestBody.Items)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]string{
		"message": fmt.Sprintf("Marked %d as paid", marked),
	}
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
