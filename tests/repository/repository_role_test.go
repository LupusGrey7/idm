package repository

import (
	"github.com/stretchr/testify/assert"
	"idm/tests/fixtures"
	"idm/tests/testutils"
	"testing"
	"time"
)

func TestRoleRepository(t *testing.T) {
	//arrange
	a := assert.New(t)
	var db = testutils.InitTestDB() //add connect, run migrations /was- var db = database.ConnectDb()
	fixture := fixtures.NewFixture(db)
	var clearDatabase = func() {
		//db.MustExec("delete from roles")
		fixture.CleanDatabase() // truncate tables
	}

	// func for cleaning DB in case panic
	defer func() {
		if err := recover(); err != nil {
		}
		clearDatabase()
	}()

	//var roleRepository = role.NewRoleRepository(db)
	repo := fixture.RoleRepository()
	var fixtureRole = fixtures.NewFixtureRole(repo)

	repoEmployee := fixture.EmployeeRepository()
	var fixtureEmployee = fixtures.NewFixtureEmployee(repoEmployee)

	t.Run("create and find an role by id", func(t *testing.T) {
		var employeeId = fixtureEmployee.Employee("Employee1 Sam")
		var roleId = fixtureRole.Role("DBU", &employeeId)

		got, err := repo.FindById(roleId)

		a.Nil(err)
		a.NotEmpty(got)

		a.NotEmpty(got.Id)
		a.NotEmpty(got.CreatedAt)
		a.NotEmpty(got.UpdatedAt)
		a.Equal(roleId, got.Id)
		a.Equal("DBU", got.Name)

		clearDatabase()
	})

	t.Run("update an role by id", func(t *testing.T) {
		var employeeId = fixtureEmployee.Employee("Employee1 Sam")
		var roleID = fixtureRole.Role("DBU", &employeeId)
		var roleEntity = fixtureRole.RoleUpdate(roleID, "DBA", &employeeId, time.Now(), time.Now())

		err := repo.UpdateEmployee(&roleEntity)

		a.Nil(err)

		clearDatabase()
	})

	t.Run("find all roles", func(t *testing.T) {
		var employeeId = fixtureEmployee.Employee("Employee1 Sam")
		var employeeId2 = fixtureEmployee.Employee("Employee2 Din")
		roleId := fixtureRole.Role("DBU", &employeeId)
		_ = fixtureRole.Role("DBA", &employeeId2)

		got, err := repo.FindAllRoles()

		a.Nil(err)
		a.NotEmpty(got)
		a.NotEmpty(got[0].Id)
		a.NotEmpty(got[1].CreatedAt)
		a.NotEmpty(got[1].UpdatedAt)
		a.Equal(roleId, got[0].Id)
		a.Equal("DBU", got[0].Name)

		clearDatabase()
	})

	t.Run("find all roles by ids", func(t *testing.T) {
		var employeeId = fixtureEmployee.Employee("Employee1 Sam")
		var employeeId2 = fixtureEmployee.Employee("Employee2 Din")
		roleOneId := fixtureRole.Role("DBU", &employeeId)
		roleTwoId := fixtureRole.Role("DBA", &employeeId2)
		var ids []int64 = []int64{roleOneId, roleTwoId}

		got, err := repo.FindAllRolesByIds(ids)

		a.Nil(err)
		a.NotEmpty(got)
		a.NotEmpty(got[0].Id)
		a.NotEmpty(got[1].CreatedAt)
		a.NotEmpty(got[1].UpdatedAt)
		a.Equal("DBU", got[0].Name)
		a.Equal("DBA", got[1].Name)

		clearDatabase()
	})

	t.Run("delete all roles by ids", func(t *testing.T) {
		roleOneId := fixtureRole.Role("DBU", nil)
		roleTwoId := fixtureRole.Role("DBA", nil)
		var ids []int64 = []int64{roleOneId, roleTwoId}

		err := repo.DeleteAllRolesByIds(ids)

		// Assert
		a.Nil(err)
		for _, id := range ids {
			_, err := repo.FindById(id)
			a.Error(err)
			a.Contains(err.Error(), "no rows")
		}

		clearDatabase()
	})

	t.Run("delete role by id", func(t *testing.T) {
		var employeeId = fixtureEmployee.Employee("Employee1 Sam")
		employeeOneId := fixtureRole.Role("DBA", &employeeId)

		err := repo.DeleteRoleById(employeeOneId)

		a.Nil(err)
		a.Equal(nil, err)

		clearDatabase()
	})
}
