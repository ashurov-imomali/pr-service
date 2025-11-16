package repository

import "github.com/ashurov-imomali/pr-service/internal/models"

type Repository interface {
	AddTeam(t *models.TeamWithMembers) error
	GetTeam(teamName string) (*models.TeamWithMembers, bool, error)
	UpdateUser(user *models.User) (bool, error)
	GetUsersReview(userID string) (*models.UsersReviews, bool, error)
	GetUserByID(id string) (*models.User, bool, error)
	CreatePullRequest(request *models.PullRequest, users []models.PrReviewer) error
	GetUsersIDByTeamName(teamName, authorID string) ([]string, bool, error)
	UpdatePullRequest(pullRequest *models.PullRequest) (bool, error)
	GetUsersIDByReviewID(reviewID string) ([]string, error)
	GetUsersIDByPRID(prID string) ([]string, error)
	GetRandomUser(userID, prID string) (string, bool, error)
	UpdateReviewer(prID, oldReviewerID, newReviewerID string) error
	GetPullRequestByID(id string) (*models.PullRequest, bool, error)
	GetReviewListByID(prID, userID string) (*models.PrReviewer, bool, error)
	GetUsersStat() ([]models.UsersStat, error)
	GetPRStats() (models.PRStats, error)
	GetUserStat(id string) (*models.UserStat, error)
	DeactivateTeam(teamName string) ([]models.User, bool, error)
	DeleteReviewer(userID, prID string) error
}
