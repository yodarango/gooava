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

var incomeDatePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

/************************************************************************
* Creates an income source for the authenticated user.
*
* status: ✅
************************************************************************/
func CreateIncome(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	var income models.Income
	if err := json.NewDecoder(r.Body).Decode(&income); err != nil {
		httpResponse.Error = "Invalid request format"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if err := validateIncomePayload(&income); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	income.UserID = authUser.Id

	if err := income.Create(); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]string{
		"message": "Income added",
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Lists the authenticated user's income sources, newest first.
*
* status: ✅
************************************************************************/
func ListIncomes(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	incomes, err := models.ListIncomes(authUser.Id)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = incomes
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Updates one of the authenticated user's incomes (every editable field).
*
* status: ✅
************************************************************************/
func UpdateIncome(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	var income models.Income
	if err := json.NewDecoder(r.Body).Decode(&income); err != nil {
		httpResponse.Error = "Invalid request format"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if income.Id <= 0 {
		httpResponse.Error = "id is required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if err := validateIncomePayload(&income); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	income.UserID = authUser.Id

	if err := income.Update(); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]string{
		"message": "Income updated",
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Deletes one of the authenticated user's incomes by id.
*
* status: ✅
************************************************************************/
func DeleteIncome(w http.ResponseWriter, r *http.Request) {
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

	if err := models.DeleteIncome(authUser.Id, requestBody.Id); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]string{
		"message": "Income deleted",
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Lists the authenticated user's income approval entries, newest first.
*
* status: ✅
************************************************************************/
func ListIncomeEntries(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	entries, err := models.ListIncomeEntries(authUser.Id)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = entries
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Records one approved income occurrence (the frontend computes the schedule).
* Any rejected rows for the same income+date are replaced so re-approval works.
*
* status: ✅
************************************************************************/
func ApproveIncomeEntry(w http.ResponseWriter, r *http.Request) {
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
		IncomeID   int64   `json:"income_id"`
		OccurredOn string  `json:"occurred_on"`
		Amount     float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		httpResponse.Error = "Invalid request format"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if err := validateIncomeEntryPayload(authUser.Id, requestBody.IncomeID, requestBody.OccurredOn, requestBody.Amount); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if err := models.ApproveIncomeEntry(authUser.Id, requestBody.IncomeID, requestBody.OccurredOn, requestBody.Amount); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]string{
		"message": "Income approved",
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Records one rejected income occurrence so it stops showing up as pending.
*
* status: ✅
************************************************************************/
func RejectIncomeEntry(w http.ResponseWriter, r *http.Request) {
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
		IncomeID   int64   `json:"income_id"`
		OccurredOn string  `json:"occurred_on"`
		Amount     float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		httpResponse.Error = "Invalid request format"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if err := validateIncomeEntryPayload(authUser.Id, requestBody.IncomeID, requestBody.OccurredOn, requestBody.Amount); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if err := models.RejectIncomeEntry(authUser.Id, requestBody.IncomeID, requestBody.OccurredOn, requestBody.Amount); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]string{
		"message": "Income rejected",
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/**************************************************************************************
* validateIncomeEntryPayload checks the shared approve/reject fields and that the
* income belongs to the user. Rejections may carry amount 0 (amount is informational).
**************************************************************************************/
func validateIncomeEntryPayload(userID uint, incomeID int64, occurredOn string, amount float64) error {
	if incomeID <= 0 {
		return fmt.Errorf("income_id is required")
	}

	if !incomeDatePattern.MatchString(occurredOn) {
		return fmt.Errorf("occurred_on must be YYYY-MM-DD")
	}
	if _, err := time.Parse("2006-01-02", occurredOn); err != nil {
		return fmt.Errorf("occurred_on must be a valid date")
	}

	if amount < 0 {
		return fmt.Errorf("amount must be 0 or more")
	}

	incomes, err := models.ListIncomes(userID)
	if err != nil {
		return err
	}
	for _, income := range incomes {
		if income.Id == incomeID {
			return nil
		}
	}

	return fmt.Errorf("income not found")
}

/**************************************************************************************
* validateIncomePayload runs the shared create/update checks and normalizes the income
* in place (trims the label, rebuilds the rule in its canonical shape per frequency).
**************************************************************************************/
func validateIncomePayload(income *models.Income) error {
	income.Label = strings.TrimSpace(income.Label)
	if income.Label == "" {
		return fmt.Errorf("label is required")
	}

	if income.Amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	if !allowedIncomeFrequencies[income.Recurring] {
		return fmt.Errorf("invalid recurring value")
	}

	rule, err := normalizeIncomeRecurRule(income.Recurring, income.RecurRule)
	if err != nil {
		return err
	}
	income.RecurRule = rule

	return nil
}

var allowedIncomeFrequencies = map[string]bool{
	"": true, "weekly": true, "biweekly": true, "monthly": true,
	"quarterly": true, "semiannually": true, "annually": true,
}

// incomeRecurRule mirrors the JSON shapes the frontend sends per frequency.
type incomeRecurRule struct {
	Date        string `json:"date"`          // one-time: YYYY-MM-DD
	DayOfWeek   *int   `json:"day_of_week"`   // weekly: 0 (Sunday) - 6 (Saturday)
	Anchor      string `json:"anchor"`        // biweekly mode "every_other_week": a real payday (YYYY-MM-DD)
	Mode        string `json:"mode"`          // biweekly: "every_other_week" | "days_of_month"
	DaysOfMonth []int  `json:"days_of_month"` // biweekly mode "days_of_month": 2 distinct days (1-31)
	DayOfMonth  *int   `json:"day_of_month"`  // monthly / quarterly / semiannually / annually: 1-31
	Month       *int   `json:"month"`         // annually: anchor month (1-12)
}

/**************************************************************************************
* normalizeIncomeRecurRule validates the recurrence details against the frequency and
* returns the canonical rule to store. Days of the month are stored as picked (1-31);
* months shorter than the picked day pay on their last day instead.
**************************************************************************************/
func normalizeIncomeRecurRule(recurring string, raw json.RawMessage) (json.RawMessage, error) {
	var rule incomeRecurRule
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &rule); err != nil {
			return nil, fmt.Errorf("invalid recur_rule: %v", err)
		}
	}

	inRange := func(value, low, high int) bool { return value >= low && value <= high }

	switch recurring {
	case "":
		if !incomeDatePattern.MatchString(rule.Date) {
			return nil, fmt.Errorf("recur_rule.date (YYYY-MM-DD) is required for one-time incomes")
		}
		if _, err := time.Parse("2006-01-02", rule.Date); err != nil {
			return nil, fmt.Errorf("recur_rule.date must be a valid date")
		}
		return json.Marshal(map[string]string{"date": rule.Date})

	case "weekly":
		if rule.DayOfWeek == nil || !inRange(*rule.DayOfWeek, 0, 6) {
			return nil, fmt.Errorf("recur_rule.day_of_week must be between 0 (Sunday) and 6 (Saturday)")
		}
		return json.Marshal(map[string]int{"day_of_week": *rule.DayOfWeek})

	case "biweekly":
		switch rule.Mode {
		case "every_other_week":
			if !incomeDatePattern.MatchString(rule.Anchor) {
				return nil, fmt.Errorf("recur_rule.anchor (YYYY-MM-DD) is required for biweekly incomes")
			}
			if _, err := time.Parse("2006-01-02", rule.Anchor); err != nil {
				return nil, fmt.Errorf("recur_rule.anchor must be a valid date")
			}
			return json.Marshal(map[string]string{"mode": rule.Mode, "anchor": rule.Anchor})
		case "days_of_month":
			if len(rule.DaysOfMonth) != 2 ||
				!inRange(rule.DaysOfMonth[0], 1, 31) || !inRange(rule.DaysOfMonth[1], 1, 31) ||
				rule.DaysOfMonth[0] == rule.DaysOfMonth[1] {
				return nil, fmt.Errorf("recur_rule.days_of_month must contain 2 distinct days (1-31)")
			}
			days := []int{rule.DaysOfMonth[0], rule.DaysOfMonth[1]}
			if days[1] < days[0] {
				days[0], days[1] = days[1], days[0]
			}
			return json.Marshal(map[string]interface{}{"mode": rule.Mode, "days_of_month": days})
		default:
			return nil, fmt.Errorf("recur_rule.mode must be every_other_week or days_of_month for biweekly incomes")
		}

	case "monthly", "quarterly", "semiannually":
		if rule.DayOfMonth == nil || !inRange(*rule.DayOfMonth, 1, 31) {
			return nil, fmt.Errorf("recur_rule.day_of_month must be between 1 and 31")
		}
		return json.Marshal(map[string]int{"day_of_month": *rule.DayOfMonth})

	case "annually":
		if rule.Month == nil || !inRange(*rule.Month, 1, 12) {
			return nil, fmt.Errorf("recur_rule.month must be between 1 and 12")
		}
		if rule.DayOfMonth == nil || !inRange(*rule.DayOfMonth, 1, 31) {
			return nil, fmt.Errorf("recur_rule.day_of_month must be between 1 and 31")
		}
		return json.Marshal(map[string]int{"month": *rule.Month, "day_of_month": *rule.DayOfMonth})
	}

	return nil, fmt.Errorf("invalid recurring value")
}
