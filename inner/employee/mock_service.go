package employee

import (
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/mock"
)

type MockEmployeeService struct {
	mock.Mock
}

func (m *MockEmployeeService) FindAll() ([]Response, error) {
	args := m.Called()
	return args.Get(0).([]Response), args.Error(1) // Важно: правильный тип
}

func (m *MockEmployeeService) FindAllByIds(ids []int64) ([]Response, error) {
	args := m.Called(ids)
	return args.Get(0).([]Response), args.Error(1) // Важно: правильный тип
}

func (m *MockEmployeeService) CreateEmployee(req CreateRequest) (Response, error) {
	args := m.Called(req)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockEmployeeService) CreateEmployeeTx(request CreateRequest) (int64, error) {
	args := m.Called(request)
	return args.Get(0).(int64), args.Error(1) // Важно: правильный тип
}

func (m *MockEmployeeService) UpdateEmployee(id int64, request UpdateRequest) (Response, error) {
	args := m.Called(id, request)
	return args.Get(0).(Response), args.Error(1) // Важно: правильный тип
}

func (m *MockEmployeeService) DeleteById(id int64) (Response, error) {
	args := m.Called(id)
	return args.Get(0).(Response), args.Error(1) // Важно: правильный тип
}

func (m *MockEmployeeService) DeleteByIds(ids []int64) (Response, error) {
	args := m.Called(ids)
	return args.Get(0).(Response), args.Error(1) // Важно: правильный тип
}

func (m *MockEmployeeService) FindEmployeeByNameTx(name string) (bool, error) {
	args := m.Called(name)
	return args.Get(0).(bool), args.Error(1) // Важно: правильный тип
}

func (m *MockEmployeeService) CloseTx(tx *sqlx.Tx, err error, s string) {
	//TODO implement me
	panic("implement me")
}

func (m *MockEmployeeService) FindById(id int64) (Response, error) {
	args := m.Called(id)
	return args.Get(0).(Response), args.Error(1) // Важно: правильный тип
}

// Добавьте остальные методы интерфейса
