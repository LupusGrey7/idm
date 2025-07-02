package info

import (
	"context"
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

func (s *Service) CheckDB(ctx context.Context) error {
	if s.db == nil {
		return fmt.Errorf("database connection is not initialized ")
	}

	err := s.db.PingContext(ctx) //Контекст контролирует таймауты/отмену операций в gracefulShutdown
	s.logger.Debug(
		"DB PING status IS:",
		zap.Any("error", err),
	)

	return err
}
