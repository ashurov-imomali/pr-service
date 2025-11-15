package usecase

import (
	"github.com/ashurov-imomali/pr-service/internal/models"
	"github.com/ashurov-imomali/pr-service/internal/repository"
	"github.com/ashurov-imomali/pr-service/pkg/logger"
	"golang.org/x/sync/errgroup"
)

type StatService struct {
	repo repository.Repository
	l    logger.Logger
}

func NewStatService(repo repository.Repository, l logger.Logger) *StatService {
	return &StatService{repo: repo, l: l}
}

func (s *StatService) GetGeneralStat() (*models.GeneralStats, *Error) {
	var result models.GeneralStats

	gr := errgroup.Group{}

	gr.Go(func() error {
		stats, err := s.repo.GetUsersStat()
		if err != nil {
			s.l.Errorf("Error in bd. Err %v", err)
			return err

		}
		result.UsersStat = stats
		return nil
	})

	gr.Go(func() error {
		stats, err := s.repo.GetPRStats()
		if err != nil {
			s.l.Errorf("Error in bd. Err %v", err)
			return err

		}
		result.PRStats = stats
		return nil
	})

	if err := gr.Wait(); err != nil {
		return nil, &Error{Code: "INTERNAL_SERVER_ERROR"}
	}

	return &result, nil
}

func (s *StatService) GetUserStat(id string) (*models.UserStat, *Error) {
	stat, err := s.repo.GetUserStat(id)
	if err != nil {
		return nil, &Error{Code: "INTERNAL_SERVER_ERROR"}
	}
	return stat, nil
}
