package info

import "github.com/jmoiron/sqlx"

type HealthService struct {
	db *sqlx.DB
}

func (s *HealthService) CheckDB() error {
	return s.db.Ping()
}
