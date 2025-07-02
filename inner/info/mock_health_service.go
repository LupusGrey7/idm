package info

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockHealthService struct {
	mock.Mock
}

func (s *MockHealthService) CheckDB(ctx context.Context) error {
	args := s.Called()
	return args.Error(0)
}
