package role

import (
	"github.com/jmoiron/sqlx"
	_ "idm/inner/employee"
	"time"
)

type RoleRepository struct {
	db *sqlx.DB
}

func NewRoleRepository(databese *sqlx.DB) *RoleRepository {
	return &RoleRepository{db: databese}
}

type RoleEntity struct {
	Id         int64     `db:"id"`
	Name       string    `db:"name"`
	EmployeeID *int64    `db:"employee_id"` // Nullable, указатель
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

// CreateRole - добавить новый элемент в коллекцию
func (r *RoleRepository) CreateRole(entity RoleEntity) (roleEntity RoleEntity, err error) {
	err = r.db.Get(&roleEntity, `
        INSERT INTO roles (name, employee_id, created_at, updated_at) 
        VALUES ($1, $2, $3, $4)
        RETURNING id, name, employee_id, created_at, updated_at`,
		entity.Name, entity.EmployeeID, time.Now(), time.Now())

	return roleEntity, err
}

// FindById - найти элемент коллекции по его id (этот метод мы реализовали на уроке)
func (r *RoleRepository) FindById(id int64) (entity RoleEntity, err error) {
	err = r.db.Get(&entity, "SELECT * FROM roles WHERE id = $1", id)
	return entity, err
}

// UpdateEmployee - UPDATE / Для Update лучше принимать указатель, так как мы модифицируем сущность: -> *
func (r *RoleRepository) UpdateEmployee(entity *RoleEntity) error {
	_, err := r.db.Exec(
		"UPDATE roles SET name = $1, updated_at = $2 WHERE id = $3",
		entity.Name, time.Now(), entity.Id)
	return err
}

// FindAllRoles - найти все элементы коллекции
func (r *RoleRepository) FindAllRoles() (roleEntities []RoleEntity, err error) {
	err = r.db.Select(&roleEntities, "SELECT * FROM roles")
	return roleEntities, err
}

// FindAllRolesByIds - найти слайс элементов коллекции по слайсу их id
func (r *RoleRepository) FindAllRolesByIds(ids []int64) (roleEntities []RoleEntity, err error) {
	query, args, err := sqlx.In("SELECT * FROM roles WHERE id IN (?)", ids)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	err = r.db.Select(&roleEntities, query, args...)
	return roleEntities, err
}

// DeleteAllRolesByIds - удалить элементы по слайсу их id
func (r *RoleRepository) DeleteAllRolesByIds(ids []int64) (err error) {
	query, args, err := sqlx.In("DELETE FROM roles WHERE id IN (?)", ids)
	if err != nil {
		return err
	}
	query = r.db.Rebind(query)
	_, err = r.db.Exec(query, args...)
	return err
}

// DeleteRoleById - удалить элемент коллекции по его id
func (r *RoleRepository) DeleteRoleById(id int64) (err error) {
	_, err = r.db.Exec("DELETE FROM roles WHERE id = $1", id)
	return err
}
