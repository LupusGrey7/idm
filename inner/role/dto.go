package role

import (
	"time"
)

type Entity struct {
	Id         int64     `db:"id"`
	Name       string    `db:"name"`
	EmployeeID *int64    `db:"employee_id"` // Nullable, указатель
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

type Response struct {
	Id         int64     `json:"id"`
	Name       string    `json:"name"`
	EmployeeID *int64    `json:"employeeID"`
	CreateAt   time.Time `json:"createAt"`
	UpdateAt   time.Time `json:"updateAt"`
}

func (e *Entity) toResponse() Response {
	return Response{
		Id:         e.Id,
		Name:       e.Name,
		EmployeeID: e.EmployeeID,
		CreateAt:   e.CreatedAt,
		UpdateAt:   e.UpdatedAt,
	}
}
