package employee

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/mock"
)

type MockEmployeeService struct {
	mock.Mock
}

func (m *MockEmployeeService) FindAll(ctx context.Context) ([]Response, error) {
	args := m.Called(ctx)
	return args.Get(0).([]Response), args.Error(1) // Важно: правильный тип
}

func (m *MockEmployeeService) FindAllByIds(ctx context.Context, ids []int64) ([]Response, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]Response), args.Error(1) // Важно: правильный тип
}

func (m *MockEmployeeService) CreateEmployee(ctx context.Context, req CreateRequest) (Response, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockEmployeeService) CreateEmployeeTx(ctx context.Context, request CreateRequest) (int64, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(int64), args.Error(1) // Важно: правильный тип
}

func (m *MockEmployeeService) UpdateEmployee(ctx context.Context, id int64, request UpdateRequest) (Response, error) {
	args := m.Called(ctx, id, request)
	return args.Get(0).(Response), args.Error(1) // Важно: правильный тип
}

func (m *MockEmployeeService) DeleteById(ctx context.Context, id int64) (Response, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(Response), args.Error(1) // Важно: правильный тип
}

func (m *MockEmployeeService) DeleteByIds(ctx context.Context, ids []int64) (Response, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).(Response), args.Error(1) // Важно: правильный тип
}

func (m *MockEmployeeService) FindEmployeeByNameTx(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(bool), args.Error(1) // Важно: правильный тип
}

func (m *MockEmployeeService) CloseTx(tx *sqlx.Tx, err error, s string) {
	//TODO implement me
	panic("implement me")
}

func (m *MockEmployeeService) FindById(ctx context.Context, id int64) (Response, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(Response), args.Error(1) // Важно: правильный тип
}

// Добавьте остальные методы интерфейса
