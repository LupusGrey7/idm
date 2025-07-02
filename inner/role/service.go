package role

import (
	"context"
	"fmt"
	"idm/inner/domain"
)

type Service struct {
	repo      Repo
	validator Validator
}

type Repo interface {
	FindAllRoles(ctx context.Context) ([]Entity, error)
	FindAllRolesByIds(ctx context.Context, ids []int64) ([]Entity, error)
	FindById(ctx context.Context, id int64) (Entity, error)
	CreateRole(ctx context.Context, entity *Entity) (Entity, error)
	UpdateRole(ctx context.Context, entity *Entity) error
	DeleteRoleById(ctx context.Context, id int64) error
	DeleteAllRolesByIds(ctx context.Context, ids []int64) error
}
type Validator interface {
	Validate(request any) error
}

// NewService - функция-конструктор
func NewService(repo Repo, validator Validator) *Service {
	return &Service{
		repo:      repo,
		validator: validator,
	}
}

// FindAll - найти все элементы коллекции
func (svc *Service) FindAll(ctx context.Context) ([]Response, error) {
	var roles, err = svc.repo.FindAllRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("error finding Roles : %w", err)
	}

	responses := make([]Response, 0, len(roles))
	for _, entity := range roles {
		responses = append(responses, entity.ToResponse())
	}
	return responses, err
}

// FindAllByIds - найти слайс элементов коллекции по слайсу их id
func (svc *Service) FindAllByIds(
	ctx context.Context,
	ids []int64,
) ([]Response, error) {
	request := FindAllByIdsRequest{IDs: ids}                // Создаем DTO для валидации
	if err := svc.validator.Validate(request); err != nil { // Валидируем запрос
		return []Response{}, domain.RequestValidationError{Message: err.Error()}
	}

	var roles, err = svc.repo.FindAllRolesByIds(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("error find all Roles with IDs: %d %w", ids, err)
	}

	responses := make([]Response, 0, len(roles))
	for _, role := range roles {
		responses = append(responses, role.ToResponse())
	}

	return responses, err
}

func (svc *Service) FindById(
	ctx context.Context,
	id int64,
) (Response, error) {
	request := FindByIDRequest{ID: id}        // Создаем DTO для валидации
	var err = svc.validator.Validate(request) // Валидируем запрос
	if err != nil {
		// возвращаем кастомную ошибку в случае, если запрос не прошёл валидацию
		return Response{}, domain.RequestValidationError{Message: err.Error()}
	}

	entity, err := svc.repo.FindById(ctx, id)
	if err != nil {
		// в случае ошибки, вернём пустую структуру Response и обёрнутую нами ошибку
		return Response{}, fmt.Errorf("error finding role with id %d: %w", id, err)
	}

	// в случае успеха вернём структуру Response и nil вместо ошибки
	return entity.ToResponse(), nil
}

func (svc *Service) CreateRole(
	ctx context.Context,
	request CreateRequest,
) (Response, error) {
	//validate
	var err = svc.validator.Validate(request) // Валидируем запрос
	if err != nil {
		// возвращаем кастомную ошибку в случае, если запрос не прошёл валидацию
		return Response{}, domain.RequestValidationError{Message: err.Error()}
	}

	//save
	entityRole := request.ToEntity()
	entityRsl, err := svc.repo.CreateRole(ctx, entityRole)
	if err != nil {
		return Response{}, fmt.Errorf("error creating Role with name %s: %w", entityRole.Name, err)
	}

	return entityRsl.ToResponse(), nil
}

func (svc *Service) UpdateRole(
	ctx context.Context,
	id int64,
	request UpdateRequest,
) (Response, error) {
	request.Id = id
	var err = svc.validator.Validate(request)
	if err != nil {
		return Response{}, domain.RequestValidationError{Message: err.Error()}
	}

	entity := request.ToEntity()
	err = svc.repo.UpdateRole(ctx, entity)
	if err != nil {
		return Response{}, fmt.Errorf("error updating Role with name %s: %w", entity.Name, err)
	}

	return entity.ToResponse(), err
}

func (svc *Service) DeleteById(
	ctx context.Context,
	id int64,
) (Response, error) {
	requestId := DeleteByIdRequest{ID: id}
	var err = svc.validator.Validate(requestId)
	if err != nil {
		return Response{}, domain.RequestValidationError{Message: err.Error()}
	}
	err = svc.repo.DeleteRoleById(ctx, id)
	if err != nil {
		return Response{}, fmt.Errorf("error delete Role by ID: %d, %w", id, err)
	}

	return Response{}, err
}

func (svc *Service) DeleteByIds(
	ctx context.Context,
	ids []int64,
) (Response, error) {
	requestIds := DeleteByIdsRequest{IDs: ids}
	var err = svc.validator.Validate(requestIds)
	if err != nil {
		return Response{}, domain.RequestValidationError{Message: err.Error()}
	}

	err = svc.repo.DeleteAllRolesByIds(ctx, ids)
	if err != nil {
		return Response{}, fmt.Errorf("error deleting Roles by IDs: %d, %w", ids, err)
	}

	return Response{}, err
}
