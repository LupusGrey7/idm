package service

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"idm/inner/employee"
	"testing"
	"time"
)

// Объявляем структуру мок-репозитория
type MockRepo struct {
	mock.Mock
}

// Реализация ВСЕХ методов интерфейса для тестов
func (m *MockRepo) FindById(id int64) (employee.Entity, error) {
	args := m.Called(id)
	return args.Get(0).(employee.Entity), args.Error(1) // Приведение типов в моках (args.Get(0).(employee.Entity))
}

func (m *MockRepo) FindAllEmployees() ([]employee.Entity, error) {
	args := m.Called()
	return args.Get(0).([]employee.Entity), args.Error(1)
}

func (m *MockRepo) FindAllEmployeesByIds(ids []int64) ([]employee.Entity, error) {
	args := m.Called(ids)
	return args.Get(0).([]employee.Entity), args.Error(1) //
}

func (m *MockRepo) Create(entity *employee.Entity) (employee.Entity, error) {
	args := m.Called(entity)
	return args.Get(0).(employee.Entity), args.Error(1)
}

func (m *MockRepo) UpdateEmployee(entity *employee.Entity) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockRepo) DeleteEmployeeById(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRepo) DeleteAllEmployeesByIds(ids []int64) error {
	args := m.Called(ids)
	return args.Error(0)
}

