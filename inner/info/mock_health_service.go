package info

import (
	"github.com/stretchr/testify/mock"
)

type MockHealthService struct {
	mock.Mock
}

func (s *MockHealthService) CheckDB() error {
	args := s.Called()
	return args.Error(0)
}
