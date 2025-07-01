package employee

import (
	"context"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"idm/inner/domain"

	"testing"
	"time"
)

// Объявляем структуру мок-репозитория
type MockRepo struct {
	mock.Mock
}

// Mock реализация методов репо
func (m *MockRepo) BeginTransaction() (*sqlx.Tx, error) {
	args := m.Called()
	return args.Get(0).(*sqlx.Tx), args.Error(1) // Приведение типов в моках (args.Get(0).(sqlx.Tx))
}
func (m *MockRepo) FindByNameTx(ctx context.Context, tx *sqlx.Tx, name string) (isExists bool, err error) {
	args := m.Called(ctx, tx, name)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockRepo) CreateEntityTx(ctx context.Context, tx *sqlx.Tx, entity *Entity) (int64, error) {
	args := m.Called(ctx, tx, entity)
	return args.Get(0).(int64), args.Error(1)
}

// Реализация ВСЕХ методов интерфейса для тестов
func (m *MockRepo) FindById(ctx context.Context, id int64) (Entity, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(Entity), args.Error(1) // Приведение типов в моках (args.Get(0).(employee.Entity))
}

func (m *MockRepo) FindAllEmployees(ctx context.Context) ([]Entity, error) {
	args := m.Called(ctx)
	return args.Get(0).([]Entity), args.Error(1)
}

func (m *MockRepo) FindAllEmployeesByIds(ctx context.Context, ids []int64) ([]Entity, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]Entity), args.Error(1) //
}

func (m *MockRepo) CreateEmployee(ctx context.Context, entity *Entity) (Entity, error) {
	args := m.Called(ctx, entity)
	return args.Get(0).(Entity), args.Error(1)
}

func (m *MockRepo) UpdateEmployee(ctx context.Context, entity *Entity) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockRepo) DeleteEmployeeById(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepo) DeleteAllEmployeesByIds(ctx context.Context, ids []int64) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

