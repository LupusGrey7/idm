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

type PageResponse struct {
	Result     []Response `json:"result"`
	PageSize   int64      `json:"page_size" `
	PageNumber int64      `json:"page_number"`
	Total      int64      `json:"total"`
}

func (e *Entity) ToResponse() Response {
	return Response{
		Id:       e.Id,
		Name:     e.Name,
		CreateAt: e.CreatedAt,
		UpdateAt: e.UpdatedAt,
	}
}
func (e *Entity) ToPageResponses(
	entities []Entity,
	pageValues []int64,
	total int64,
) PageResponse {
	var response []Response
	for _, entity := range entities {
		response = append(response, entity.ToResponse())
	}

	return PageResponse{
		Result:     response,
		PageNumber: pageValues[0],
		PageSize:   pageValues[1],
		Total:      total,
	}
}

func (e *Entity) ToResponseList(entities []*Entity) []Response {
	var responses []Response
	for _, entity := range entities {
		responses = append(responses, entity.ToResponse())
	}
	return responses
}

type CreateRequest struct {
	Name string `json:"name" validate:"required,min=2,max=155"`
}

func (req *CreateRequest) ToEntity() *Entity {
	return &Entity{Name: req.Name}
}

type UpdateRequest struct {
	Id        int64     `json:"id" validate:"required,min=1"`
	Name      string    `json:"name" validate:"required,min=2,max=155"`
	CreatedAt time.Time `json:"createdAt" validate:"required"`
	UpdatedAt time.Time `json:"updatedAt" validate:"required"`
}

func (req *UpdateRequest) ToEntity() *Entity {
	return &Entity{
		Id:        req.Id,
		Name:      req.Name,
		CreatedAt: req.CreatedAt,
		UpdatedAt: req.UpdatedAt,
	}
}

type UpdateByIDRequest struct {
	ID int64 `validate:"required,min=1"`
}

type FindAllByIdsRequest struct {
	IDs []int64 `validate:"required,min=1,dive,min=1"`
}

type FindByIDRequest struct {
	ID int64 `validate:"required,min=1"`
}

type PageRequest struct {
	PageSize   int64 `validate:"required,min=1,max=155"` //gt=0,lte=100
	PageNumber int64 `validate:"required,gt=0"`
	TextFilter string
}

type DeleteByIdsRequest struct {
	IDs []int64 `validate:"required,min=1,dive,min=1"`
}

type DeleteByIdRequest struct {
	ID int64 `validate:"required,min=1"`
}
