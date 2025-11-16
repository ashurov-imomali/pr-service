package usecase

import (
	"github.com/ashurov-imomali/pr-service/internal/models"
	"github.com/ashurov-imomali/pr-service/internal/repository"
	"github.com/ashurov-imomali/pr-service/pkg/logger"
	"net/http"
	"strings"
)

type TeamService struct {
	repo repository.Repository
	l    logger.Logger
}

func NewTeamService(repo repository.Repository, l logger.Logger) *TeamService {
	return &TeamService{repo: repo, l: l}
}

func (s *TeamService) AddTeam(twm *models.TeamWithMembers) (int, *Error) {
	if len(strings.TrimSpace(twm.TeamName)) == 0 {
		s.l.Warnf("incorrect team name: %s", twm.TeamName)
		return http.StatusUnprocessableEntity, &Error{Code: "INVALID_TEAM_NAME"}
	}

	for i := 0; i < len(twm.Members); i++ {
		twm.Members[i].TeamName = twm.TeamName
	}

	err := s.repo.AddTeam(twm)
	if err != nil {
		s.l.Errorf("Current team name:%s or transaction error:%v", twm.TeamName, err)
		return http.StatusBadRequest, &Error{Code: "TEAM_EXISTS", Message: "team_name already exists"}
	}

	return http.StatusCreated, nil
}

func (s *TeamService) GetTeam(teamName string) (*models.TeamWithMembers, int, *Error) {
	result, notFound, err := s.repo.GetTeam(teamName)
	if notFound {
		s.l.Warnf("Team not found. teamName:%s", teamName)
		return nil, http.StatusNotFound, &Error{Code: "NOT_FOUND", Message: "resource not found"}
	}

	if err != nil {
		s.l.Errorf("Error in DB (get team). Error: %v", err)
		return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
	}

	return result, http.StatusOK, nil
}

func (s *TeamService) DeactivateTeam(teamName string) ([]models.User, int, *Error) {
	deactiveUsers, notFound, err := s.repo.DeactivateTeam(teamName)
	if notFound {
		return nil, http.StatusNotFound, &Error{Code: "TEAM_NOT_FOUND"}
	}
	if err != nil {
		s.l.Errorf("Err in bd. Err %v", err)
		return nil, http.StatusInternalServerError, &Error{Code: "INTERNAL_SERVER_ERROR"}
	}

	return deactiveUsers, http.StatusOK, nil
}
