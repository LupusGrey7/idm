package service

import (
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"idm/inner/employee"
	"idm/tests/unit/mocks"
	"testing"
	"time"
)

// StubEmployeeRepository is a stub implementation of employee.Repository.
type StubEmployeeRepository struct {
	// Поля для хранения заранее заданных данных
	employee employee.Entity
	err      error
}

func (s *StubEmployeeRepository) CreateEntityTx(tx *sqlx.Tx, entity *employee.Entity) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (s *StubEmployeeRepository) FindByNameTx(tx *sqlx.Tx, name string) (isExists bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (s *StubEmployeeRepository) BeginTransaction() (tx *sqlx.Tx, err error) {
	//TODO implement me
	panic("implement me")
}

func (s *StubEmployeeRepository) FindAllEmployees() ([]employee.Entity, error) {
	//TODO implement me
	panic("implement me")
}

func (s *StubEmployeeRepository) FindAllEmployeesByIds(ids []int64) ([]employee.Entity, error) {
	//TODO implement me
	panic("implement me")
}

func (s *StubEmployeeRepository) CreateEmployee(entity *employee.Entity) (employee.Entity, error) {
	//TODO implement me
	panic("implement me")
}

func (s *StubEmployeeRepository) UpdateEmployee(entity *employee.Entity) error {
	//TODO implement me
	panic("implement me")
}

func (s *StubEmployeeRepository) DeleteEmployeeById(id int64) error {
	//TODO implement me
	panic("implement me")
}

func (s *StubEmployeeRepository) DeleteAllEmployeesByIds(ids []int64) error {
	//TODO implement me
	panic("implement me")
}

func (s *StubEmployeeRepository) FindById(id int64) (employee.Entity, error) {
	return s.employee, s.err
}

func TestEmployeeService_GetEmployeeById(t *testing.T) {
	// создаём экземпляр объекта с ассерт c функциями
	var a = assert.New(t)

	t.Run("when get Employee By ID then returns employee", func(t *testing.T) {
		// Создаём тестовые данные
		now := time.Now()

		expectedEntity := employee.Entity{
			Id:        1,
			Name:      "John Doe",
			CreatedAt: now,
			UpdatedAt: now,
		}
		expectedResponse := expectedEntity.ToResponse()

		// Создаём заглушку
		repo := &StubEmployeeRepository{ // <- & создаёт указатель (как `new E()` в Java)
			employee: expectedEntity,
			err:      nil,
		}

		// Создаём сервис с заглушкой
		validator := new(mocks.MockValidator)           // TODO mocks validator
		service := employee.NewService(repo, validator) // создаём новый экземпляр сервиса (чтобы передать ему новый мок репозитория)

		// Вызываем метод сервиса
		got, err := service.FindById(1)

		// Проверяем результаты
		a.NoError(err)
		a.Equal(expectedResponse, got)
	})

	t.Run("when get Employee By ID then returns error", func(t *testing.T) {
		// Создаём заглушку с ошибкой
		repo := &StubEmployeeRepository{ // <- & создаёт указатель (как `new E()` в Java)
			employee: employee.Entity{},
			err:      errors.New("employee not found"),
		}

		// Создаём сервис с заглушкой
		validator := new(mocks.MockValidator)           // TODO mocks validator
		service := employee.NewService(repo, validator) // создаём новый экземпляр сервиса (чтобы передать ему новый мок репозитория)

		// Вызываем метод сервиса
		got, err := service.FindById(1)

		// Проверяем результаты
		a.Error(err)
		a.Equal(employee.Response{}, got)
		a.Equal("error finding employee with id 1: employee not found", err.Error())
	})
}
