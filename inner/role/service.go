package role

import (
	"fmt"
)

type Service struct {
	repo Repo
}

type Repo interface {
	FindById(id int64) (Entity, error)
	CreateRole(entity Entity) (Entity, error)
	UpdateRole(entity *Entity) error
	FindAllRoles() ([]Entity, error)
	FindAllRolesByIds(ids []int64) ([]Entity, error)
	DeleteRoleById(id int64) error
	DeleteAllRolesByIds(ids []int64) error
}

// NewService - функция-конструктор
func NewService(repo Repo) *Service {
	return &Service{
		repo: repo,
	}
}

// FindAll - найти все элементы коллекции
func (svc *Service) FindAll() ([]Response, error) {
	var roles, err = svc.repo.FindAllRoles()
	if err != nil {
		return nil, fmt.Errorf("error finding Roles : %w", err)
	}

	responses := make([]Response, 0, len(roles))
	for _, entity := range roles {
		responses = append(responses, entity.toResponse())
	}
	return responses, err
}

// FindAllByIds - найти слайс элементов коллекции по слайсу их id
func (svc *Service) FindAllByIds(ids []int64) ([]Response, error) {
	var roles, err = svc.repo.FindAllRolesByIds(ids)
	if err != nil {
		return nil, fmt.Errorf("error find all Roles with IDs: %d %w", ids, err)
	}

	responses := make([]Response, 0, len(roles))
	for _, role := range roles {
		responses = append(responses, role.toResponse())
	}

	return responses, err
}

func (svc *Service) FindById(id int64) (Response, error) {
	entity, err := svc.repo.FindById(id)
	if err != nil {
		// в случае ошибки, вернём пустую структуру Response и обёрнутую нами ошибку
		return Response{}, fmt.Errorf("error finding role with id %d: %w", id, err)
	}

	// в случае успеха вернём структуру Response и nil вместо ошибки
	return entity.toResponse(), nil
}

func (svc *Service) Create(entity Entity) (Response, error) {
	var entityRsl, err = svc.repo.CreateRole(entity)
	if err != nil {
		return Response{}, fmt.Errorf("error creating Role with name %s: %w", entity.Name, err)
	}

	return entityRsl.toResponse(), nil
}

func (svc *Service) Update(entity *Entity) (Response, error) {
	var err = svc.repo.UpdateRole(entity)
	if err != nil {
		return Response{}, fmt.Errorf("error updating Role with name %s: %w", entity.Name, err)
	}

	return Response{}, err
}

func (svc *Service) DeleteById(id int64) (Response, error) {
	var err = svc.repo.DeleteRoleById(id)
	if err != nil {
		return Response{}, fmt.Errorf("error delete Role by ID: %d, %w", id, err)
	}

	return Response{}, err
}

func (svc *Service) DeleteByIds(ids []int64) (Response, error) {
	var err = svc.repo.DeleteAllRolesByIds(ids)
	if err != nil {
		return Response{}, fmt.Errorf("error deleting Roles by IDs: %d, %w", ids, err)
	}

	return Response{}, err
}
