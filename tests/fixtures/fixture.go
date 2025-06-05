package fixtures

import (
	"github.com/jmoiron/sqlx"
	"idm/inner/employee"
	"idm/inner/role"
)

// Fixture - общая фикстура для всех сущностей
type Fixture struct {
	db        *sqlx.DB
	employees *employee.EmployeeRepository
	roles     *role.RoleRepository
}

// NewFixture создает новую фикстуру
func NewFixture(db *sqlx.DB) *Fixture {
	return &Fixture{
		db:        db,
		employees: employee.NewEmployeeRepository(db),
		roles:     role.NewRoleRepository(db),
	}
}

// CleanDatabase - очищает все таблицы
func (f *Fixture) CleanDatabase() {
	f.db.MustExec("TRUNCATE TABLE employees, roles RESTART IDENTITY CASCADE")
}

// EmployeeRepository возвращает репозиторий для работы с сотрудниками
func (f *Fixture) EmployeeRepository() *employee.EmployeeRepository {
	return f.employees
}

// RoleRepository возвращает репозиторий для работы с ролями
func (f *Fixture) RoleRepository() *role.RoleRepository {
	return f.roles
}
