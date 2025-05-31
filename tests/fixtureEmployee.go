package tests

import (
	"idm/inner/employee"
	"time"
)

// FixtureEmployee - create fixture like a OOP stile
type FixtureEmployee struct {
	employees *employee.EmployeeRepository
}

// NewFixture - функция-конструктор, принимающая на вход employee.Repository (3) для работы с employee.Entity
func NewFixture(employees *employee.EmployeeRepository) *FixtureEmployee {
	return &FixtureEmployee{employees}
}

func (f *FixtureEmployee) Employee(name string) int64 {
	employeeEntity := employee.EmployeeEntity{
		Name: name,
	}
	var result, err = f.employees.CreateEmployee(employeeEntity)
	if err != nil {
		panic(err)
	}
	return result.Id
}

func (f *FixtureEmployee) EmployeeUpdate(
	id int64,
	name string,
	createAt time.Time,
	updateAt time.Time,
) employee.EmployeeEntity {
	employeeEntity := employee.EmployeeEntity{
		Id:        id,
		Name:      name,
		CreatedAt: createAt,
		UpdatedAt: updateAt,
	}
	return employeeEntity
}
