package employee

import (
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/mock"
)

type MockEmployeeService struct {
	mock.Mock
}

func (m *MockEmployeeService) FindAll() ([]Response, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockEmployeeService) FindAllByIds(ids []int64) ([]Response, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockEmployeeService) CreateEmployee(req CreateRequest) (Response, error) {
	args := m.Called(req)
	return args.Get(0).(Response), args.Error(1)
}

func (m *MockEmployeeService) CreateEmployeeTx(request CreateRequest) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockEmployeeService) UpdateEmployee(id int64, request UpdateRequest) (Response, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockEmployeeService) DeleteById(id int64) (Response, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockEmployeeService) DeleteByIds(ids []int64) (Response, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockEmployeeService) FindEmployeeByNameTx(name string) (bool, err error) {
	//TODO implement me
	panic("implement me")
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
