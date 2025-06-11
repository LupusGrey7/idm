package employee

import (
	"fmt"
)

type Service struct {
	repo Repo
}

type Repo interface {
	FindById(id int64) (Entity, error)
	FindAllEmployees() ([]Entity, error)
	FindAllEmployeesByIds(ids []int64) ([]Entity, error)
	Create(entity *Entity) (Entity, error)
	UpdateEmployee(entity *Entity) error
	DeleteEmployeeById(id int64) error
	DeleteAllEmployeesByIds(ids []int64) error
}

// NewService - функция-конструктор
func NewService(repo Repo) *Service {
	return &Service{
		repo: repo,
	}
}

func (svc *Service) FindAll() ([]Response, error) {
	entities, err := svc.repo.FindAllEmployees()
	if err != nil {
		return nil, fmt.Errorf("error finding employees: %w", err)
	}

	responses := make([]Response, 0, len(entities))
	for _, entity := range entities {
		responses = append(responses, entity.ToResponse())
	}

	return responses, nil
}

func (svc *Service) FindAllByIds(ids []int64) ([]Response, error) {
	entities, err := svc.repo.FindAllEmployeesByIds(ids)
	if err != nil {
		return nil, fmt.Errorf("error finding employees: %w", err)
	}

	responses := make([]Response, 0, len(entities))
	for _, entity := range entities {
		responses = append(responses, entity.ToResponse())
	}

	return responses, err
}

func (svc *Service) FindById(id int64) (Response, error) {
	entity, err := svc.repo.FindById(id)
	if err != nil {
		// в случае ошибки, вернём пустую структуру Response и обёрнутую нами ошибку
		return Response{}, fmt.Errorf("error finding employee with id %d: %w", id, err)
	}

	// в случае успеха вернём структуру Response и nil вместо ошибки
	return entity.ToResponse(), nil
}

func (svc *Service) Create(entity *Entity) (Response, error) {
	var entityRsl, err = svc.repo.Create(entity)
	if err != nil {
		return Response{}, fmt.Errorf("error creating employee with name %s: %w", entity.Name, err)
	}

	return entityRsl.ToResponse(), nil
}

func (svc *Service) Update(entity *Entity) (Response, error) {
	var err = svc.repo.UpdateEmployee(entity)
	if err != nil {
		return Response{}, fmt.Errorf("error updating employee with name %s: %w", entity.Name, err)
	}

	return entity.ToResponse(), nil // <- Преобразуем Entity в Response
}

func (svc *Service) DeleteById(id int64) (Response, error) {
	var err = svc.repo.DeleteEmployeeById(id)
	if err != nil {
		return Response{}, fmt.Errorf("error delete employee by ID: %d, %w", id, err)
	}

	return Response{}, err
}

func (svc *Service) DeleteByIds(ids []int64) (Response, error) {
	var err = svc.repo.DeleteAllEmployeesByIds(ids)
	if err != nil {
		return Response{}, fmt.Errorf("error deleting employees by IDs: %d, %w", ids, err)
	}

	return Response{}, err
}
