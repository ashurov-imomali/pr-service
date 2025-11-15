package repository

import (
	"errors"
	"github.com/ashurov-imomali/pr-service/internal/models"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type repo struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repo{db: db}
}

func (r *repo) AddTeam(t *models.TeamWithMembers) error {
	tx := r.db.Begin()
	team := models.Team{Name: t.TeamName}

	if err := tx.Create(&team).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(t.Members) == 0 {
		tx.Commit()
		return nil
	}

	if err := tx.Save(&t.Members).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (r *repo) GetTeam(teamName string) (*models.TeamWithMembers, bool, error) {
	var result models.TeamWithMembers

	gr := errgroup.Group{}
	gr.Go(func() error {
		var team models.Team
		if err := r.db.First(&team, "name=?", teamName).Error; err != nil {
			return err
		}
		result.TeamName = team.Name
		return nil
	})

	gr.Go(func() error {
		var members []models.User
		if err := r.db.Where("team_name = ?", teamName).Find(&members).Error; err != nil {
			return err
		}
		result.Members = members
		return nil
	})

	if err := gr.Wait(); err != nil {
		return nil, errors.Is(err, gorm.ErrRecordNotFound), err
	}
	return &result, false, nil
}

func (r *repo) UpdateUser(user *models.User) (bool, error) {
	tx := r.db.Model(&user).
		Clauses(clause.Returning{}).
		Update("is_active", user.IsActive)

	if tx.Error != nil {
		return false, tx.Error
	}
	return tx.RowsAffected == 0, nil
}

func (r *repo) GetUsersReview(userID string) (*models.UsersReviews, bool, error) {
	result := models.UsersReviews{
		UserId: userID,
	}

	tx := r.db.Table("pull_requests pr").
		Joins("join pr_reviewers p on pr.id = p.pull_request_id").
		Where("p.reviewer_id=?", userID).Find(&result.PullRequests)

	if tx.Error != nil {
		return nil, false, tx.Error
	}
	return &result, tx.RowsAffected == 0, nil
}

func (r *repo) GetUserById(id string) (*models.User, bool, error) {
	var result models.User
	if err := r.db.First(&result, "id=?", id).Error; err != nil {
		return nil, errors.Is(gorm.ErrRecordNotFound, err), err
	}
	return &result, false, nil
}

func (r *repo) CreatePullRequest(pr *models.PullRequest, reviewers []models.PrReviewer) error {
	tx := r.db.Begin()
	if err := tx.Create(pr).Error; err != nil {
		tx.Rollback()
		return err
	}
	if len(reviewers) == 0 {
		tx.Commit()
		return nil
	}

	if err := tx.Create(&reviewers).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (r *repo) GetUsersIdByTeamName(teamName, authorId string) ([]string, bool, error) {
	var ids []string

	tx := r.db.Select("id").Table("users").
		Where("is_active=? and team_name=? and id != ?", true, teamName, authorId).
		Order("random()").Limit(2).Scan(&ids)
	if tx.Error != nil {
		return nil, false, tx.Error
	}
	return ids, tx.RowsAffected == 0, nil
}

func (r *repo) UpdatePullRequest(pullRequest *models.PullRequest) (bool, error) {
	tx := r.db.Model(&pullRequest).
		Clauses(clause.Returning{}).
		Updates(map[string]interface{}{
			"status":    pullRequest.Status,
			"merged_at": pullRequest.MergedAt})

	if tx.Error != nil {
		return false, tx.Error
	}

	return tx.RowsAffected == 0, nil
}

func (r *repo) GetUsersIdByReviewId(plID string) ([]string, error) {
	var result []string
	return result, r.db.Model(&models.PrReviewer{}).Select("reviewer_id").
		Where("pull_request_id=?", plID).Scan(&result).Error
}

func (r *repo) GetUsersIdByPRId(prId string) ([]string, error) {
	var result []string
	return result, r.db.Select("reviewer_id").Table("pr_reviewers").
		Where("pull_request_id=?", prId).Scan(&result).Error
}

func (r *repo) GetRandomUser(userId, prID string) (string, bool, error) {
	var id string
	tx := r.db.Select("u1.id").Table("users u1").
		Joins("join users u2 on u2.team_name = u1.team_name").
		Joins("left join pr_reviewers pr on u1.id = pr.reviewer_id and pr.pull_request_id=?", prID).
		Where("u2.id=? and u1.id !=? and u1.is_active=? and pr.reviewer_id is null", userId, userId, true).
		Order("random()").Scan(&id)
	if tx.Error != nil {
		return "", false, tx.Error
	}
	return id, tx.RowsAffected == 0, nil
}

func (r *repo) UpdateReviewer(prId, oldReviewerId, newReviewerId string) error {
	return r.db.Model(&models.PrReviewer{}).
		Where("pull_request_id=? and reviewer_id=?", prId, oldReviewerId).
		Update("reviewer_id", newReviewerId).Error
}

func (r *repo) GetPullRequestById(id string) (*models.PullRequest, bool, error) {
	var result models.PullRequest
	if err := r.db.First(&result, "id=?", id).Error; err != nil {
		return nil, errors.Is(err, gorm.ErrRecordNotFound), err
	}
	return &result, false, nil
}

func (r *repo) GetReviewListById(prId, userId string) (*models.PrReviewer, bool, error) {
	var result models.PrReviewer
	if err := r.db.First(&result, "pull_request_id=? and reviewer_id=?", prId, userId).
		Error; err != nil {
		return nil, errors.Is(err, gorm.ErrRecordNotFound), err
	}
	return &result, false, nil
}
