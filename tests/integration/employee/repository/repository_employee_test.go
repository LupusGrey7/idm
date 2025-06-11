package repository

import (
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

	t.Run("create and find employee by id", func(t *testing.T) {
		var newEmployeeId = fixtureEmployee.Employee("John Doe")

		got, err := repo.FindById(newEmployeeId)

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

		var newEmployeeResult = fixtureEmployee.Employee("Test Name")
		var newEmployee = fixtureEmployee.EmployeeUpdate(newEmployeeResult, "Test2 Name", time.Now(), time.Now())

		err := repo.UpdateEmployee(&newEmployee)

		a.Nil(err)

		clearDatabase()
	})

	t.Run("find all employees", func(t *testing.T) {
		employeeId := fixtureEmployee.Employee("John Doe")
		_ = fixtureEmployee.Employee("Test2 2Name")

		got, err := repo.FindAllEmployees()

		a.Nil(err)
		a.NotEmpty(got)
		a.NotEmpty(got[0].Id)
		a.NotEmpty(got[1].CreatedAt)
		a.NotEmpty(got[1].UpdatedAt)
		a.Equal(employeeId, got[0].Id)
		a.Equal("John Doe", got[0].Name)

		clearDatabase()
	})

	t.Run("find all employees by ids", func(t *testing.T) {
		employeeOneId := fixtureEmployee.Employee("Test Name")
		employeeTwoId := fixtureEmployee.Employee("Test2 2Name")
		var ids []int64 = []int64{employeeOneId, employeeTwoId}

		got, err := repo.FindAllEmployeesByIds(ids)

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
		employeeOneId := fixtureEmployee.Employee("Test Name")
		employeeTwoId := fixtureEmployee.Employee("Test2 2Name")
		var ids []int64 = []int64{employeeOneId, employeeTwoId}

		err := repo.DeleteAllEmployeesByIds(ids)
		a.Nil(err, "Delete should not return error")

		res, err := repo.FindById(employeeTwoId)
		log.Println("Result Set: ", res.Name, " ", err)

		a.Error(err, "Should return error after deletion")
		a.Contains(err.Error(), "no rows in result set",
			"Error should indicate missing row")

		// 5. Дополнительная проверка стиля (enterprise-вариант)
		if err == nil || !strings.Contains(err.Error(), "no rows in result set") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}

		// Проверка, что оба сотрудника удалены
		_, err1 := repo.FindById(employeeOneId)
		_, err2 := repo.FindById(employeeTwoId)
		a.Error(err1)
		a.Error(err2)
		clearDatabase()
	})

	t.Run("delete employee by id", func(t *testing.T) {
		employeeOneId := fixtureEmployee.Employee("Test Name")

		err := repo.DeleteEmployeeById(employeeOneId)

		a.Nil(err)

		_, err = repo.FindById(employeeOneId)
		a.Error(err)
		a.Contains(err.Error(), "no rows")

		clearDatabase()
	})
}
