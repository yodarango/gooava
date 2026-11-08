package api

import (
	"encoding/json"
	"fmt"
	"gooava/constants"
	"gooava/internal/models"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var monthPattern = regexp.MustCompile(`^\d{4}-\d{2}$`)

// occurrenceEnsureItem is one (expense_id, due_on) pair to ensure, plus the optional
// plan-computed amount (only the put away expense carries one).
type occurrenceEnsureItem struct {
	ExpenseID int64   `json:"expense_id"`
	DueOn     string  `json:"due_on"`
	Amount    float64 `json:"amount"` // >0 only for plan-materialized rows (the put away expense)
}

// occurrenceEnsureRequest is the shared body of the ensure and review endpoints: the
// date range to read back plus the items to lazily create.
type occurrenceEnsureRequest struct {
	From  string                 `json:"from"` // YYYY-MM-DD
	To    string                 `json:"to"`   // YYYY-MM-DD
	Items []occurrenceEnsureItem `json:"items"`
}

/************************************************************************
* Parses the ensure/review request, validates every date, lazily creates the given
* occurrences, and returns the requested range plus the validated items. Writes the
* error response and returns ok=false on any failure.
************************************************************************/
func ensureOccurrencesFromRequest(w http.ResponseWriter, r *http.Request, userID uint) (from, to string, items []occurrenceEnsureItem, ok bool) {
	var httpResponse models.HttpResponse
	fail := func(message string) (string, string, []occurrenceEnsureItem, bool) {
		httpResponse.Error = message
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return "", "", nil, false
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
		}{ExpenseID: item.ExpenseID, DueOn: item.DueOn, Amount: item.Amount})
	}

	if err := models.EnsureExpenseOccurrences(userID, ensureItems); err != nil {
		return fail(fmt.Sprintf("%v", err))
	}

	return requestBody.From, requestBody.To, requestBody.Items, true
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

	from, to, items, valid := ensureOccurrencesFromRequest(w, r, authUser.Id)
	if !valid {
		return
	}

	// the put away expense is materialized by the plan: remove its stale unpaid,
	// non-custom rows — the rendered set must exactly match the plan's current math
	putAwayExpenseID, err := models.GetPutAwayExpenseID(authUser.Id)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}
	if putAwayExpenseID > 0 {
		keepDates := make([]string, 0)
		for _, item := range items {
			if item.ExpenseID == putAwayExpenseID {
				keepDates = append(keepDates, item.DueOn)
			}
		}
		if err := models.PruneExpenseOccurrences(authUser.Id, putAwayExpenseID, from, to, keepDates); err != nil {
			httpResponse.Error = fmt.Sprintf("%v", err)
			httpResponse.Success = false
			httpResponse.Data = nil
			httpResponse.Send(w)
			return
		}
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

	from, to, _, valid := ensureOccurrencesFromRequest(w, r, authUser.Id)
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
* Lists one month of the authenticated user's billing history: every paid occurrence
* whose due date falls in the requested month (?month=YYYY-MM, default the current
* month), newest due date first. has_more tells the client whether any earlier month
* holds paid rows, so it knows when to stop paging.
*
* status: ✅
************************************************************************/
func ListExpenseOccurrenceHistory(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	monthStart := time.Now()
	if monthParam := r.URL.Query().Get("month"); monthParam != "" {
		if !monthPattern.MatchString(monthParam) {
			httpResponse.Error = "month must be YYYY-MM"
			httpResponse.Success = false
			httpResponse.Data = nil
			httpResponse.Send(w)
			return
		}
		parsed, err := time.Parse("2006-01", monthParam)
		if err != nil {
			httpResponse.Error = "month must be a valid month"
			httpResponse.Success = false
			httpResponse.Data = nil
			httpResponse.Send(w)
			return
		}
		monthStart = parsed
	}

	from := time.Date(monthStart.Year(), monthStart.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	to := time.Date(monthStart.Year(), monthStart.Month()+1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

	occurrences, err := models.ListPaidExpenseOccurrences(authUser.Id, from, to)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	hasMore, err := models.HasPaidExpenseOccurrencesBefore(authUser.Id, from)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]interface{}{
		"items":    occurrences,
		"has_more": hasMore,
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


/************************************************************************
* Hand-sets the amount of one of the authenticated user's unpaid expense
* occurrences (the pencil edit next to a bill). Paid rows are billing history
* and stay frozen.
*
* status: ✅
************************************************************************/
func UpdateExpenseOccurrenceAmount(w http.ResponseWriter, r *http.Request) {
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
		Id     int64    `json:"id"`
		Amount *float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil || requestBody.Id <= 0 || requestBody.Amount == nil {
		httpResponse.Error = "a valid id and amount are required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if *requestBody.Amount < 0 {
		httpResponse.Error = "amount must be 0 or more"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	changed, err := models.UpdateExpenseOccurrenceAmount(authUser.Id, requestBody.Id, *requestBody.Amount)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if !changed {
		httpResponse.Error = "expense occurrence not found or already paid"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]string{
		"message": "Amount updated",
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}