// https://pkg.go.dev/github.com/stretchr/testify/mock@v1.10.0#Mock.AssertCalled
func TestEmployeeService(t *testing.T) {

	appContext := context.Background() //— если нужно проверить таймауты
	var a = assert.New(t)              // создаём экземпляр объекта с ассерт-функциями

	t.Run("should return All found employees by IDs", func(t *testing.T) {
		now := time.Now()
		var repo = new(MockRepo)
		validator := new(MockValidator)
		var service = NewService(repo, validator) // Используем конструктор
		var ids = []int64{1, 2, 3}
		var requestIds = FindAllByIdsRequest{IDs: ids}
		entities := []Entity{
			{Id: 1, Name: "John", CreatedAt: now},
			{Id: 2, Name: "Jane", CreatedAt: now},
			{Id: 3, Name: "Jim", CreatedAt: now},
		}

		expectedResponses := []Response{
			{Id: 1, Name: "John", CreateAt: now},
			{Id: 2, Name: "Jane", CreateAt: now},
			{Id: 3, Name: "Jim", CreateAt: now},
		}
		validator.On("Validate", requestIds).Return(nil)
		repo.On("FindAllEmployeesByIds", appContext, ids).Return(entities, nil).Once()

		responses, err := service.FindAllByIds(appContext, ids)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponses, responses)
		a.EqualValues(expectedResponses[0].Name, entities[0].Name)
		repo.AssertExpectations(t)
	})
	t.Run("should return error failed to get employees by IDs", func(t *testing.T) {
		var repo = new(MockRepo)
		validator := new(MockValidator)
		var service = NewService(repo, validator) // Используем конструктор
		var ids = []int64{1, 2, 3}
		var requestIds = FindAllByIdsRequest{IDs: ids}
		var expectedErr = errors.New("database error") // ошибка, которую вернёт репозиторий
		var errRsl = fmt.Errorf("error finding employees: %w", expectedErr)

		validator.On("Validate", requestIds).Return(nil).Once()
		repo.On("FindAllEmployeesByIds", appContext, ids).Return([]Entity{}, expectedErr).Once()

		// Act - вызываем метод сервиса
		responses, err := service.FindAllByIds(appContext, ids)

		// Assert - проверяем результаты теста
		assert.Error(t, err)                         // Должна быть ошибка
		assert.Equal(t, errRsl.Error(), err.Error()) // Проверяем конкретную ошибку
		assert.Nil(t, responses)                     // Результат должен быть nil
		repo.AssertExpectations(t)                   // Проверяем, что мок был вызван)
	})

	t.Run("should return All found employees", func(t *testing.T) {
		repo := new(MockRepo)
		validator := new(MockValidator)
		service := NewService(repo, validator)
		now := time.Now()

		entities := []Entity{
			{Id: 1, Name: "John", CreatedAt: now},
			{Id: 2, Name: "Jane", CreatedAt: now},
		}

		expectedResponses := []Response{
			{Id: 1, Name: "John", CreateAt: now},
			{Id: 2, Name: "Jane", CreateAt: now},
		}

		repo.On("FindAllEmployees", appContext).Return(entities, nil).Once() // Настройка возврата среза

		// Act - вызываем метод сервиса
		responses, err := service.FindAll(appContext)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponses, responses)
		repo.AssertExpectations(t)
	})

	t.Run("should return found employee by ID", func(t *testing.T) {
		var repo = new(MockRepo)               // Создаём экземпляр мок-объекта
		validator := new(MockValidator)        // Создаём экземпляр сервиса, который собираемся тестировать. Передаём в его конструктор мок вместо реального репозитория
		service := NewService(repo, validator) // Используем конструктор
		var ID int64 = 1
		request := FindByIDRequest{ID: ID}
		var entity = Entity{ // создаём Entity, которую должен вернуть репозиторий
			Id:        1,
			Name:      "John Doe",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		var want = entity.ToResponse() // создаём Response, который ожидаем получить от сервиса

		// конфигурируем поведение мок-репозитория (при вызове метода FindById с аргументом 1 вернуть Entity, созданную нами выше)
		// Настраиваем ожидание с ТОЧНЫМ типом аргумента
		validator.On("Validate", request).Return(nil)
		repo.On("FindById", appContext, int64(1)).Return(entity, nil)

		var got, err = service.FindById(appContext, ID) // вызываем сервис с аргументом id = 1

		a.Nil(err)                                         // проверяем, что сервис не вернул ошибку
		a.Equal(want, got)                                 // проверяем, что сервис вернул нам тот employee.Response, который мы ожилали получить
		a.True(repo.AssertNumberOfCalls(t, "FindById", 1)) // проверяем, что сервис вызвал репозиторий ровно 1 раз
		repo.AssertExpectations(t)
	})

	t.Run("should return wrapped error", func(t *testing.T) {

		/* Мы собираемся проверить счётчик вызовов, поэтому хотим,
		*чтобы счётчик содержал количество вызовов к репозиторию, выполненных в рамках одного нашего теста.
		*Ели сделать мок общим для нескольких тестов, то он посчитает вызовы, которые сделали все тесты
		 */
		var repo = new(MockRepo)               // Создаём для теста новый экземпляр мока репозитория.
		validator := new(MockValidator)        // mock Validator
		service := NewService(repo, validator) // создаём новый экземпляр сервиса (чтобы передать ему новый мок репозитория)

		var entity = Entity{}                  // Создаём пустую структуру employee.Entity, которую сервис вернёт вместе с ошибкой
		var err = errors.New("database error") // ошибка, которую вернёт репозиторий
		var ID int64 = 1
		request := FindByIDRequest{ID: ID}
		// ошибка, которую должен будет вернуть сервис
		var want = fmt.Errorf("error finding employee with id 1: %w", err)

		validator.On("Validate", request).Return(nil)
		repo.On("FindById", appContext, int64(1)).Return(entity, err).Once()

		var response, got = service.FindById(appContext, ID)

		// Assert - проверяем результаты теста
		a.Empty(response)
		a.NotNil(got)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "FindById", 1))
		repo.AssertExpectations(t) // проверяем что были вызваны все объявленные ожидания
	})

	t.Run("when create Employee should return Response", func(t *testing.T) {
		var repo = new(MockRepo)               // Создаём для теста новый экземпляр мока репозитория.
		validator := new(MockValidator)        //  mocks validator
		service := NewService(repo, validator) // создаём новый экземпляр сервиса (чтобы передать ему новый мок репозитория)
		now := time.Now()
		entityRequest := CreateRequest{ // <- & создаёт указатель (как `new E()` в Java)
			Name: "John Sena",
		}
		entityResult := Entity{
			Id:        1,
			Name:      "John Sena",
			CreatedAt: now,
			UpdatedAt: now,
		}

		expectedEntity := entityRequest.ToEntity()
		expectedResponse := entityResult.ToResponse()

		validator.On("Validate", entityRequest).Return(nil)
		repo.On("CreateEmployee", appContext, expectedEntity).Return(entityResult, nil).Once() // Настройка возврата, Настраиваем мок. Обратить внимание - ожидаем указатель!

		responses, err := service.CreateEmployee(appContext, entityRequest)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, responses)
		repo.AssertExpectations(t)
	})
	//Тест на валидацию + Update
	t.Run("when update Employee should return Response", func(t *testing.T) {
		var repo = new(MockRepo)               // Создаём для теста новый экземпляр мока репозитория.
		validator := new(MockValidator)        //
		service := NewService(repo, validator) // создаём новый экземпляр сервиса (чтобы передать ему новый мок репозитория)

		now := time.Now()
		entityRequest := UpdateRequest{ // <- & создаёт указатель (как `new` в Java)// Создаём объекта
			Id:        1,
			Name:      "John Doe",
			CreatedAt: now,
			UpdatedAt: now,
		}
		expectedEntity := entityRequest.ToEntity()
		expectedResponse := expectedEntity.ToResponse()

		validator.On("Validate", entityRequest).Return(nil)
		repo.On("UpdateEmployee", appContext, expectedEntity).Return(nil).Once() // Настраиваем мок. Обратите внимание - ожидаем указатель!

		response, err := service.UpdateEmployee(appContext, 1, entityRequest) //передача объекта=указателя

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
		repo.AssertExpectations(t)
		repo.AssertNumberOfCalls(t, "UpdateEmployee", 1)
	})
	//--- Тест на ошибку валидации Name --//
	t.Run("should return validation error", func(t *testing.T) {
		var repo = new(MockRepo)               // Создаём для теста новый экземпляр мока репозитория.
		validator := new(MockValidator)        //
		service := NewService(repo, validator) // создаём новый экземпляр сервиса (чтобы передать ему новый мок репозитория)
		now := time.Now()
		invalidRequest := UpdateRequest{
			Id:        1,
			Name:      "", // невалидное имя
			CreatedAt: now,
			UpdatedAt: now,
		}

		validator.On("Validate", invalidRequest).Return(errors.New("name is required")).Once()

		_, err := service.UpdateEmployee(appContext, 1, invalidRequest)

		assert.Error(t, err)
		assert.IsType(t, domain.RequestValidationError{}, err)
		repo.AssertNotCalled(t, "UpdateEmployee")
		repo.AssertNumberOfCalls(t, "UpdateEmployee", 0)
	})
	//--- Тест на ошибку валидации ID ---//
	t.Run("should return validation error", func(t *testing.T) {
		var repo = new(MockRepo)               // Создаём для теста новый экземпляр мока репозитория.
		validator := new(MockValidator)        //
		service := NewService(repo, validator) // создаём новый экземпляр сервиса (чтобы передать ему новый мок репозитория)
		now := time.Now()
		invalidRequest := UpdateRequest{
			Id:        0,
			Name:      "John Sena", // невалидное имя
			CreatedAt: now,
			UpdatedAt: now,
		}

		validator.On("Validate", invalidRequest).Return(errors.New("id is required")).Once()

		_, err := service.UpdateEmployee(appContext, 0, invalidRequest)

		assert.Error(t, err)
		assert.IsType(t, domain.RequestValidationError{}, err)
		repo.AssertNotCalled(t, "UpdateEmployee")
		repo.AssertNumberOfCalls(t, "UpdateEmployee", 0)
	})

	t.Run("when delete All Employees by employee IDs", func(t *testing.T) {

		var repo = new(MockRepo)               // Создаём для теста новый экземпляр мока репозитория.
		validator := new(MockValidator)        //
		service := NewService(repo, validator) // создаём новый экземпляр сервиса (чтобы передать ему новый мок репозитория)
		var IDs = []int64{1, 2, 3}
		var requestIds = DeleteByIdsRequest{IDs: IDs}

		// конфигурируем поведение мок-репозитория (при вызове метода FindById с аргументом 1 вернуть Entity, созданную нами выше)
		// Настраиваем ожидание с ТОЧНЫМ типом аргумента

		validator.On("Validate", requestIds).Return(nil)
		repo.On("DeleteAllEmployeesByIds", appContext, IDs).Return(nil).Once()

		// вызываем сервис с аргументом id = 1
		_, err := service.DeleteByIds(appContext, IDs)

		// проверяем, что сервис не вернул ошибку
		a.Nil(err)

		// проверяем, что сервис вызвал репозиторий ровно 1 раз
		a.True(repo.AssertNumberOfCalls(t, "DeleteAllEmployeesByIds", 1))
		repo.AssertExpectations(t) //что были вызваны все кого планировали

	})

	t.Run("when delete Employee by ID should return err", func(t *testing.T) {

		var repo = new(MockRepo)               // Создаём для теста новый экземпляр мока репозитория.
		validator := new(MockValidator)        //
		service := NewService(repo, validator) // создаём новый экземпляр сервиса (чтобы передать ему новый мок репозитория)
		var Id int64 = 1
		var requestId = DeleteByIdRequest{ID: Id}
		// ошибка, которую вернёт репозиторий
		var err = errors.New("database error")
		// ошибка, которую должен будет вернуть сервис
		var want = fmt.Errorf("error delete employee by ID: 1, %w", err)

		validator.On("Validate", requestId).Return(nil)
		repo.On("DeleteEmployeeById", appContext, int64(1)).Return(err).Once()

		var response, got = service.DeleteById(appContext, 1)

		// проверяем результаты теста
		a.Empty(response)
		a.NotNil(got)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "DeleteEmployeeById", 1))
		repo.AssertExpectations(t) // проверяем что были вызваны все объявленные ожидания
	})

	t.Run("when delete Employee by ID", func(t *testing.T) {
		var repo = new(MockRepo)               // Создаём для теста новый экземпляр мока репозитория.
		validator := new(MockValidator)        //
		service := NewService(repo, validator) // создаём новый экземпляр сервиса (чтобы передать ему новый мок репозитория)
		var Id int64 = 1
		var requestId = DeleteByIdRequest{ID: Id}
		// ошибка, которую вернёт репозиторий
		//var err = errors.New("database error")
		var err = error(nil)

		var responseRsl = Response{}

		validator.On("Validate", requestId).Return(nil)
		repo.On("DeleteEmployeeById", appContext, Id).Return(err).Once()

		var rsl, got = service.DeleteById(appContext, Id)

		// Assert - проверяем результаты теста
		a.Nil(got)
		a.Empty(rsl)
		a.Equal(responseRsl, rsl)
		a.True(repo.AssertNumberOfCalls(t, "DeleteEmployeeById", 1))
		repo.AssertExpectations(t) // проверяем что были вызваны все объявленные ожидания
	})

}
