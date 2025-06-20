package role

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

// Объявляем структуру мок-репозитория
type MockRepo struct {
	mock.Mock
}

// Реализация ВСЕХ методов репозитория интерфейса для тестов
func (m *MockRepo) FindById(id int64) (Entity, error) {
	args := m.Called(id) // обращаемся в Mock
	return args.Get(0).(Entity), args.Error(1)
}

func (m *MockRepo) CreateRole(entity *Entity) (Entity, error) {
	args := m.Called(entity)
	return args.Get(0).(Entity), args.Error(1)
}

func (m *MockRepo) UpdateRole(entity *Entity) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockRepo) FindAllRoles() ([]Entity, error) {
	args := m.Called()
	return args.Get(0).([]Entity), args.Error(1)
}

func (m *MockRepo) FindAllRolesByIds(ids []int64) ([]Entity, error) {
	args := m.Called(ids)
	return args.Get(0).([]Entity), args.Error(1)
}

func (m *MockRepo) DeleteRoleById(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRepo) DeleteAllRolesByIds(ids []int64) error {
	args := m.Called(ids)
	return args.Error(0)
}

func TestRoleService(t *testing.T) {
	var a = assert.New(t)

	t.Run("when should return All Roles by IDs", func(t *testing.T) {
		now := time.Now()                          // Создаем текущее время
		mockRepo := new(MockRepo)                  // Создаем мок-репозиторий
		validator := new(MockValidator)            //
		service := NewService(mockRepo, validator) // создаём новый экземпляр сервиса (чтобы передать ему новый мок репозитория)
		roleIDs := []int64{1, 2, 3}                // Создаем список идентификаторов ролей
		var validateR = FindAllByIdsRequest{IDs: roleIDs}

		roles := []Entity{ // Создаем список ролей
			{Id: 1, Name: "Admin", EmployeeID: nil, CreatedAt: now, UpdatedAt: now},
			{Id: 2, Name: "User", EmployeeID: nil, CreatedAt: now, UpdatedAt: now},
			{Id: 3, Name: "Guest", EmployeeID: nil, CreatedAt: now, UpdatedAt: now},
		}
		expectedResponses := []Response{ // Создаем список ответов
			{Id: 1, Name: "Admin", EmployeeID: nil, CreateAt: now, UpdateAt: now},
			{Id: 2, Name: "User", EmployeeID: nil, CreateAt: now, UpdateAt: now},
			{Id: 3, Name: "Guest", EmployeeID: nil, CreateAt: now, UpdateAt: now},
		}

		// Задаем ожидаемое поведение мок-репозитория
		validator.On("Validate", validateR).Return(nil)
		mockRepo.On("FindAllRolesByIds", roleIDs).Return(roles, nil)

		// Act - вызываем метод сервиса
		result, err := service.FindAllByIds(roleIDs)

		// Assert - проверяем результаты теста
		a.NoError(err)
		a.Equal(roles[0].Id, result[0].Id)
		a.EqualValues(expectedResponses[0], result[0])
		a.True(mockRepo.AssertNumberOfCalls(t, "FindAllRolesByIds", 1))
		mockRepo.AssertExpectations(t) // проверяем что были вызваны все объявленные ожидания
	})

	t.Run("when should return error when failed to get roles", func(t *testing.T) {

		mockRepo := new(MockRepo)                  // Создаем мок-репозиторий
		validator := new(MockValidator)            //
		service := NewService(mockRepo, validator) // создаём новый экземпляр сервиса (чтобы передать ему новый мок репозитория)
		roleIDs := []int64{1, 2, 3}                // Создаем список идентификаторов ролей
		var validateR = FindAllByIdsRequest{IDs: roleIDs}
		var expectedErr = errors.New("database error") // ошибка, которую вернёт репозиторий

		validator.On("Validate", validateR).Return(nil)
		mockRepo.On("FindAllRolesByIds", roleIDs).Return([]Entity{}, expectedErr) // Задаем ожидаемое поведение мок-репозитория

		// Act - вызываем метод сервиса
		_, err := service.FindAllByIds(roleIDs)

		// Assert - проверяем результаты теста
		assert.Error(t, err)                                                               // Должна быть ошибка
		assert.EqualError(t, err, "error find all Roles with IDs: [1 2 3] database error") // Проверяем конкретную ошибку
		mockRepo.AssertExpectations(t)                                                     // Проверяем, что мок был вызван
	})

	t.Run("when should return All founded Roles", func(t *testing.T) {
		mockRepo := new(MockRepo)                  // Создаем мок-репозиторий
		validator := new(MockValidator)            //
		service := NewService(mockRepo, validator) // создаём новый экземпляр сервиса (чтобы передать ему новый мок репозитория)
		roles := []Entity{                         // Создаем список ролей
			{Id: 1, Name: "Admin"},
			{Id: 2, Name: "User"},
			{Id: 3, Name: "Guest"},
		}
		expectedResponses := []Response{ // Создаем список ответов
			{Id: 1, Name: "Admin"},
			{Id: 2, Name: "User"},
			{Id: 3, Name: "Guest"},
		}

		mockRepo.On("FindAllRoles").Return(roles, nil) // Задаем ожидаемое поведение мок-репозитория

		// Act - вызываем метод сервиса
		result, err := service.FindAll()

		// Assert - проверяем результат
		a.Nil(err)
		a.NotNil(result)
		a.EqualValues(expectedResponses, result)
		a.True(mockRepo.AssertNumberOfCalls(t, "FindAllRoles", 1))
		mockRepo.AssertExpectations(t) // проверяем что были вызваны все объявленные ожидания
	})
	t.Run("when should return error when failed to get roles", func(t *testing.T) {
		mockRepo := new(MockRepo)                                   // Создаем мок-репозиторий
		validator := new(MockValidator)                             //
		service := NewService(mockRepo, validator)                  // создаём новый экземпляр сервиса (чтобы передать ему новый мок репозитория)
		want := errors.New("failed to get roles")                   // Создаем ошибку
		mockRepo.On("FindAllRoles").Return(make([]Entity, 0), want) // Задаем ожидаемое поведение мок-репозитория

		// Act - вызываем метод сервиса
		_, err := service.FindAll()

		// Assert - проверяем результат
		a.Error(err)
		a.NotEmpty(err)
		a.True(mockRepo.AssertNumberOfCalls(t, "FindAllRoles", 1))
		mockRepo.AssertExpectations(t) // проверяем что были вызваны все объявленные ожидания
	})
	t.Run("when should create Role", func(t *testing.T) {
		now := time.Now()
		var repo = new(MockRepo)
		validator := new(MockValidator)        //
		service := NewService(repo, validator) // создаём новый экземпляр сервиса (чтобы передать ему новый мок репозитория)
		//	want := errors.New("failed to get roles") // Создаем ошибку

		entityRequest := CreateRequest{ // request
			Name: "Admin",
		}

		expectedRole := Entity{ // Создаем роль
			Id:         1,
			Name:       "Admin",
			EmployeeID: nil,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		expectedEntity := entityRequest.ToEntity()
		expectedResponses := expectedRole.ToResponse()

		validator.On("Validate", entityRequest).Return(nil)
		repo.On("CreateRole", expectedEntity).Return(expectedRole, nil) // Задаем ожидаемое поведение мок-репозитория

		// Act - вызываем метод сервиса
		result, err := service.CreateRole(entityRequest)

		// Assert - проверяем результат
		a.Nil(err)
		a.NotEmpty(result)
		a.EqualValues(expectedResponses, result)
		a.True(repo.AssertNumberOfCalls(t, "CreateRole", 1))
		repo.AssertExpectations(t) // проверяем что были вызваны все объявленные ожидания
	})
	t.Run("when success update Role", func(t *testing.T) {
		now := time.Now()
		mockRepo := new(MockRepo)                  // Создаем мок-репозиторий
		validator := new(MockValidator)            //
		service := NewService(mockRepo, validator) // создаём новый экземпляр сервиса (чтобы передать ему новый мок репозитория)

		var empID = int64(1)
		entityRequest := UpdateRequest{ // request
			Id:         1,
			Name:       "Admin",
			EmployeeID: &empID,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		//want := errors.New("failed to get roles")
		//	errR := fmt.Errorf("error updating Role with name %s: %w", entityRequest.Name, want)

		expectedEntity := entityRequest.ToEntity()
		expectedResponse := expectedEntity.ToResponse()

		validator.On("Validate", entityRequest).Return(nil)
		mockRepo.On("UpdateRole", expectedEntity).Return(nil) // Задаем ожидаемое поведение мок-репозитория
		// Act - вызываем метод сервиса
		result, err := service.UpdateRole(empID, entityRequest)

		// Assert - проверяем результат
		a.Nil(err)
		a.NotNil(result)
		a.Equal(expectedResponse.Name, result.Name)
		a.True(mockRepo.AssertNumberOfCalls(t, "UpdateRole", 1))
	})
	t.Run("when delete Role by ID", func(t *testing.T) {
		var repo = new(MockRepo)
		validator := new(MockValidator)        //
		service := NewService(repo, validator) // создаём новый экземпляр сервиса (чтобы передать ему новый мок репозитория)
		var err = error(nil)                   // ошибка, которую вернёт репозиторий
		var empID = int64(1)
		var requestId = DeleteByIdRequest{ID: empID}

		var responseRsl = Response{}

		validator.On("Validate", requestId).Return(nil)
		repo.On("DeleteRoleById", empID).Return(err)

		// Act - вызываем метод сервиса
		var rsl, got = service.DeleteById(1)

		// Assert - проверяем результаты теста
		a.Nil(got)
		a.Empty(rsl)
		a.Equal(responseRsl, rsl)
		a.True(repo.AssertNumberOfCalls(t, "DeleteRoleById", 1))
		repo.AssertExpectations(t) // проверяем что были вызваны все объявленные ожидания
	})
}
