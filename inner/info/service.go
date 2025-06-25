package info

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
)

type Service struct {
	db *sqlx.DB
}

type Repo interface {
}

// NewService - function constructor
func NewService(db *sqlx.DB) *Service {
	return &Service{db: db}
}

func (s *Service) CheckDB() error {
	if s.db == nil {
		return fmt.Errorf("database connection is not initialized ")
	}
	err := s.db.Ping()
	log.Printf("DB PING ERROR status IS: %v\n", err)
	return err
}
