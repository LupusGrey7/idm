package fixtures

import (
	"github.com/jmoiron/sqlx"
	"idm/inner/employee"
	"idm/inner/role"
)

// Fixture - общая фикстура для всех сущностей
type Fixture struct {
	db        *sqlx.DB
	employees *employee.Repository
	roles     *role.Repository
}

// NewFixture - функция конструктор, создает новую фикстуру
func NewFixture(db *sqlx.DB) *Fixture {
	return &Fixture{
		db:        db,
		employees: employee.NewRepository(db),
		roles:     role.NewRepository(db),
	}
}

// CleanDatabase - очищает все таблицы
func (f *Fixture) CleanDatabase() {
	f.db.MustExec("TRUNCATE TABLE employees, roles RESTART IDENTITY CASCADE")
}

// EmployeeRepository возвращает репозиторий для работы с сотрудниками
func (f *Fixture) EmployeeRepository() *employee.Repository {
	return f.employees
}

// RoleRepository возвращает репозиторий для работы с ролями
func (f *Fixture) RoleRepository() *role.Repository {
	return f.roles
}