func TestEmployeeService(t *testing.T) {

	var a = assert.New(t) // создаём экземпляр объекта с ассерт-функциями

	t.Run("should return All found employees by IDs", func(t *testing.T) {
		now := time.Now()
		var repo = new(MockRepo)
		var service = employee.NewService(repo) // Используем конструктор

		entities := []employee.Entity{
			{Id: 1, Name: "John", CreatedAt: now},
			{Id: 2, Name: "Jane", CreatedAt: now},
			{Id: 3, Name: "Jim", CreatedAt: now},
		}

		var ids = []int64{1, 2, 3}
		expectedResponses := []employee.Response{
			{Id: 1, Name: "John", CreateAt: now},
			{Id: 2, Name: "Jane", CreateAt: now},
			{Id: 3, Name: "Jim", CreateAt: now},
		}

		repo.On("FindAllEmployeesByIds", ids).Return(entities, nil)
		responses, err := service.FindAllByIds(ids)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponses, responses)
		a.EqualValues(expectedResponses[0].Name, entities[0].Name)
		repo.AssertExpectations(t)
	})
	t.Run("should return error failed to get employees by IDs", func(t *testing.T) {
		var repo = new(MockRepo)
		var service = employee.NewService(repo) // Используем конструктор
		var ids = []int64{1, 2, 3}

		var expectedErr = errors.New("database error") // ошибка, которую вернёт репозиторий
		var errRsl = fmt.Errorf("error finding employees: %w", expectedErr)
		repo.On("FindAllEmployeesByIds", ids).Return([]employee.Entity{}, expectedErr)

		// Act - вызываем метод сервиса
		responses, err := service.FindAllByIds(ids)

		// Assert - проверяем результаты теста
		assert.Error(t, err)                         // Должна быть ошибка
		assert.Equal(t, errRsl.Error(), err.Error()) // Проверяем конкретную ошибку
		assert.Nil(t, responses)                     // Результат должен быть nil
		repo.AssertExpectations(t)                   // Проверяем, что мок был вызван)
	})

	t.Run("should return All found employees", func(t *testing.T) {
		repo := new(MockRepo)
		service := employee.NewService(repo)
		now := time.Now()

		entities := []employee.Entity{
			{Id: 1, Name: "John", CreatedAt: now},
			{Id: 2, Name: "Jane", CreatedAt: now},
		}

		expectedResponses := []employee.Response{
			{Id: 1, Name: "John", CreateAt: now},
			{Id: 2, Name: "Jane", CreateAt: now},
		}

		repo.On("FindAllEmployees").Return(entities, nil) // Настройка возврата среза

		// Act - вызываем метод сервиса
		responses, err := service.FindAll()

		assert.NoError(t, err)
		assert.Equal(t, expectedResponses, responses)
		repo.AssertExpectations(t)
	})

	t.Run("should return found employee by ID", func(t *testing.T) {
		var repo = new(MockRepo) // Создаём экземпляр мок-объекта
		// Создаём экземпляр сервиса, который собираемся тестировать. Передаём в его конструктор мок вместо реального репозитория
		var service = employee.NewService(repo) // Используем конструктор
		var entity = employee.Entity{           // создаём Entity, которую должен вернуть репозиторий
			Id:        1,
			Name:      "John Doe",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		var want = entity.ToResponse() // создаём Response, который ожидаем получить от сервиса

		// конфигурируем поведение мок-репозитория (при вызове метода FindById с аргументом 1 вернуть Entity, созданную нами выше)
		// Настраиваем ожидание с ТОЧНЫМ типом аргумента
		repo.On("FindById", int64(1)).Return(entity, nil)

		// вызываем сервис с аргументом id = 1
		var got, err = service.FindById(1)

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
		var repo = new(MockRepo)                // Создаём для теста новый экземпляр мока репозитория.
		var service = employee.NewService(repo) // создаём новый экземпляр сервиса (чтобы передать ему новый мок репозитория)
		var entity = employee.Entity{}          // Создаём пустую структуру employee.Entity, которую сервис вернёт вместе с ошибкой
		var err = errors.New("database error")  // ошибка, которую вернёт репозиторий

		// ошибка, которую должен будет вернуть сервис
		var want = fmt.Errorf("error finding employee with id 1: %w", err)

		repo.On("FindById", int64(1)).Return(entity, err)

		var response, got = service.FindById(1)

		// Assert - проверяем результаты теста
		a.Empty(response)
		a.NotNil(got)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "FindById", 1))
		repo.AssertExpectations(t) // проверяем что были вызваны все объявленные ожидания
	})

	t.Run("when create Employee should return Response", func(t *testing.T) {
		var repo = new(MockRepo)
		var service = employee.NewService(repo) // Используем конструктор
		entityRequest := &employee.Entity{      // <- & создаёт указатель (как `new E()` в Java)
			Name:      "John Doe",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		expectedEntity := employee.Entity{
			Id:        1, // ID будет присвоен в репозитории
			Name:      entityRequest.Name,
			CreatedAt: entityRequest.CreatedAt,
			UpdatedAt: entityRequest.UpdatedAt,
		}

		expectedResponses := expectedEntity.ToResponse()

		// Настройка возврата, Настраиваем мок. Обратить внимание - ожидаем указатель!
		repo.On("Create", entityRequest).Return(expectedEntity, nil)

		responses, err := service.Create(entityRequest)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponses, responses)
		repo.AssertExpectations(t)
	})

	t.Run("when update Employee should return Response", func(t *testing.T) {
		repo := new(MockRepo)
		service := employee.NewService(repo)

		// Создаём объекта
		entityRequest := &employee.Entity{ // <- & создаёт указатель (как `new` в Java)
			Id:        1,
			Name:      "John Doe",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		expectedResponse := entityRequest.ToResponse()

		// Настраиваем мок. Обратите внимание - ожидаем указатель!
		repo.On("UpdateEmployee", entityRequest).Return(nil)

		response, err := service.Update(entityRequest) //передача объекта=указателя

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
		repo.AssertExpectations(t)
	})

	t.Run("when delete All Employees by employee IDs", func(t *testing.T) {

		var repo = new(MockRepo)
		var service = employee.NewService(repo) // Используем конструктор
		var IDs = []int64{1, 2, 3}

		// конфигурируем поведение мок-репозитория (при вызове метода FindById с аргументом 1 вернуть Entity, созданную нами выше)
		// Настраиваем ожидание с ТОЧНЫМ типом аргумента
		repo.On("DeleteAllEmployeesByIds", IDs).Return(nil)

		// вызываем сервис с аргументом id = 1
		_, err := service.DeleteByIds(IDs)

		// проверяем, что сервис не вернул ошибку
		a.Nil(err)

		// проверяем, что сервис вызвал репозиторий ровно 1 раз
		a.True(repo.AssertNumberOfCalls(t, "DeleteAllEmployeesByIds", 1))
		repo.AssertExpectations(t) //что были вызваны все кого планировали

	})

	t.Run("when delete Employee by ID should return err", func(t *testing.T) {

		var repo = new(MockRepo)
		var service = employee.NewService(repo)

		// ошибка, которую вернёт репозиторий
		var err = errors.New("database error")
		// ошибка, которую должен будет вернуть сервис
		var want = fmt.Errorf("error delete employee by ID: 1, %w", err)

		repo.On("DeleteEmployeeById", int64(1)).Return(err)

		var response, got = service.DeleteById(1)

		// проверяем результаты теста
		a.Empty(response)
		a.NotNil(got)
		a.Equal(want, got)
		a.True(repo.AssertNumberOfCalls(t, "DeleteEmployeeById", 1))
		repo.AssertExpectations(t) // проверяем что были вызваны все объявленные ожидания
	})

	t.Run("when delete Employee by ID", func(t *testing.T) {
		var repo = new(MockRepo)
		var service = employee.NewService(repo)

		// ошибка, которую вернёт репозиторий
		//var err = errors.New("database error")
		var err = error(nil)

		var responseRsl = employee.Response{}

		repo.On("DeleteEmployeeById", int64(1)).Return(err)

		var rsl, got = service.DeleteById(1)

		// Assert - проверяем результаты теста
		a.Nil(got)
		a.Empty(rsl)
		a.Equal(responseRsl, rsl)
		a.True(repo.AssertNumberOfCalls(t, "DeleteEmployeeById", 1))
		repo.AssertExpectations(t) // проверяем что были вызваны все объявленные ожидания
	})
}
