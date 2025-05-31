package tests

import (
	"github.com/stretchr/testify/assert"
	"idm/inner/database"
	"idm/inner/role"
	"testing"
	"time"
)

func TestRoleRepository(t *testing.T) {

	//arrange
	a := assert.New(t)
	var db = database.ConnectDb()
	var clearDatabase = func() {
		db.MustExec("delete from roles")
	}

	// func for cleaning DB in case panic
	defer func() {
		if err := recover(); err != nil {
		}
		clearDatabase()
	}()

	var roleRepository = role.NewRoleRepository(db)
	var fixture = NewFixtureRole(roleRepository)

	t.Run("find an role by id", func(t *testing.T) {
		var newEmployeeId = fixture.Role("Test Name")

		got, err := roleRepository.FindById(newEmployeeId)

		a.Nil(err)
		a.NotEmpty(got)

		a.NotEmpty(got.Id)
		a.NotEmpty(got.CreatedAt)
		a.NotEmpty(got.UpdatedAt)
		a.Equal("Test Name", got.Name)
		clearDatabase()
	})

	t.Run("update an role by id", func(t *testing.T) {
		var newEmployeeResult = fixture.Role("Test Name")
		var newEmployee = fixture.RoleUpdate(newEmployeeResult, "Test2 Name", time.Now(), time.Now())

		err := roleRepository.UpdateEmployee(&newEmployee)

		a.Nil(err)

		clearDatabase()
	})

	t.Run("find all roles", func(t *testing.T) {
		_ = fixture.Role("Test Name")
		_ = fixture.Role("Test2 2Name")

		got, err := roleRepository.FindAllRoles()

		a.Nil(err)
		a.NotEmpty(got)
		a.NotEmpty(got[0].Id)
		a.NotEmpty(got[1].CreatedAt)
		a.NotEmpty(got[1].UpdatedAt)
		a.Equal("Test Name", got[0].Name)

		clearDatabase()
	})

	t.Run("find all roles by ids", func(t *testing.T) {
		employeeOneId := fixture.Role("Test Name")
		employeeTwoId := fixture.Role("Test2 2Name")
		var ids []int64 = []int64{employeeOneId, employeeTwoId}

		got, err := roleRepository.FindAllRolesByIds(ids)

		a.Nil(err)
		a.NotEmpty(got)
		a.NotEmpty(got[0].Id)
		a.NotEmpty(got[1].CreatedAt)
		a.NotEmpty(got[1].UpdatedAt)
		a.Equal("Test Name", got[0].Name)
		a.Equal("Test2 2Name", got[1].Name)

		clearDatabase()
	})

	t.Run("delete all roles by ids", func(t *testing.T) {
		employeeOneId := fixture.Role("Test Name")
		employeeTwoId := fixture.Role("Test2 2Name")
		var ids []int64 = []int64{employeeOneId, employeeTwoId}

		err := roleRepository.DeleteAllRolesByIds(ids)
		res, err1 := roleRepository.FindById(employeeTwoId)

		a.Nil(err)
		a.Nil(err1)
		a.Equal(nil, res.Name) //? not shore

		clearDatabase()
	})

	t.Run("delete role by id", func(t *testing.T) {
		employeeOneId := fixture.Role("Test Name")

		err := roleRepository.DeleteRoleById(employeeOneId)

		a.Nil(err)
		a.Equal(nil, err)

		clearDatabase()
	})
}
