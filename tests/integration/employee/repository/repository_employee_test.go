package repository

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"idm/tests/fixtures"
	"idm/tests/testutils"
	"log"
	"strings"
	"testing"
	"time"
)

func TestEmployeeRepository(t *testing.T) {
	//arrange
	appContext := context.Background() //— если не нужно проверить таймауты
	a := assert.New(t)
	var db = testutils.InitTestDB() //var db = database.ConnectDb()
	fixture := fixtures.NewFixture(db)
	var clearDatabase = func() {
		fixture.CleanDatabase()
	}

	// func for cleaning DB in case panic
	defer func() {
		if err := recover(); err != nil {
			log.Print("The recovery function received an error while executing!")
		}
		clearDatabase()
	}()

	//var employeeRepository = employee.NewEmployeeRepository(db)
	repo := fixture.EmployeeRepository()
	var fixtureEmployee = fixtures.NewFixtureEmployee(repo)

	t.Run("create and find Employee by ID using Transaction", func(t *testing.T) {
		employeeID, err := fixtureEmployee.EmployeeTx(appContext, "John Sena")
		if !a.NoError(err) {
			return
		}
		a.NotZero(employeeID)
		log.Printf("Created EmployeeID: %d", employeeID)

		// Даем время на коммит транзакции
		time.Sleep(100 * time.Millisecond)
		log.Printf("EmployeeID -->> : %d", employeeID)

		got, err := repo.FindById(appContext, employeeID)

		a.Nil(err)
		a.NotEmpty(got)
		a.NotEmpty(got.Id)
		a.NotEmpty(got.CreatedAt)
		a.NotEmpty(got.UpdatedAt)
		a.Equal(employeeID, got.Id)
		a.Equal("John Sena", got.Name)
		a.True(got.CreatedAt.After(time.Now().Add(-5 * time.Second)))
		// сколько раз вызван сервис?
		clearDatabase()
	})

	t.Run("when create Employee using Transaction and Entity already exist ", func(t *testing.T) {
		var expErr = errors.New("employee with name John Sena already exists")

		var newEmployeeId = fixtureEmployee.Employee(appContext, "John Sena")
		a.NotZero(newEmployeeId)

		expt, err := repo.FindById(appContext, newEmployeeId)
		a.Nil(err)
		a.NotEmpty(expt)

		_, err = fixtureEmployee.EmployeeTx(appContext, "John Sena")
		time.Sleep(100 * time.Millisecond)

		a.NotNil(err)
		a.NotEmpty(err)
		a.Equal(err, expErr)

		// сколько раз вызван сервис?
		clearDatabase()
	})

	t.Run("create and find employee by id", func(t *testing.T) {
		var newEmployeeId = fixtureEmployee.Employee(appContext, "John Doe")

		got, err := repo.FindById(appContext, newEmployeeId)

		a.Nil(err)
		a.NotEmpty(got)
		a.NotEmpty(got.Id)
		a.NotEmpty(got.CreatedAt)
		a.NotEmpty(got.UpdatedAt)
		a.Equal(newEmployeeId, got.Id)
		a.Equal("John Doe", got.Name)
		a.True(got.CreatedAt.After(time.Now().Add(-5 * time.Second)))

		clearDatabase()
	})

	t.Run("update an employee by id", func(t *testing.T) {
		a := assert.New(t)

		var newEmployeeResult = fixtureEmployee.Employee(appContext, "Test Name")
		var newEmployee = fixtureEmployee.EmployeeUpdate(newEmployeeResult, "Test2 Name", time.Now(), time.Now())

		err := repo.UpdateEmployee(appContext, &newEmployee)

		a.Nil(err)

		clearDatabase()
	})

	t.Run("find all employees", func(t *testing.T) {
		var newEmployeeId1 = fixtureEmployee.Employee(appContext, "Alice Marcus")
		_ = fixtureEmployee.Employee(appContext, "John Sena")

		// делаем промежуточную проверку на наличие сотрудника
		got1, err := repo.FindById(appContext, newEmployeeId1)
		t.Log("Result Set: ", got1.Name, " ", err)

		got, err := repo.FindAllEmployees(appContext)

		a.Nil(err)
		a.NotEmpty(got)
		a.NotEmpty(got[0].Id)
		a.NotEmpty(got[1].CreatedAt)
		a.NotEmpty(got[1].UpdatedAt)
		a.Equal(newEmployeeId1, got[0].Id)
		a.Equal("Alice Marcus", got[0].Name)

		clearDatabase()
	})

	t.Run("find all employees by ids", func(t *testing.T) {
		employeeOneId := fixtureEmployee.Employee(appContext, "Test Name")
		employeeTwoId := fixtureEmployee.Employee(appContext, "Test2 2Name")
		var ids = []int64{employeeOneId, employeeTwoId}

		got, err := repo.FindAllEmployeesByIds(appContext, ids)

		a.Nil(err)
		a.NotEmpty(got)
		a.NotEmpty(got[0].Id)
		a.NotEmpty(got[1].CreatedAt)
		a.NotEmpty(got[1].UpdatedAt)
		a.Equal("Test Name", got[0].Name)
		a.Equal("Test2 2Name", got[1].Name)

		clearDatabase()
	})

	t.Run("delete all employees by ids", func(t *testing.T) {
		employeeOneId := fixtureEmployee.Employee(appContext, "Test Name")
		employeeTwoId := fixtureEmployee.Employee(appContext, "Test2 2Name")
		var ids = []int64{employeeOneId, employeeTwoId}

		err := repo.DeleteAllEmployeesByIds(appContext, ids)
		a.Nil(err, "Delete should not return error")

		res, err := repo.FindById(appContext, employeeTwoId)
		log.Println("Result Set: ", res.Name, " ", err)

		a.Error(err, "Should return error after deletion")
		a.Contains(err.Error(), "no rows in result set",
			"Error should indicate missing row")

		// 5. Дополнительная проверка стиля (enterprise-вариант)
		if err == nil || !strings.Contains(err.Error(), "no rows in result set") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}

		// Проверка, что оба сотрудника удалены
		_, err1 := repo.FindById(appContext, employeeOneId)
		_, err2 := repo.FindById(appContext, employeeTwoId)
		a.Error(err1)
		a.Error(err2)
		clearDatabase()
	})

	t.Run("delete employee by id", func(t *testing.T) {
		employeeOneId := fixtureEmployee.Employee(appContext, "Test Name")

		err := repo.DeleteEmployeeById(appContext, employeeOneId)

		a.Nil(err)

		_, err = repo.FindById(appContext, employeeOneId)
		a.Error(err)
		a.Contains(err.Error(), "no rows")

		clearDatabase()
	})
}
