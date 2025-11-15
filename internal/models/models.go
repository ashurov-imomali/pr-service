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
	ID        string     `json:"pull_request_id"`
	Name      string     `json:"pull_request_name"`
	AuthorID  string     `json:"author_id"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"-"`
	MergedAt  *time.Time `json:"merged_at,omitempty" gorm:"type:timestamptz"`
}

type UsersReviews struct {
	UserID       string        `json:"user_id"`
	PullRequests []PullRequest `json:"pull_requests"`
}

type PrReviewer struct {
	PullRequestID string `json:"pull_request_id"`
	ReviewerID    string `json:"reviewer_id"`
}

type Review struct {
	PullRequest
	AssignedReviewers []string `json:"assigned_reviewers"`
}

type UpdateReviewer struct {
	PullRequestID string `json:"pull_request_id"`
	OldReviewerID string `json:"old_reviewer_id"`
}

type UpdatedPR struct {
	PullRequest
	AssignedReviewers []string `json:"assigned_reviewers"`
	ReplacedBy        string   `json:"replaced_by"`
}

type GeneralStats struct {
	UsersStat []UsersStat `json:"users_stat"`
	PRStats   PRStats     `json:"pr_stats"`
}

type UsersStat struct {
	UserID  string `json:"user_id"`
	PRCount int64  `json:"pr_count"`
}

type PRStats struct {
	Total int64 `json:"total"`
	Open  int64 `json:"open"`
	Merge int64 `json:"merge"`
}

type UserStat struct {
	User
	PullRequestsCount  int64 `json:"pull_requests_count"`
	ReviewsCount       int64 `json:"reviews_count"`
	MergedReviewsCount int64 `json:"merged_reviews_count"`
	OpenReviewsCount   int64 `json:"open_reviews_count"`
}
