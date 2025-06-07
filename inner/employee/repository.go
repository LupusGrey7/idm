package employee

import (
	"github.com/jmoiron/sqlx"
	"time"
)

type Repository struct {
	db *sqlx.DB
}

// NewRepository - функция-конструктор
func NewRepository(database *sqlx.DB) *Repository {
	return &Repository{db: database}
}

// FindAllEmployees - найти все элементы коллекции
func (r *Repository) FindAllEmployees() (employees []Entity, err error) {
	err = r.db.Select(&employees, "SELECT * FROM employees")

	return employees, err
}

// FindAllEmployeesByIds - найти слайс элементов коллекции по слайсу их id
func (r *Repository) FindAllEmployeesByIds(ids []int64) (employees []Entity, err error) {
	query, args, err := sqlx.In("SELECT * FROM employees WHERE id IN (?)", ids)

	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	err = r.db.Select(&employees, query, args...)

	return employees, err
}

// FindById - найти элемент коллекции по его id
func (r *Repository) FindById(id int64) (employee Entity, err error) {
	err = r.db.Get(&employee, "SELECT * FROM employees WHERE id = $1", id)
	return employee, err
}

// CreateEmployee - добавить новый элемент в коллекцию
func (r *Repository) CreateEmployee(entity Entity) (employee Entity, err error) {
	err = r.db.Get(&employee,
		"INSERT INTO employees(name, created_at, updated_at) VALUES($1, $2, $3) RETURNING *",
		entity.Name, time.Now(), time.Now())
	return employee, err
}

// UPDATE - Для Update лучше принимать указатель, так как мы модифицируем сущность: -> *
func (r *Repository) UpdateEmployee(entity *Entity) error {
	_, err := r.db.Exec(
		"UPDATE employees SET name = $1, updated_at = $2 WHERE id = $3",
		entity.Name, time.Now(), entity.Id)

	return err
}

// DeleteAllEmployeesByIds - удалить элементы по слайсу их id
func (r *Repository) DeleteAllEmployeesByIds(ids []int64) error {

	query, args, err := sqlx.In("DELETE FROM employees WHERE id IN (?)", ids)
	if err != nil {
		return err
	}
	query = r.db.Rebind(query)
	_, err = r.db.Exec(query, args...)

	return err
}

// DeleteEmployeeById - удалить элемент коллекции по его id
func (r *Repository) DeleteEmployeeById(id int64) error {
	_, err := r.db.Exec("DELETE FROM employees WHERE id = $1", id)
	return err
}
