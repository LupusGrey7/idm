package info

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"idm/inner/common"
)

type Service struct {
	db     *sqlx.DB
	logger *common.Logger
}

type Repo interface {
}

// NewService - function constructor
func NewService(
	db *sqlx.DB,
	logger *common.Logger,
) *Service {
	return &Service{
		db:     db,
		logger: logger,
	}
}

func (s *Service) CheckDB() error {
	if s.db == nil {
		return fmt.Errorf("database connection is not initialized ")
	}
	err := s.db.Ping()
	s.logger.Debug("DB PING status IS: %s", zap.Error(err))

	return err
}
