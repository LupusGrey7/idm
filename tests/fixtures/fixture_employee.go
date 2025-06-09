package fixtures

import (
	"idm/inner/employee"
	"time"
)

// FixtureEmployee - create fixtures like a OOP stile
type FixtureEmployee struct {
	employees *employee.Repository
}

// NewFixtureEmployee - функция-конструктор, принимающая на вход employee.Repository (3) для работы с employee.Entity
func NewFixtureEmployee(employees *employee.Repository) *FixtureEmployee {
	return &FixtureEmployee{employees}
}

// Employee - создает тестового сотрудника
func (f *FixtureEmployee) Employee(name string) int64 {
	employeeEntity := employee.Entity{
		Name: name,
	}
	var result, err = f.employees.CreateEmployee(employeeEntity)
	if err != nil {
		panic(err)
	}
	return result.Id
}

// EmployeeUpdate - Entity создает сущность сотрудника для обновления
func (f *FixtureEmployee) EmployeeUpdate(
	id int64,
	name string,
	createAt time.Time,
	updateAt time.Time,
) employee.Entity {
	employeeEntity := employee.Entity{
		Id:        id,
		Name:      name,
		CreatedAt: createAt,
		UpdatedAt: updateAt,
	}
	return employeeEntity
}
