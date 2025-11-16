package api

import (
	"encoding/json"
	"github.com/ashurov-imomali/pr-service/internal/models"
	"github.com/ashurov-imomali/pr-service/internal/usecase"
	"net/http"
	"strings"
)

type Handler struct {
	prs *usecase.PRService
	us  *usecase.UserService
	ts  *usecase.TeamService
	ss  *usecase.StatService
}

func New(prs *usecase.PRService, us *usecase.UserService, ts *usecase.TeamService, ss *usecase.StatService) *Handler {
	return &Handler{
		prs: prs,
		us:  us,
		ts:  ts,
		ss:  ss,
	}
}

func (h *Handler) addTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var team models.TeamWithMembers
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		writeError(w, "INVALID_JSON")
		return
	}

	status, wErr := h.ts.AddTeam(&team)
	if wErr != nil {
		writeJSON(w, status, wErr)
		return
	}

	writeJSON(w, status, map[string]interface{}{"team": team})
}

func (h *Handler) getTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		writeError(w, "invalid team name")
		return
	}

	team, status, err := h.ts.GetTeam(teamName)
	if err != nil {
		writeJSON(w, status, err)
		return
	}
	writeJSON(w, status, team)
}

func (h *Handler) setUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeError(w, "INVALID_JSON")
		return
	}

	updatedUser, status, wErr := h.us.UpdateUser(user)
	if wErr != nil {
		writeJSON(w, status, wErr)
		return
	}
	writeJSON(w, status, updatedUser)
}

func (h *Handler) getUserReviews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if len(strings.TrimSpace(userID)) == 0 {
		writeError(w, "invalid userID")
		return
	}

	reviews, status, err := h.us.GetUsersReview(userID)
	if err != nil {
		writeJSON(w, status, err)
		return
	}

	writeJSON(w, status, reviews)
}

func (h *Handler) createPullRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var pullRequest models.PullRequest
	if err := json.NewDecoder(r.Body).Decode(&pullRequest); err != nil {
		writeError(w, "INVALID_JSON")
		return
	}

	review, status, err := h.prs.CreatePullRequest(pullRequest)
	if err != nil {
		writeJSON(w, status, err)
		return
	}

	writeJSON(w, status, map[string]interface{}{"pr": review})

}

func (h *Handler) mergePullRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var pullRequest models.PullRequest
	if err := json.NewDecoder(r.Body).Decode(&pullRequest); err != nil {
		writeError(w, "INVALID_JSON")
		return
	}
	review, status, wErr := h.prs.MergePullRequest(pullRequest.ID)
	if wErr != nil {
		writeJSON(w, status, wErr)
		return
	}
	writeJSON(w, status, map[string]interface{}{"pr": review})
}

func (h *Handler) reassignPullRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var newReviewer models.UpdateReviewer
	if err := json.NewDecoder(r.Body).Decode(&newReviewer); err != nil {
		writeError(w, "INVALID_JSON")
		return
	}

	result, status, err := h.prs.UpdateReviewer(newReviewer)
	if err != nil {
		writeJSON(w, status, err)
		return
	}

	writeJSON(w, status, result)
}

func (h *Handler) getGeneralStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	stat, err := h.ss.GetGeneralStat()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, stat)
}

func (h *Handler) getUsersStat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeError(w, "invalid user id")
		return
	}

	stat, err := h.ss.GetUserStat(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, stat)
}

func (h *Handler) deactivateTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var team models.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		writeError(w, "invalid json")
		return
	}
	deactiveUsers, status, err := h.ts.DeactivateTeam(team.Name)
	if err != nil {
		writeJSON(w, status, err)
		return
	}

	writeJSON(w, status, deactiveUsers)
}
