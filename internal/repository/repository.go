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
		UserID: userID,
	}

	tx := r.db.Table("pull_requests pr").
		Joins("join pr_reviewers p on pr.id = p.pull_request_id").
		Where("p.reviewer_id=?", userID).Find(&result.PullRequests)

	if tx.Error != nil {
		return nil, false, tx.Error
	}
	return &result, tx.RowsAffected == 0, nil
}

func (r *repo) GetUserByID(id string) (*models.User, bool, error) {
	var result models.User
	if err := r.db.First(&result, "id=?", id).Error; err != nil {
		return nil, errors.Is(err, gorm.ErrRecordNotFound), err
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

func (r *repo) GetUsersIDByTeamName(teamName, authorID string) ([]string, bool, error) {
	var ids []string

	tx := r.db.Select("id").Table("users").
		Where("is_active=? and team_name=? and id != ?", true, teamName, authorID).
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

func (r *repo) GetUsersIDByReviewID(prID string) ([]string, error) {
	var result []string
	return result, r.db.Model(&models.PrReviewer{}).Select("reviewer_id").
		Where("pull_request_id=?", prID).Scan(&result).Error
}

func (r *repo) GetUsersIDByPRID(prID string) ([]string, error) {
	var result []string
	return result, r.db.Select("reviewer_id").Table("pr_reviewers").
		Where("pull_request_id=?", prID).Scan(&result).Error
}

func (r *repo) GetRandomUser(userID, prID string) (string, bool, error) {
	var id string
	//SELECT u1.id, u2.id, p.id, pr.* FROM users u1
	//join users u2 on u1.team_name = u2.team_name and u2.id = 'u-2'
	//left join pull_requests p on u1.id = p.author_id
	//left join pr_reviewers pr on u1.id = pr.reviewer_id and pr.pull_request_id='tpr-2'
	//where u1.is_active=true and pr.reviewer_id is null and p.id is null
	//ORDER BY random();
	tx := r.db.Select("u1.id").Table("users u1").
		Joins("join users u2 on u1.team_name=u2.team_name and u2.id=?", userID).                     // беру тех кто из команды
		Joins("left join pull_requests p on u1.id = p.author_id").                                   // исключаю автора
		Joins("left join pr_reviewers pr on u1.id = pr.reviewer_id and pr.pull_request_id=?", prID). //исключаю тех кто уже как ревьюЕО стоят
		Where("u1.is_active=? and pr.reviewer_id is null and p.id is null", true).
		Order("random()").Scan(&id)
	if tx.Error != nil {
		return "", false, tx.Error
	}
	return id, tx.RowsAffected == 0, nil
}

func (r *repo) UpdateReviewer(prID, oldReviewerID, newReviewerID string) error {
	return r.db.Model(&models.PrReviewer{}).
		Where("pull_request_id=? and reviewer_id=?", prID, oldReviewerID).
		Update("reviewer_id", newReviewerID).Error
}

func (r *repo) GetPullRequestByID(id string) (*models.PullRequest, bool, error) {
	var result models.PullRequest
	if err := r.db.First(&result, "id=?", id).Error; err != nil {
		return nil, errors.Is(err, gorm.ErrRecordNotFound), err
	}
	return &result, false, nil
}

func (r *repo) GetReviewListByID(prID, userID string) (*models.PrReviewer, bool, error) {
	var result models.PrReviewer
	if err := r.db.First(&result, "pull_request_id=? and reviewer_id=?", prID, userID).
		Error; err != nil {
		return nil, errors.Is(err, gorm.ErrRecordNotFound), err
	}
	return &result, false, nil
}

func (r *repo) GetUsersStat() ([]models.UsersStat, error) {
	var stat []models.UsersStat
	return stat, r.db.Select("users.id user_id, count(pr.*) pr_count").Model(&models.User{}).
		Joins("left join pr_reviewers pr on users.id = pr.reviewer_id").
		Group("users.id").Find(&stat).Error
}

func (r *repo) GetPRStats() (models.PRStats, error) {
	var result models.PRStats
	return result, r.db.Select(`count(*) total,
count(*) filter(where status = 'OPEN') open,
count(*) filter(where status = 'MERGE') merged
`).Model(&models.PullRequest{}).Find(&result).Error
}

func (r *repo) GetUserStat(id string) (*models.UserStat, error) {
	//select u.id,
	//       count(pr.*),
	//       count(p.*),
	//       count(p.*) filter (where pr1.status = 'MERGED'),
	//       count(p.*) filter ( where pr1.status = 'OPEN')
	//from users u
	//left join pull_requests pr on u.id = pr.author_id
	//left join pr_reviewers p on p.reviewer_id = u.id
	//left join pull_requests pr1 on pr1.id = p.pull_request_id
	//group by u.id;
	var result models.UserStat
	return &result, r.db.Select(`u.*,
	       count(pr.*) pull_requests_count,
	       count(p.*) reviews_count,
	       count(p.*) filter (where pr1.status = 'MERGED') merged_reviews_count,
	       count(p.*) filter ( where pr1.status = 'OPEN') open_reviews_count`).
		Table("users u").
		Joins("left join pull_requests pr on u.id = pr.author_id").
		Joins("left join pr_reviewers p on p.reviewer_id = u.id").
		Joins("left join pull_requests pr1 on pr1.id = p.pull_request_id").
		Where("u.id=?", id).
		Group("u.id").Find(&result).Error
}

func (r *repo) DeactivateTeam(teamName string) ([]models.User, bool, error) {
	var result []models.User
	tx := r.db.Begin()

	if err := tx.Model(&result).
		Clauses(clause.Returning{}).
		Where("team_name=?", teamName).
		Update("is_active", false).Error; err != nil {
		tx.Rollback()
		return nil, false, err
	}

	for _, user := range result {
		if err := tx.Exec(`
    delete from pr_reviewers p
    using pull_requests pr
    where p.pull_request_id = pr.id
      and pr.status = 'OPEN'
      and p.reviewer_id = ?;
`, user.ID).Error; err != nil {
			tx.Rollback()
			return nil, false, err
		}

	}

	tx.Commit()
	return result, len(result) == 0, nil
}

func (r *repo) DeleteReviewer(userID, prID string) error {
	return r.db.Delete(models.PrReviewer{}, "reviewer_id=? and pull_request_id=?", userID, prID).Error
}
