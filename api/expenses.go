package api

import (
	"encoding/json"
	"fmt"
	"gooava/constants"
	"gooava/internal/models"
	"net/http"
	"strings"
)

/************************************************************************
* Creates a manually entered expense for the authenticated user.
*
* status: ✅
************************************************************************/
func CreateExpense(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	var expense models.Expense
	if err := json.NewDecoder(r.Body).Decode(&expense); err != nil {
		httpResponse.Error = "Invalid request format"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	expense.Label = strings.TrimSpace(expense.Label)
	if expense.Label == "" {
		httpResponse.Error = "label is required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if expense.Amount <= 0 {
		httpResponse.Error = "amount must be greater than 0"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if !allowedRecurringFrequencies[expense.Recurring] {
		httpResponse.Error = "invalid recurring value"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if err := validateRecurRule(expense.Recurring, expense.RecurRule); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	// daily and one-time expenses carry no rule
	if expense.Recurring == "" || expense.Recurring == "daily" {
		expense.RecurRule = nil
	}

	expense.UserID = authUser.Id

	if err := expense.Create(); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]string{
		"message": "Expense added",
	}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

var allowedRecurringFrequencies = map[string]bool{
	"": true, "daily": true, "weekly": true, "biweekly": true,
	"monthly": true, "quarterly": true, "yearly": true,
}

// recurRule mirrors the JSON shapes the frontend sends per frequency.
type recurRule struct {
	DayOfWeek   *int  `json:"day_of_week"`   // weekly: 0 (Sunday) - 6 (Saturday)
	DaysOfMonth []int `json:"days_of_month"` // biweekly: 2 distinct days (1-31)
	DayOfMonth  *int  `json:"day_of_month"`  // monthly: 1-31
	Day         *int  `json:"day"`           // quarterly/yearly: day of month (1-31)
	Month       *int  `json:"month"`         // yearly: 1-12
	Months      []int `json:"months"`        // quarterly: 4 distinct months (1-12)
}

/**************************************************************************************
* validateRecurRule checks that the recurrence details match the chosen frequency.
* daily and one-time ("") expenses need no rule.
**************************************************************************************/
func validateRecurRule(recurring string, raw json.RawMessage) error {
	if recurring == "" || recurring == "daily" {
		return nil
	}

	if len(raw) == 0 {
		return fmt.Errorf("recur_rule is required when recurring is %s", recurring)
	}

	var rule recurRule
	if err := json.Unmarshal(raw, &rule); err != nil {
		return fmt.Errorf("invalid recur_rule: %v", err)
	}

	inRange := func(value, low, high int) bool { return value >= low && value <= high }

	switch recurring {
	case "weekly":
		if rule.DayOfWeek == nil || !inRange(*rule.DayOfWeek, 0, 6) {
			return fmt.Errorf("recur_rule.day_of_week must be between 0 (Sunday) and 6 (Saturday)")
		}
	case "biweekly":
		if len(rule.DaysOfMonth) != 2 ||
			!inRange(rule.DaysOfMonth[0], 1, 31) || !inRange(rule.DaysOfMonth[1], 1, 31) ||
			rule.DaysOfMonth[0] == rule.DaysOfMonth[1] {
			return fmt.Errorf("recur_rule.days_of_month must contain 2 distinct days (1-31)")
		}
	case "monthly":
		if rule.DayOfMonth == nil || !inRange(*rule.DayOfMonth, 1, 31) {
			return fmt.Errorf("recur_rule.day_of_month must be between 1 and 31")
		}
	case "quarterly":
		if rule.Day == nil || !inRange(*rule.Day, 1, 31) {
			return fmt.Errorf("recur_rule.day must be between 1 and 31")
		}
		if len(rule.Months) != 4 {
			return fmt.Errorf("recur_rule.months must contain 4 months")
		}
		seen := make(map[int]bool)
		for _, month := range rule.Months {
			if !inRange(month, 1, 12) || seen[month] {
				return fmt.Errorf("recur_rule.months must contain 4 distinct months (1-12)")
			}
			seen[month] = true
		}
	case "yearly":
		if rule.Month == nil || !inRange(*rule.Month, 1, 12) || rule.Day == nil || !inRange(*rule.Day, 1, 31) {
			return fmt.Errorf("recur_rule.month (1-12) and recur_rule.day (1-31) are required")
		}
	}

	return nil
}
