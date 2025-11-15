package repository

import "github.com/ashurov-imomali/pr-service/internal/models"

type Repository interface {
	AddTeam(t *models.TeamWithMembers) error
	GetTeam(teamName string) (*models.TeamWithMembers, bool, error)
	UpdateUser(user *models.User) (bool, error)
	GetUsersReview(userID string) (*models.UsersReviews, bool, error)
	GetUserById(id string) (*models.User, bool, error)
	CreatePullRequest(request *models.PullRequest, users []models.PrReviewer) error
	GetUsersIdByTeamName(teamName, authorId string) ([]string, bool, error)
	UpdatePullRequest(pullRequest *models.PullRequest) (bool, error)
	GetUsersIdByReviewId(reviewId string) ([]string, error)
	GetUsersIdByPRId(prId string) ([]string, error)
	GetRandomUser(userId, prID string) (string, bool, error)
	UpdateReviewer(prId, oldReviewerId, newReviewerId string) error
	GetPullRequestById(id string) (*models.PullRequest, bool, error)
	GetReviewListById(prId, userID string) (*models.PrReviewer, bool, error)
	GetUsersStat() ([]models.UsersStat, error)
	GetPRStats() (models.PRStats, error)
	GetUserStat(id string) (*models.UserStat, error)
}
