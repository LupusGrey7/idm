package tests

import (
	"idm/inner/role"
	"time"
)

type FixtureRole struct {
	role *role.RoleRepository
}

// NewFixture - функция-конструктор, принимающая на вход employee.Repository (3) для работы с employee.Entity
func NewFixtureRole(roles *role.RoleRepository) *FixtureRole {
	return &FixtureRole{roles}
}

func (f *FixtureRole) Role(name string) int64 {
	roleEntity := role.RoleEntity{
		Name: name,
	}

	var result, err = f.role.CreateRole(roleEntity)
	if err != nil {
		panic(err)
	}

	return result.Id
}

func (f *FixtureRole) RoleUpdate(
	id int64,
	name string,
	createAt time.Time,
	updateAt time.Time,
) role.RoleEntity {
	roleEntity := role.RoleEntity{
		Id:        id,
		Name:      name,
		CreatedAt: createAt,
		UpdatedAt: updateAt,
	}

	return roleEntity
}
