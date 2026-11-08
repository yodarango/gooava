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

/************************************************************************
* Lists the authenticated user's savings milestones with live progress.
************************************************************************/
func ListMilestones(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	milestones, err := models.ListMilestones(authUser.Id)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = milestones
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Creates a savings milestone and automatically breaks down the contribution
* per paycheck across upcoming paydays until the target date.
************************************************************************/
func CreateMilestone(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	var req struct {
		Title        string  `json:"title"`
		TargetAmount float64 `json:"target_amount"`
		TargetDate   string  `json:"target_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpResponse.Error = "Invalid request format"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		httpResponse.Error = "title is required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}
	if req.TargetAmount <= 0 {
		httpResponse.Error = "target_amount must be greater than 0"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}
	if _, err := time.Parse("2006-01-02", req.TargetDate); err != nil {
		httpResponse.Error = "target_date must be YYYY-MM-DD"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	milestone, err := models.CreateMilestone(authUser.Id, title, req.TargetAmount, req.TargetDate)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = milestone
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Updates a savings milestone and re-synchronizes its schedule.
************************************************************************/
func UpdateMilestone(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	var req struct {
		Id           int64   `json:"id"`
		Title        string  `json:"title"`
		TargetAmount float64 `json:"target_amount"`
		TargetDate   string  `json:"target_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Id <= 0 {
		httpResponse.Error = "a valid id is required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		httpResponse.Error = "title is required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}
	if req.TargetAmount <= 0 {
		httpResponse.Error = "target_amount must be greater than 0"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}
	if _, err := time.Parse("2006-01-02", req.TargetDate); err != nil {
		httpResponse.Error = "target_date must be YYYY-MM-DD"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	milestone, err := models.UpdateMilestone(authUser.Id, req.Id, title, req.TargetAmount, req.TargetDate)
	if err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = milestone
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}

/************************************************************************
* Deletes a savings milestone and cascades its linked expense and occurrences.
************************************************************************/
func DeleteMilestone(w http.ResponseWriter, r *http.Request) {
	var httpResponse models.HttpResponse

	authUser, ok := r.Context().Value(constants.USER_CONTEXT_AUTH_KEY).(*models.AuthUser)
	if !ok {
		httpResponse.Error = "Authentication required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	var req struct {
		Id int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Id <= 0 {
		httpResponse.Error = "a valid id is required"
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	if err := models.DeleteMilestone(authUser.Id, req.Id); err != nil {
		httpResponse.Error = fmt.Sprintf("%v", err)
		httpResponse.Success = false
		httpResponse.Data = nil
		httpResponse.Send(w)
		return
	}

	httpResponse.Data = map[string]string{"message": "Milestone deleted"}
	httpResponse.Success = true
	httpResponse.Error = nil
	httpResponse.Send(w)
}
