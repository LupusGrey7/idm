package employee

import (
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

// StubEmployeeRepository is a stub implementation of employee.Repository.
type StubEmployeeRepository struct {
	// Поля для хранения заранее заданных данных
	employee Entity
	err      error
}

func (s *StubEmployeeRepository) GetPageByValues(ctx context.Context, values []int64) ([]Entity, int64, error) {
	//TODO implement me
	panic("implement me")
}

func (s *StubEmployeeRepository) CreateEntityTx(ctx context.Context, tx *sqlx.Tx, entity *Entity) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (s *StubEmployeeRepository) FindByNameTx(ctx context.Context, tx *sqlx.Tx, name string) (isExists bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (s *StubEmployeeRepository) BeginTransaction() (tx *sqlx.Tx, err error) {
	//TODO implement me
	panic("implement me")
}

func (s *StubEmployeeRepository) FindAllEmployees(ctx context.Context) ([]Entity, error) {
	//TODO implement me
	panic("implement me")
}

func (s *StubEmployeeRepository) FindAllEmployeesByIds(ctx context.Context, ids []int64) ([]Entity, error) {
	//TODO implement me
	panic("implement me")
}

func (s *StubEmployeeRepository) CreateEmployee(ctx context.Context, entity *Entity) (Entity, error) {
	return Entity{
		Id:        1,
		Name:      entity.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (s *StubEmployeeRepository) UpdateEmployee(ctx context.Context, entity *Entity) error {
	//TODO implement me
	panic("implement me")
}

func (s *StubEmployeeRepository) DeleteEmployeeById(ctx context.Context, id int64) error {
	//TODO implement me
	panic("implement me")
}

func (s *StubEmployeeRepository) DeleteAllEmployeesByIds(ctx context.Context, ids []int64) error {
	//TODO implement me
	panic("implement me")
}

func (s *StubEmployeeRepository) FindById(ctx context.Context, id int64) (Entity, error) {
	return s.employee, s.err
}

func TestEmployeeService_GetEmployeeById(t *testing.T) {
	// создаём экземпляр объекта с ассерт c функциями
	var a = assert.New(t)
	appContext := context.Background()

	t.Run("when get Employee By ID then returns employee", func(t *testing.T) {
		// Создаём тестовые данные
		now := time.Now()

		expectedEntity := Entity{
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
		validator := new(MockValidator)        // TODO mocks validator
		service := NewService(repo, validator) // создаём новый экземпляр сервиса (чтобы передать ему новый мок репозитория)
		validator.On("Validate", mock.Anything).Return(nil)
		// Вызываем метод сервиса
		got, err := service.FindById(appContext, (1))

		// Проверяем результаты
		a.NoError(err)
		a.Equal(expectedResponse, got)
	})

	t.Run("when get Employee By ID then returns error", func(t *testing.T) {
		// Создаём заглушку с ошибкой
		repo := &StubEmployeeRepository{ // <- & создаёт указатель (как `new E()` в Java)
			employee: Entity{},
			err:      errors.New("employee not found"),
		}

		// Создаём сервис с заглушкой
		validator := new(MockValidator)        // TODO mocks validator
		service := NewService(repo, validator) // создаём новый экземпляр сервиса (чтобы передать ему новый мок репозитория)

		validator.On("Validate", mock.Anything).Return(errors.New("error finding employee with id 1: employee not found"))
		// Вызываем метод сервиса
		got, err := service.FindById(appContext, 1)

		// Проверяем результаты
		a.Error(err)
		a.Equal(Response{}, got)
		a.Equal("error finding employee with id 1: employee not found", err.Error())
	})
}
