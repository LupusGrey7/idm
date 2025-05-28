package employee

import (
	"github.com/jmoiron/sqlx"
	"time"
)

type EmployeeRepository struct {
	db *sqlx.DB
}

func NewEmployeeRepository(database *sqlx.DB) *EmployeeRepository {
	return &EmployeeRepository{db: database}
}

type EmployeeEntity struct {
	Id        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// FindById - найти элемент коллекции по его id
func (r *EmployeeRepository) FindById(id int64) (employee EmployeeEntity, err error) {
	err = r.db.Get(&employee, "SELECT * FROM employees WHERE id = $1", id)
	return employee, err
}

// CreateEmployee - добавить новый элемент в коллекцию
func (r *EmployeeRepository) CreateEmployee(entity EmployeeEntity) (employee EmployeeEntity, err error) {
	err = r.db.Get(&employee,
		"INSERT INTO employees(name, created_at, updated_at) VALUES($1, $2, $3) RETURNING *",
		entity.Name, time.Now(), time.Now())
	return employee, err
}

// FindAllEmployees - найти все элементы коллекции
func (r *EmployeeRepository) FindAllEmployees() (employees []EmployeeEntity, err error) {
	err = r.db.Select(&employees, "SELECT * FROM employees")
	return employees, err
}

// FindAllEmployeesByIds - найти слайс элементов коллекции по слайсу их id
func (r *EmployeeRepository) FindAllEmployeesByIds(ids []int64) (employees []EmployeeEntity, err error) {
	query, args, err := sqlx.In("SELECT * FROM employees WHERE id IN (?)", ids)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	err = r.db.Select(&employees, query, args...)
	return employees, err
}

// DeleteAllEmployeesByIds - удалить элементы по слайсу их id
func (r *EmployeeRepository) DeleteAllEmployeesByIds(ids []int64) (err error) {
	query, args, err := sqlx.In("DELETE FROM employees WHERE id IN (?)", ids)
	if err != nil {
		return err
	}
	query = r.db.Rebind(query)
	_, err = r.db.Exec(query, args...)
	return err
}

// DeleteEmployeeById - удалить элемент коллекции по его id
func (r *EmployeeRepository) DeleteEmployeeById(id int64) (err error) {
	_, err = r.db.Exec("DELETE FROM employees WHERE id = $1", id)
	return err
}
