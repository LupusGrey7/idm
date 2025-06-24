package fixtures

import (
	"fmt"
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
	employeeEntity := &employee.Entity{
		Name: name,
	}
	var result, err = f.employees.CreateEmployee(employeeEntity)
	if err != nil {
		panic(err)
	}
	return result.Id
}

func (f *FixtureEmployee) EmployeeTx(name string) (int64, error) {
	employeeEntity := &employee.Entity{
		Name: name,
	}

	tx, err := f.employees.BeginTransaction() // create for transaction methods logic new tx
	if err != nil {
		return 0, fmt.Errorf("begin transaction failed: %w", err)
	}
	// Гарантируем закрытие транзакции в любом случае
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				fmt.Printf("commit failed: %v", commitErr)
			}
		}
	}()

	isExist, err := f.employees.FindByNameTx(tx, employeeEntity.Name)
	if err != nil {
		panic(err)
	}

	if isExist {
		return 0, fmt.Errorf("employee with name %s already exists", employeeEntity.Name)
	}
	result, err := f.employees.CreateEntityTx(tx, employeeEntity)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Employee --> : %v", result)

	return result, err
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
