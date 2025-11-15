package usecase

import (
	"github.com/ashurov-imomali/pr-service/internal/models"
	"github.com/ashurov-imomali/pr-service/internal/repository"
	"github.com/ashurov-imomali/pr-service/pkg/logger"
	"net/http"
	"strings"
	"time"
)

type PRService struct {
	repo repository.Repository
	l    logger.Logger
}

func NewPRService(repo repository.Repository, l logger.Logger) *PRService {
	return &PRService{repo: repo, l: l}
}

func (s *PRService) CreatePullRequest(pr models.PullRequest) (*models.Review, int, *Error) {
	if len(strings.TrimSpace(pr.Name)) == 0 {
		return nil, http.StatusUnprocessableEntity, &Error{Code: "INVALID_PULL_REQUEST_NAME"}
	}

	user, notFound, err := s.repo.GetUserByID(pr.AuthorID) //не активный пользователь
	if notFound {
		s.l.Warnf("user not found. ID: %s", pr.AuthorID)
		return nil, http.StatusNotFound, &Error{Code: "NOT_FOUND", Message: "resource not found"}
	}

	if err != nil {
		s.l.Errorf("Error in DB (get user). Error %v", err)
		return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
	}

	randUsers, _, err := s.repo.GetUsersIDByTeamName(user.TeamName, pr.AuthorID)
	if err != nil {
		s.l.Errorf("Error in BD (get user by teamnam). err %v", err)
		return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
	}

	reviewers := make([]models.PrReviewer, len(randUsers))
	for i := 0; i < len(randUsers); i++ {
		reviewers[i].PullRequestID = pr.ID
		reviewers[i].ReviewerID = randUsers[i]
	}

	pr.Status = "OPEN"
	if err := s.repo.CreatePullRequest(&pr, reviewers); err != nil {
		s.l.Errorf("Error in BD (create PR). Err %v", err)
		return nil, http.StatusConflict, &Error{Code: "PR_EXISTS", Message: "PR id already exists"}
	}

	return &models.Review{PullRequest: pr, AssignedReviewers: randUsers}, http.StatusCreated, nil
}

func (s *PRService) MergePullRequest(pullRequestID string) (*models.Review, int, *Error) {
	now := time.Now()
	pr := models.PullRequest{
		ID:       pullRequestID,
		Status:   "MERGED",
		MergedAt: &now,
	}

	notFound, err := s.repo.UpdatePullRequest(&pr)
	if notFound {
		s.l.Warnf("PullRequest not found. id: %s", pullRequestID)
		return nil, http.StatusNotFound, &Error{Code: "NOT_FOUND", Message: "resource not found"}
	}
	if err != nil {
		s.l.Errorf("Err in bd (update PR). Err: %v", err)
		return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
	}

	ids, err := s.repo.GetUsersIDByReviewID(pr.ID)
	if err != nil {
		s.l.Errorf("Err in bd (get reviewers). Err %v", err)
		return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
	}

	return &models.Review{PullRequest: pr, AssignedReviewers: ids}, http.StatusOK, nil
}

func (s *PRService) UpdateReviewer(review models.UpdateReviewer) (*models.UpdatedPR, int, *Error) {
	pr, status, wErr := s.validateReassign(review)
	if wErr != nil {
		return nil, status, wErr
	}

	newReviewer, notFound, err := s.repo.GetRandomUser(review.OldReviewerID, review.PullRequestID)
	if notFound {
		s.l.Warnf("No candidate %+v", review)
		return nil, http.StatusConflict, &Error{Code: "NO_CANDIDATE", Message: "no active replacement candidate in team"}
	}

	if err != nil {
		s.l.Errorf("Error in bd (get random user). Err %v", err)
		return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
	}

	if err := s.repo.UpdateReviewer(review.PullRequestID, review.OldReviewerID, newReviewer); err != nil {
		s.l.Errorf("Error in bd (update review). Err %v", err)
		return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
	}

	assignedReviewers, err := s.repo.GetUsersIDByPRID(pr.ID)
	if err != nil {
		s.l.Errorf("Error in bd (update review). Err %v", err)
		return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
	}

	return &models.UpdatedPR{PullRequest: *pr,
		AssignedReviewers: assignedReviewers,
		ReplacedBy:        newReviewer}, http.StatusOK, nil
}

func (s *PRService) validateReassign(review models.UpdateReviewer) (*models.PullRequest, int, *Error) {
	pr, notFound, err := s.repo.GetPullRequestByID(review.PullRequestID)
	switch {
	case notFound:
		s.l.Warnf("Not found. request: %+v", review)
		return nil, http.StatusConflict, &Error{Code: "NOT_FOUND", Message: "resource not found"}
	case err != nil:
		s.l.Errorf("Error in bd (get pr by id). Err %v", err)
		return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
	case pr.Status == "MERGED":
		return nil, http.StatusConflict, &Error{Code: "PR_MERGED", Message: "cannot reassign on merged PR"}

	}

	_, notFound, err = s.repo.GetReviewListByID(review.PullRequestID, review.OldReviewerID)
	if notFound {
		s.l.Warnf("Not found. request: %+v", review)
		return nil, http.StatusConflict, &Error{Code: "NOT_ASSIGNED", Message: "reviewer is not assigned to this PR"}
	}
	if err != nil {
		s.l.Errorf("Error in bd (get review by id). Err %v", err)
		return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
	}

	return pr, http.StatusOK, nil
}
