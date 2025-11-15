package models

import "time"

type User struct {
	ID        string    `json:"user_id" gorm:"primaryKey;default:gen_random_uuid()"`
	Username  string    `json:"username"`
	IsActive  bool      `json:"is_active"`
	TeamName  string    `json:"team_name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Team struct {
	Name string `json:"team_name"`
}

type TeamWithMembers struct {
	TeamName string `json:"team_name"`
	Members  []User `json:"members"`
}

type PullRequest struct {
	Id        string     `json:"pull_request_id"`
	Name      string     `json:"pull_request_name"`
	AuthorId  string     `json:"author_id"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"-"`
	MergedAt  *time.Time `json:"merged_at,omitempty" gorm:"type:timestamptz"`
}

type UsersReviews struct {
	UserId       string        `json:"user_id"`
	PullRequests []PullRequest `json:"pull_requests"`
}

type PrReviewer struct {
	PullRequestId string `json:"pull_request_id"`
	ReviewerId    string `json:"reviewer_id"`
}

type Review struct {
	PullRequest
	AssignedReviewers []string `json:"assigned_reviewers"`
}

type UpdateReviewer struct {
	PullRequestId string `json:"pull_request_id"`
	OldReviewerId string `json:"old_reviewer_id"`
}

type UpdatedPR struct {
	PullRequest
	AssignedReviewers []string `json:"assigned_reviewers"`
	ReplacedBy        string
}
