package employee

import (
	"time"
)

type Entity struct {
	Id        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Response struct {
	Id       int64     `json:"id"`
	Name     string    `json:"name"`
	CreateAt time.Time `json:"createAt"`
	UpdateAt time.Time `json:"updateAt"`
}

func (e *Entity) ToResponse() Response {
	return Response{
		Id:       e.Id,
		Name:     e.Name,
		CreateAt: e.CreatedAt,
		UpdateAt: e.UpdatedAt,
	}
}
