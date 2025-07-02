package fixtures

import (
	"context"
	"idm/inner/role"
	"time"
)

type FixtureRole struct {
	role *role.Repository
}

// NewFixture - функция-конструктор, принимающая на вход employee.Repository (3) для работы с employee.Entity
func NewFixtureRole(roles *role.Repository) *FixtureRole {
	return &FixtureRole{roles}
}

// Role создает тестовую роль
func (f *FixtureRole) Role(
	ctx context.Context,
	name string,
	employeeID *int64,
) int64 {
	roleEntity := &role.Entity{
		Name:       name,
		EmployeeID: employeeID,
	}

	var result, err = f.role.CreateRole(ctx, roleEntity)
	if err != nil {
		panic(err)
	}

	return result.Id
}

func (f *FixtureRole) RoleUpdate(
	id int64,
	name string,
	employeeID *int64,
	createAt time.Time,
	updateAt time.Time,
) role.Entity {
	roleEntity := role.Entity{
		Id:         id,
		Name:       name,
		EmployeeID: employeeID,
		CreatedAt:  createAt,
		UpdatedAt:  updateAt,
	}

	return roleEntity
}
