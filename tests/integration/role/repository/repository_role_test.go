package repository

import (
	"context"
	"github.com/stretchr/testify/assert"
	"idm/inner/role"
	"idm/tests/fixtures"
	"idm/tests/testutils"
	"log"
	"strings"
	"testing"
	"time"
)

func TestRoleRepository(t *testing.T) {
	//arrange
	appContext := context.Background() //— если не нужно проверить таймауты
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
			log.Print("The recovery function received an error while executing!")
		}
		clearDatabase()
	}()

	//var roleRepository = role.NewRoleRepository(db)
	repo := fixture.RoleRepository()
	var fixtureRole = fixtures.NewFixtureRole(repo)

	employeeRepo := fixture.EmployeeRepository()
	var fixtureEmployee = fixtures.NewFixtureEmployee(employeeRepo)

	t.Run("when create and then find role by id", func(t *testing.T) {

		empID := fixtureEmployee.Employee(appContext, "John Doe") // Создаём сотрудника
		roleID := fixtureRole.Role(appContext, "DBA", &empID)     // Создаём роль с сотрудником

		got, err := repo.FindById(appContext, roleID)

		a.Nil(err)
		a.NotEmpty(got)

		a.NotEmpty(got.Id)
		a.NotEmpty(got.CreatedAt)
		a.NotEmpty(got.UpdatedAt)
		a.Equal(roleID, got.Id)
		a.Equal("DBA", got.Name)

		clearDatabase()
	})

	t.Run("update an role by id", func(t *testing.T) {
		// Создаём сотрудника и роль
		empID := fixtureEmployee.Employee(appContext, "John Doe")
		roleID := fixtureRole.Role(appContext, "DBA", &empID)

		var roleEntity = fixtureRole.RoleUpdate(roleID, "DBA", &empID, time.Now(), time.Now())
		err := repo.UpdateRole(appContext, &roleEntity)

		a.Nil(err)

		clearDatabase()
	})

	t.Run("find all roles", func(t *testing.T) {
		var empID1 = fixtureEmployee.Employee(appContext, "John Doe")
		var empID2 = fixtureEmployee.Employee(appContext, "Alice Marcus")
		roleId := fixtureRole.Role(appContext, "DBU", &empID1)
		_ = fixtureRole.Role(appContext, "DBA", &empID2)

		got, err := repo.FindAllRoles(appContext)

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
		var empID1 = fixtureEmployee.Employee(appContext, "John Doe")
		var empID2 = fixtureEmployee.Employee(appContext, "Alice Marcus")
		roleOneId := fixtureRole.Role(appContext, "DBU", &empID1)
		roleTwoId := fixtureRole.Role(appContext, "DBA", &empID2)
		var ids = []int64{roleOneId, roleTwoId}

		got, err := repo.FindAllRolesByIds(appContext, ids)

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
		roleOneId := fixtureRole.Role(appContext, "DBU", nil)
		roleTwoId := fixtureRole.Role(appContext, "DBA", nil)
		var ids = []int64{roleOneId, roleTwoId}

		err := repo.DeleteAllRolesByIds(appContext, ids)

		// Assert - Проверка, что оба сотрудника удалены
		a.Nil(err)
		for _, id := range ids {
			_, err := repo.FindById(appContext, id)
			a.Error(err)
			a.Contains(err.Error(), "no rows")
		}

		clearDatabase()
	})

	t.Run("delete role by id", func(t *testing.T) {
		// Создаём сотрудника и роль
		empID := fixtureEmployee.Employee(appContext, "John Doe")
		roleID := fixtureRole.Role(appContext, "DBA", &empID)

		// Удаляем роль
		err := repo.DeleteRoleById(appContext, roleID)
		a.Nil(err, "DeleteRoleById should not return error")

		// Пытаемся найти удалённую роль
		res, err := repo.FindById(appContext, roleID)
		expected := role.Entity{}

		// Проверяем, что роль не найдена
		assert.Error(t, err)
		a.Contains(err.Error(), "no rows in result set", "Error should be 'not found'")
		assert.Equal(t, expected.Id, res.Id)
		assert.Equal(t, expected.Name, res.Name)
		assert.True(t, res.CreatedAt.IsZero())
		assert.True(t, res.UpdatedAt.IsZero())

		// Дополнительная проверка (можно опустить, так как a.Contains уже проверяет ошибку)
		if err == nil || !strings.Contains(err.Error(), "no rows in result set") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}

		clearDatabase()
	})

	t.Run("when delete employee, role should be deleted (CASCADE)", func(t *testing.T) {
		// Создаём сотрудника и роль
		empID := fixtureEmployee.Employee(appContext, "John Doe")
		roleID := fixtureRole.Role(appContext, "DBA", &empID)

		// Проверяем, что роль привязана к сотруднику
		role, err := repo.FindById(appContext, roleID)
		assert.NoError(t, err, "Role should exist")
		assert.Equal(t, empID, *role.EmployeeID, "Role should be linked to employee")

		// Удаляем сотрудника (должно удалить роль из-за ON DELETE CASCADE)
		err = employeeRepo.DeleteEmployeeById(appContext, empID)
		assert.NoError(t, err, "DeleteEmployeeById should not fail")

		// Проверяем, что роль удалилась
		_, err = repo.FindById(appContext, roleID)
		assert.Error(t, err, "Role should be deleted after employee deletion")
		assert.Contains(t, err.Error(), "no rows in result set", "Error should be 'not found'")

		clearDatabase()
	})
}
