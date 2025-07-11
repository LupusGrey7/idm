package employee

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
	"strings"
	"time"
)

// Repository - infra layer
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

// GetPageByValues
// LIMIT number_of_rows: Определяет максимальное количество строк, которое будет возвращено запросом.
// OFFSET starting_row: Указывает, сколько строк нужно пропустить в начале набора результатов, прежде чем начать выборку.
// TextFilter не менее, 3 не пробельных (" ", "\n", "\t" и т.п.) символов.
func (r *Repository) GetPageByValues(
	ctx context.Context,
	pageValues []int64, // [pageSize, offset]
	textFilter string,
) ([]Entity, int64, error) {
	// 1. Валидация pageValues
	if len(pageValues) != 2 {
		return nil, 0, fmt.Errorf("invalid page values format")
	}

	// 2. Подготовка запросов
	baseQuery := "SELECT id, name, created_at, updated_at FROM employees"
	countQuery := "SELECT COUNT(*) FROM employees"

	var args []interface{}
	paramCounter := 1

	// 3. Добавляем фильтр только если он не пустой
	if textFilter != "" {
		filteredText := strings.TrimSpace(textFilter)
		if len(filteredText) >= 3 {
			safePattern := "%" + strings.ReplaceAll(filteredText, "%", "\\%") + "%"
			baseQuery += fmt.Sprintf(" WHERE name ILIKE $%d", paramCounter)
			countQuery += fmt.Sprintf(" WHERE name ILIKE $%d", paramCounter)
			args = append(args, safePattern)
			paramCounter++
		}
	}

	// 4. Добавляем пагинацию
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", paramCounter, paramCounter+1)
	args = append(args, pageValues[0], pageValues[1])

	// 5. Выполняем запросы
	var employees []Entity
	err := r.db.SelectContext(ctx, &employees, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get employees: %w", err)
	}

	var total int64
	err = r.db.GetContext(ctx, &total, countQuery, args[:paramCounter-1]...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return employees, total, nil
}

// FindAllEmployeesByIds - найти слайс элементов коллекции по слайсу их id
func (r *Repository) FindAllEmployeesByIds(
	ctx context.Context,
	ids []int64,
) (employees []Entity, err error) {
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
