package employee

import (
	"context"
	"github.com/jmoiron/sqlx"
	"log"
	"time"
)

type Repository struct {
	db *sqlx.DB
}

// NewRepository - функция-конструктор
func NewRepository(database *sqlx.DB) *Repository {
	return &Repository{db: database}
}

// BeginTransaction - great transaction for Repository
func (r *Repository) BeginTransaction() (tx *sqlx.Tx, err error) {
	return r.db.Beginx()
}

// FindAllEmployees - найти все элементы коллекции
func (r *Repository) FindAllEmployees(ctx context.Context) (employees []Entity, err error) {
	//	err = r.db.Select(&employees, "SELECT * FROM employees")
	query := `SELECT * FROM employees`
	err = r.db.SelectContext(ctx, &employees, query)

	return employees, err
}

// FindAllEmployeesByIds - найти слайс элементов коллекции по слайсу их id
func (r *Repository) FindAllEmployeesByIds(ctx context.Context, ids []int64) (employees []Entity, err error) {
	query, args, err := sqlx.In("SELECT * FROM employees WHERE id IN (?)", ids)

	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	//	err = r.db.Select(&employees, query, args...)
	err = r.db.SelectContext(ctx, &employees, query, args...)

	return employees, err
}

// FindById - найти элемент коллекции по его id
func (r *Repository) FindById(ctx context.Context, id int64) (employee Entity, err error) {
	//err = r.db.Get(&employee, "SELECT * FROM employees WHERE id = $1", id)
	err = r.db.GetContext(ctx, &employee, "SELECT * FROM employees WHERE id = $1", id)

	return employee, err
}

// FindByNameTx - Проверить наличие в базе данных сотрудника с заданным именем
func (r *Repository) FindByNameTx(
	ctx context.Context,
	tx *sqlx.Tx,
	name string,
) (isExists bool, err error) {
	//err = tx.Get(
	//	&isExists,
	//	"select exists(select 1 from employees where name = $1)",
	//	name,
	//)
	err = tx.GetContext(
		ctx,
		&isExists,
		"select exists(select 1 from employees where name = $1)",
		name)

	return isExists, err
}

// CreateEntityTx - created Employee using DB Transaction
func (r *Repository) CreateEntityTx(
	ctx context.Context,
	tx *sqlx.Tx,
	entity *Entity,
) (employeeId int64, err error) {
	//err = tx.Get(
	//	&employeeId,
	//	"INSERT INTO employees(name, created_at, updated_at) VALUES($1, $2, $3) RETURNING id",
	//	entity.Name, time.Now(), time.Now(),
	//)

	err = tx.GetContext(
		ctx,
		&employeeId,
		"INSERT INTO employees(name, created_at, updated_at) VALUES($1, $2, $3) RETURNING id",
		entity.Name, time.Now(), time.Now(),
	)
	return employeeId, err
}

// CreateEmployee - добавить новый элемент в коллекцию
func (r *Repository) CreateEmployee(
	ctx context.Context,
	entity *Entity,
) (result Entity, err error) {

	//query, args, err := sqlx.In("INSERT INTO employees(name, created_at, updated_at) VALUES($1, NOW(), NOW()) RETURNING *", entity.Name)
	query := `
		INSERT INTO employees(name, created_at, updated_at) VALUES($1, NOW(), NOW()) RETURNING *
	`
	args := []interface{}{entity.Name}

	err = r.db.GetContext(ctx, &result, query, args...)
	log.Printf("Result Employee ->> %v", result)
	return result, err
}

// UpdateEmployee - Для Update лучше принимать указатель, так как мы модифицируем сущность: -> *
func (r *Repository) UpdateEmployee(
	ctx context.Context,
	entity *Entity,
) error {
	//_, err := r.db.Exec(
	//	"UPDATE employees SET name = $1, updated_at = $2 WHERE id = $3",
	//	entity.Name, time.Now(), entity.Id)

	_, err := r.db.ExecContext(
		ctx,
		"UPDATE employees SET name = $1, updated_at = $2 WHERE id = $3",
		entity.Name, time.Now(), entity.Id)
	return err
}

// DeleteAllEmployeesByIds - удалить элементы по слайсу их id
func (r *Repository) DeleteAllEmployeesByIds(
	ctx context.Context,
	ids []int64,
) error {

	query, args, err := sqlx.In("DELETE FROM employees WHERE id IN (?)", ids)
	if err != nil {
		return err
	}
	query = r.db.Rebind(query)
	//_, err = r.db.Exec(query, args...)
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

// DeleteEmployeeById - удалить элемент коллекции по его id
func (r *Repository) DeleteEmployeeById(
	ctx context.Context,
	id int64,
) error {
	//	_, err := r.db.Exec("DELETE FROM employees WHERE id = $1", id)
	_, err := r.db.ExecContext(ctx, "DELETE FROM employees WHERE id = $1", id)

	return err
}
