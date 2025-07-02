package role

import (
	"context"
	"github.com/jmoiron/sqlx"
	"time"
)

type Repository struct {
	db *sqlx.DB
}

// NewRepository - функция-конструктор
func NewRepository(databese *sqlx.DB) *Repository {
	return &Repository{db: databese}
}

// FindAllRoles - найти все элементы коллекции
func (r *Repository) FindAllRoles(ctx context.Context) (roleEntities []Entity, err error) {
	//	err = r.db.Select(&roleEntities, "SELECT * FROM roles")
	query := `SELECT * FROM roles`
	err = r.db.SelectContext(ctx, &roleEntities, query)

	return roleEntities, err
}

// FindAllRolesByIds - найти слайс элементов коллекции по слайсу их id
func (r *Repository) FindAllRolesByIds(ctx context.Context, ids []int64) (roleEntities []Entity, err error) {
	query, args, err := sqlx.In("SELECT * FROM roles WHERE id IN (?)", ids)

	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	//	err = r.db.Select(&roleEntities, query, args...)
	err = r.db.SelectContext(ctx, &roleEntities, query, args...)

	return roleEntities, err
}

// CreateRole - добавить новый элемент в коллекцию
func (r *Repository) CreateRole(ctx context.Context, entity *Entity) (roleEntity Entity, err error) {
	err = r.db.GetContext(
		ctx,
		&roleEntity,
		`INSERT INTO roles (name, employee_id, created_at, updated_at) 
        VALUES ($1, $2, $3, $4)
        RETURNING id, name, employee_id, created_at, updated_at`,
		entity.Name, entity.EmployeeID, time.Now(), time.Now(),
	)

	return roleEntity, err
}

// FindById - найти элемент коллекции по его id (этот метод мы реализовали на уроке)
func (r *Repository) FindById(ctx context.Context, id int64) (entity Entity, err error) {
	//err = r.db.Get(&entity, "SELECT * FROM roles WHERE id = $1", id)
	err = r.db.GetContext(ctx, &entity, "SELECT * FROM roles WHERE id = $1", id)

	return entity, err
}

// UpdateEmployee - UPDATE / Для Update лучше принимать указатель, так как мы модифицируем сущность: -> *
func (r *Repository) UpdateRole(ctx context.Context, entity *Entity) error {
	_, err := r.db.ExecContext(
		ctx,
		"UPDATE roles SET name = $1, updated_at = $2 WHERE id = $3",
		entity.Name, time.Now(), entity.Id,
	)

	return err
}

// DeleteAllRolesByIds - удалить элементы по слайсу их id
func (r *Repository) DeleteAllRolesByIds(ctx context.Context, ids []int64) (err error) {
	query, args, err := sqlx.In("DELETE FROM roles WHERE id IN (?)", ids)
	if err != nil {
		return err
	}

	query = r.db.Rebind(query)
	_, err = r.db.ExecContext(ctx, query, args...)

	return err
}

// DeleteRoleById - удалить элемент коллекции по его id
func (r *Repository) DeleteRoleById(ctx context.Context, id int64) (err error) {
	_, err = r.db.ExecContext(ctx, "DELETE FROM roles WHERE id = $1", id)

	return err
}
