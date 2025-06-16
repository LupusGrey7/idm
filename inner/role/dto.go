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

func (e *Entity) ToResponse() Response {
	return Response{
		Id:         e.Id,
		Name:       e.Name,
		EmployeeID: e.EmployeeID,
		CreateAt:   e.CreatedAt,
		UpdateAt:   e.UpdatedAt,
	}
}

type CreateRequest struct {
	Name string `json:"name" validate:"required,min=2,max=155"`
}

func (req *CreateRequest) ToEntity() *Entity {
	return &Entity{Name: req.Name}
}

type UpdateRequest struct {
	Id         int64     `json:"id" validate:"required,min=1,max=2147483647`
	EmployeeID *int64    `json:"employeeID" validate:"required,min=1,max=2147483647"` // fixme?
	Name       string    `json:"name" validate:"required,min=2,max=155"`
	CreatedAt  time.Time `json:"createdAt" validate:"required"`
	UpdatedAt  time.Time `json:"updatedAt" validate:"required"`
}

func (req *UpdateRequest) ToEntity() *Entity {
	return &Entity{Id: req.Id, EmployeeID: req.EmployeeID, Name: req.Name, CreatedAt: req.CreatedAt, UpdatedAt: req.UpdatedAt}
}

type UpdateByIDRequest struct {
	ID int64 `validate:"required,min=1"`
}

type FindByIDRequest struct {
	ID int64 `validate:"required,min=1"`
}

type DeleteByIdsRequest struct {
	IDs []int64 `validate:"required,min=1,dive,min=1"`
}

type FindAllByIdsRequest struct {
	IDs []int64 `validate:"required,min=1,dive,min=1"`
}

type DeleteByIdRequest struct {
	ID int64 `validate:"required,min=1"`
}
