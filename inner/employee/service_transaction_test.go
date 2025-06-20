package employee

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"testing"
	"time"
)

func TestEmployeeServiceCreateEmployeeTx(t *testing.T) {

	var a = assert.New(t)          // создаём экземпляр объекта с ассерт-функциями
	db, mock, err := sqlmock.New() // Создаем mocks DB
	require.NoError(t, err)
	defer func() {
		err := db.Close()
		require.NoError(t, err)
	}()

	// Обертываем в sqlx
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	// var1 - успешное создание нового работника
	t.Run("when success create Employee using Trx", func(t *testing.T) {
		// Arrange
		repo := NewRepository(sqlxDB)
		validator := new(MockValidator)
		service := NewService(repo, validator)
		now := time.Now()
		//db, mocks, err := sqlmock.New()

		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer func() {
			err := db.Close()
			require.NoError(t, err)
		}()

		//// Создаем mocks транзакции
		//mockTx := &sqlx.Tx{
		//	Tx: &sql.Tx{}, // Только базовый Tx
		//}

		//Тестовые данные
		//entityRequest := employee.Entity{
		//	Name:      "John Sena",
		//	CreatedAt: time.Now(),
		//	UpdatedAt: time.Now(),
		//}

		//expectedEntity := employee.Entity{
		//	Id:        1,
		//	Name:      entityRequest.Name,
		//	CreatedAt: entityRequest.CreatedAt,
		//	UpdatedAt: entityRequest.UpdatedAt,
		//}

		entityRequest := &CreateRequest{Name: "John Sena"}

		//---/
		mock.ExpectBegin()
		mock.ExpectQuery("select exists(select 1 from employees where name = $1)").
			WithArgs("John Sena").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
		mock.ExpectQuery("INSERT INTO employees(name, created_at, updated_at) VALUES($1, $2, $3) RETURNING *").
			WithArgs("John Sena", now, now).
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
					AddRow(1, "John Sena", now, now),
			)
		mock.ExpectCommit()

		//---//

		// 5. Вызов метода
		result, err := service.CreateEmployeeTx(*entityRequest)

		// 6. Проверки
		a.NoError(err)
		//require.NoError(t, err)
		require.NotNil(t, result)
		//require.Equal(t, int64(1), result.Id)
		//require.NoError(t, mocks.ExpectationsWereMet())
		// we make sure that all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
	//var2 create employee
	t.Run("", func(t *testing.T) {
		// Arrange
		repo := new(MockRepo)
		validator := new(MockValidator)
		service := NewService(repo, validator)

		now := time.Now()
		//entityRequest := &employee.Entity{
		//	Name:      "John Sena",
		//	CreatedAt: now,
		//	UpdatedAt: now,
		//}

		entityRequest := CreateRequest{Name: "John Sena"}
		expectedEntity := Entity{
			Id:        1,
			Name:      "John Sena",
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Настройка моков
		repo.On("BeginTransaction").Return(&sqlx.Tx{}, nil)
		repo.On("FindByNameTx", "John Sena").Return(false, nil)
		repo.On("CreateEntityTx", entityRequest).Return(expectedEntity, nil)

		// Act
		result, err := service.CreateEmployeeTx(entityRequest)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedEntity.ToResponse(), result)
		repo.AssertExpectations(t)
	})

	//   - не удалось создать транзакцию
	t.Run("should return error when transaction creation fails", func(t *testing.T) {
		repo := new(MockRepo)
		validator := new(MockValidator)
		service := NewService(repo, validator)

		// Подготавливаем тестовые данные
		entityRequest := CreateRequest{Name: "John Sena"}
		expectedErr := errors.New("tx creation error")

		// Настраиваем мок: возвращаем ошибку при создании транзакции
		repo.On("BeginTransaction").Return((*sqlx.Tx)(nil), expectedErr)

		// Вызываем метод
		result, err := service.CreateEmployeeTx(entityRequest)

		// Проверяем
		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.Empty(t, result)
		repo.AssertExpectations(t)
		a.Nil(err)
	})
	// Ошибка при проверке наличия работника
	t.Run("should return error when employee check fails", func(t *testing.T) {
		repo := new(MockRepo)
		validator := new(MockValidator)
		service := NewService(repo, validator)
		entityRequest := CreateRequest{Name: "John Sena"}
		tx := (*sqlx.Tx)(nil) // nil указатель правильного типа
		checkErr := errors.New("check error")

		repo.On("BeginTransaction").Return(tx, nil)
		repo.On("FindByNameTx", tx, entityRequest.Name).Return(false, checkErr)

		result, err := service.CreateEmployeeTx(entityRequest)

		assert.Error(t, err)
		assert.ErrorIs(t, err, checkErr)
		assert.Empty(t, result)
		repo.AssertExpectations(t)
	})

	//Работник уже существует
	t.Run("should return error when employee already exists", func(t *testing.T) {
		repo := new(MockRepo)
		validator := new(MockValidator)
		service := NewService(repo, validator)
		entityRequest := CreateRequest{Name: "John Sena"}
		tx := (*sqlx.Tx)(nil)

		repo.On("BeginTransaction").Return(tx, nil)
		repo.On("FindByNameTx", tx, entityRequest.Name).Return(true, nil) // Возвращаем true - работник существует

		result, err := service.CreateEmployeeTx(entityRequest)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
		assert.Empty(t, result)
		repo.AssertExpectations(t)
	})

	//Ошибка при создании работника
	t.Run("should return error when employee creation fails", func(t *testing.T) {
		repo := new(MockRepo)
		validator := new(MockValidator)
		service := NewService(repo, validator)
		entityRequest := CreateRequest{Name: "John Sena"}
		tx := (*sqlx.Tx)(nil)
		createErr := errors.New("creation error")

		repo.On("BeginTransaction").Return(tx, nil)
		repo.On("FindByNameTx", tx, entityRequest.Name).Return(false, nil)
		repo.On("CreateEntityTx", tx, entityRequest).Return(Entity{}, createErr)

		result, err := service.CreateEmployeeTx(entityRequest)

		assert.Error(t, err)
		assert.ErrorIs(t, err, createErr)
		assert.Empty(t, result)
		repo.AssertExpectations(t)
	})
}

//Сервис:

//   - ошибка при проверке наличия работника с таким именем
//   - работник с таким именем уже есть в базе данных
//   - работника с таким именем нет в базе данных, но создание нового работника завершилось ошибкой
