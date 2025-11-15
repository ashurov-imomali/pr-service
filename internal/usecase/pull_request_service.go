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

	user, notFound, err := s.repo.GetUserById(pr.AuthorId) //не активный пользователь
	if notFound {
		s.l.Warnf("user not found. ID: %s", pr.AuthorId)
		return nil, http.StatusNotFound, &Error{Code: "NOT_FOUND", Message: "resource not found"}
	}

	if err != nil {
		s.l.Errorf("Error in DB (get user). Error %v", err)
		return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
	}

	randUsers, _, err := s.repo.GetUsersIdByTeamName(user.TeamName, pr.AuthorId)
	if err != nil {
		s.l.Errorf("Error in BD (get user by teamnam). err %v", err)
		return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
	}

	reviewers := make([]models.PrReviewer, len(randUsers))
	for i := 0; i < len(randUsers); i++ {
		reviewers[i].PullRequestId = pr.Id
		reviewers[i].ReviewerId = randUsers[i]
	}

	pr.Status = "OPEN"
	if err := s.repo.CreatePullRequest(&pr, reviewers); err != nil {
		s.l.Errorf("Error in BD (create PR). Err %v", err)
		return nil, http.StatusConflict, &Error{Code: "PR_EXISTS", Message: "PR id already exists"}
	}

	return &models.Review{PullRequest: pr, AssignedReviewers: randUsers}, http.StatusCreated, nil
}

func (s *PRService) MergePullRequest(pullRequestId string) (*models.Review, int, *Error) {
	now := time.Now()
	pr := models.PullRequest{
		Id:       pullRequestId,
		Status:   "MERGED",
		MergedAt: &now,
	}

	notFound, err := s.repo.UpdatePullRequest(&pr)
	if notFound {
		s.l.Warnf("PullRequest not found. id: %s", pullRequestId)
		return nil, http.StatusNotFound, &Error{Code: "NOT_FOUND", Message: "resource not found"}
	}
	if err != nil {
		s.l.Errorf("Err in bd (update PR). Err: %v", err)
		return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
	}

	ids, err := s.repo.GetUsersIdByReviewId(pr.Id)
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

	newReviewer, notFound, err := s.repo.GetRandomUser(pr.AuthorId, review.PullRequestId)
	if notFound {
		s.l.Warnf("No candidate %+v", review)
		return nil, http.StatusConflict, &Error{Code: "NO_CANDIDATE", Message: "no active replacement candidate in team"}
	}

	if err != nil {
		s.l.Errorf("Error in bd (get random user). Err %v", err)
		return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
	}

	if err := s.repo.UpdateReviewer(review.PullRequestId, review.OldReviewerId, newReviewer); err != nil {
		s.l.Errorf("Error in bd (update review). Err %v", err)
		return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
	}

	assignedReviewers, err := s.repo.GetUsersIdByPRId(pr.Id)
	if err != nil {
		s.l.Errorf("Error in bd (update review). Err %v", err)
		return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
	}

	return &models.UpdatedPR{PullRequest: *pr,
		AssignedReviewers: assignedReviewers,
		ReplacedBy:        newReviewer}, http.StatusOK, nil
}

func (s *PRService) validateReassign(review models.UpdateReviewer) (*models.PullRequest, int, *Error) {
	pr, notFound, err := s.repo.GetPullRequestById(review.PullRequestId)
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

	_, notFound, err = s.repo.GetReviewListById(review.PullRequestId, review.OldReviewerId)
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
