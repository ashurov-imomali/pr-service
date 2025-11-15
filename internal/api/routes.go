package api

import "net/http"

func (h *Handler) RegisterRouters(mux *http.ServeMux) {
	{
		mux.HandleFunc("/team/add", h.addTeam) //done
		mux.HandleFunc("/team/get", h.getTeam) //done
	}

	{
		mux.HandleFunc("/users/setIsActive", h.setUser)
		mux.HandleFunc("/users/getReview", h.getUserReviews) //todo all time 200 ???
	}

	{
		mux.HandleFunc("/pullRequests/create", h.createPullRequest)
		mux.HandleFunc("/pullRequests/merge", h.mergePullRequest)
		mux.HandleFunc("/pullRequests/reassign", h.reassignPullRequest)
	}
}
