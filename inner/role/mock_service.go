package role

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockRoleService struct {
	mock.Mock
}

func (m *MockRoleService) FindAll(ctx context.Context) ([]Response, error) {
	args := m.Called(ctx)
	return args.Get(0).([]Response), args.Error(1) // Важно: правильный тип
}

func (m *MockRoleService) FindAllByIds(ctx context.Context, ids []int64) ([]Response, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]Response), args.Error(1) // Важно: правильный тип
}

func (m *MockRoleService) FindById(ctx context.Context, id int64) (Response, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(Response), args.Error(1) // Важно: правильный тип
}

func (m *MockRoleService) CreateRole(ctx context.Context, req CreateRequest) (Response, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockRoleService) UpdateRole(ctx context.Context, id int64, request UpdateRequest) (Response, error) {
	args := m.Called(ctx, id, request)
	return args.Get(0).(Response), args.Error(1) // Важно: правильный тип
}

func (m *MockRoleService) DeleteById(ctx context.Context, id int64) (Response, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(Response), args.Error(1) // Важно: правильный тип
}

func (m *MockRoleService) DeleteByIds(ctx context.Context, ids []int64) (Response, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).(Response), args.Error(1) // Важно: правильный тип
}

// Добавьте остальные методы интерфейса
