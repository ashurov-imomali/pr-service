package usecase

import (
	"github.com/ashurov-imomali/pr-service/internal/models"
	"github.com/ashurov-imomali/pr-service/internal/repository"
	"github.com/ashurov-imomali/pr-service/pkg/logger"
	"net/http"
	"strings"
)

type UserService struct {
	repo repository.Repository
	l    logger.Logger
}

func NewUserService(repo repository.Repository, l logger.Logger) *UserService {
	return &UserService{repo: repo, l: l}
}

func (s *UserService) UpdateUser(user models.User) (*models.User, int, *Error) {
	if len(strings.TrimSpace(user.ID)) == 0 {
		return nil, http.StatusNotFound, &Error{Code: "NOT_FOUND", Message: "resource not found"}
	}

	review, notFound, err := s.repo.GetUsersReview(user.ID)
	if err != nil {
		s.l.Errorf("Error in bd. Err %v", err)
		return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
	}
	if user.IsActive == false && !notFound {
		for _, pr := range review.PullRequests {
			if pr.Status == "MERGED" {
				continue
			}

			randomUser, notFound, err := s.repo.GetRandomUser(user.ID, pr.ID)
			if err != nil {
				s.l.Errorf("Error in bd. Err %v", err)
				return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
			}
			if notFound {
				if err := s.repo.DeleteReviewer(user.ID, pr.ID); err != nil {
					s.l.Errorf("Error in bd. Err %v", err)
					return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
				}
				continue
			}
			if err := s.repo.UpdateReviewer(pr.ID, user.ID, randomUser); err != nil {
				s.l.Errorf("Error in bd. Err %v", err)
				return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
			}
		}
	}

	notFound, err = s.repo.UpdateUser(&user)
	if notFound {
		s.l.Warnf("User not found. userID: %s", user.ID)
		return nil, http.StatusNotFound, &Error{Code: "NOT_FOUND", Message: "resource not found"}
	}
	if err != nil {
		s.l.Errorf("Error in DB. Err:%v", err)
		return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
	}

	return &user, http.StatusOK, nil
}

func (s *UserService) GetUsersReview(userID string) (*models.UsersReviews, int, *Error) {
	reviews, _, err := s.repo.GetUsersReview(userID)
	//if notFound {
	//	return nil, http.StatusNotFound, &Error{Code: "NOT_FOUND", Message: "resource not found"}
	//}
	if err != nil {
		s.l.Errorf("Error get review. Error %v", err)
		return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
	}
	return reviews, http.StatusOK, nil
}
