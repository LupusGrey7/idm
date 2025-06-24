package role

import (
	"github.com/stretchr/testify/mock"
)

type MockRoleService struct {
	mock.Mock
}

func (m *MockRoleService) FindAll() ([]Response, error) {
	args := m.Called()
	return args.Get(0).([]Response), args.Error(1) // Важно: правильный тип
}

func (m *MockRoleService) FindAllByIds(ids []int64) ([]Response, error) {
	args := m.Called(ids)
	return args.Get(0).([]Response), args.Error(1) // Важно: правильный тип
}

func (m *MockRoleService) FindById(id int64) (Response, error) {
	args := m.Called(id)
	return args.Get(0).(Response), args.Error(1) // Важно: правильный тип
}

func (m *MockRoleService) CreateRole(req CreateRequest) (Response, error) {
	args := m.Called(req)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockRoleService) UpdateRole(id int64, request UpdateRequest) (Response, error) {
	args := m.Called(id, request)
	return args.Get(0).(Response), args.Error(1) // Важно: правильный тип
}

func (m *MockRoleService) DeleteById(id int64) (Response, error) {
	args := m.Called(id)
	return args.Get(0).(Response), args.Error(1) // Важно: правильный тип
}

func (m *MockRoleService) DeleteByIds(ids []int64) (Response, error) {
	args := m.Called(ids)
	return args.Get(0).(Response), args.Error(1) // Важно: правильный тип
}

// Добавьте остальные методы интерфейса
