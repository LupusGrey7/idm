package tests

import (
	"github.com/stretchr/testify/assert"
	"idm/inner/database"
	"idm/inner/employee"
	"testing"
	"time"
)

func TestEmployeeRepository(t *testing.T) {
	//arrange
	a := assert.New(t)
	var db = database.ConnectDb()
	var clearDatabase = func() {
		db.MustExec("delete from employees")
	}

	// func for cleaning DB in case panic
	defer func() {
		if err := recover(); err != nil {
		}
		clearDatabase()
	}()

	var employeeRepository = employee.NewEmployeeRepository(db)
	var fixture = NewFixture(employeeRepository)

	t.Run("find an employee by id", func(t *testing.T) {
		var newEmployeeId = fixture.Employee("Test Name")

		got, err := employeeRepository.FindById(newEmployeeId)

		a.Nil(err)
		a.NotEmpty(got)
		a.NotEmpty(got.Id)
		a.NotEmpty(got.CreatedAt)
		a.NotEmpty(got.UpdatedAt)
		a.Equal("Test Name", got.Name)
		clearDatabase()
	})

	t.Run("update an employee by id", func(t *testing.T) {
		var newEmployeeResult = fixture.Employee("Test Name")
		var newEmployee = fixture.EmployeeUpdate(newEmployeeResult, "Test2 Name", time.Now(), time.Now())

		err := employeeRepository.UpdateEmployee(&newEmployee)

		a.Nil(err)

		clearDatabase()
	})

	t.Run("find all employees", func(t *testing.T) {
		_ = fixture.Employee("Test Name")
		_ = fixture.Employee("Test2 2Name")

		got, err := employeeRepository.FindAllEmployees()

		a.Nil(err)
		a.NotEmpty(got)
		a.NotEmpty(got[0].Id)
		a.NotEmpty(got[1].CreatedAt)
		a.NotEmpty(got[1].UpdatedAt)
		a.Equal("Test Name", got[0].Name)

		clearDatabase()
	})

	t.Run("find all employees by ids", func(t *testing.T) {
		employeeOneId := fixture.Employee("Test Name")
		employeeTwoId := fixture.Employee("Test2 2Name")
		var ids []int64 = []int64{employeeOneId, employeeTwoId}

		got, err := employeeRepository.FindAllEmployeesByIds(ids)

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
		employeeOneId := fixture.Employee("Test Name")
		employeeTwoId := fixture.Employee("Test2 2Name")
		var ids []int64 = []int64{employeeOneId, employeeTwoId}

		err := employeeRepository.DeleteAllEmployeesByIds(ids)
		res, err1 := employeeRepository.FindById(employeeTwoId)

		a.Nil(err)
		a.Nil(err1)
		a.Equal(nil, res.Name) //? not shore

		clearDatabase()
	})

	t.Run("delete employee by id", func(t *testing.T) {
		employeeOneId := fixture.Employee("Test Name")

		err := employeeRepository.DeleteEmployeeById(employeeOneId)

		a.Nil(err)
		a.Equal(nil, err)

		clearDatabase()
	})
}
