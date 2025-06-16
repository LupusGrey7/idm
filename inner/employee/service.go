package employee

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"idm/inner/common"
)

type Service struct {
	repo      Repo
	validator Validator
}

type Repo interface {
	BeginTransaction() (tx *sqlx.Tx, err error)
	FindById(id int64) (Entity, error)
	FindAllEmployees() ([]Entity, error)
	FindAllEmployeesByIds(ids []int64) ([]Entity, error)
	FindByNameTx(tx *sqlx.Tx, name string) (bool, error)
	CreateEmployee(entity *Entity) (Entity, error)
	CreateEntityTx(tx *sqlx.Tx, entity *Entity) (int64, error)
	UpdateEmployee(entity *Entity) error
	DeleteEmployeeById(id int64) error
	DeleteAllEmployeesByIds(ids []int64) error
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
	request := FindAllByIdsRequest{IDs: ids}                // Создаем DTO для валидации
	if err := svc.validator.Validate(request); err != nil { // Валидируем запрос
		return []Response{}, common.RequestValidationError{Message: err.Error()}
	}

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
	request := FindByIDRequest{ID: id}        // Создаем DTO для валидации
	var err = svc.validator.Validate(request) // Валидируем запрос
	if err != nil {
		// возвращаем кастомную ошибку в случае, если запрос не прошёл валидацию
		return Response{}, common.RequestValidationError{Message: err.Error()}
	}

	entity, err := svc.repo.FindById(id)
	if err != nil {
		// в случае ошибки, вернём пустую структуру Response и обёрнутую нами ошибку
		return Response{}, fmt.Errorf("error finding employee with id %d: %w", id, err)
	}

	// в случае успеха вернём структуру Response и nil вместо ошибки
	return entity.ToResponse(), nil
}

func (svc *Service) CreateEmployee(createRequest CreateRequest) (Response, error) {
	// Создаем DTO для валидации
	if err := svc.validator.Validate(createRequest); err != nil { // Валидируем запрос
		return Response{}, common.RequestValidationError{Message: err.Error()}
	}

	var toEntity = createRequest.ToEntity()
	var entityRsl, err = svc.repo.CreateEmployee(toEntity)
	if err != nil {
		return Response{}, fmt.Errorf("error creating employee with name %s: %w", createRequest.Name, err)
	}

	return entityRsl.ToResponse(), nil
}

func (svc *Service) UpdateEmployee(id int64, request UpdateRequest) (Response, error) {
	// Создаем DTO для валидации
	request.Id = id                                         // <- Устанавливаем ID в запросе
	if err := svc.validator.Validate(request); err != nil { // Валидируем запрос
		return Response{}, common.RequestValidationError{Message: err.Error()}
	}

	var employeeEntity = request.ToEntity()
	var err = svc.repo.UpdateEmployee(employeeEntity)
	if err != nil {
		return Response{}, fmt.Errorf("error updating employee with name %s: %w", employeeEntity.Name, err)
	}

	return employeeEntity.ToResponse(), nil // <- Преобразуем Entity в Response
}

func (svc *Service) DeleteById(id int64) (Response, error) {
	requestId := DeleteByIdRequest{ID: id}
	var err = svc.validator.Validate(requestId)
	if err != nil {
		return Response{}, common.RequestValidationError{Message: err.Error()}
	}

	err = svc.repo.DeleteEmployeeById(id)
	if err != nil {
		return Response{}, fmt.Errorf("error delete employee by ID: %d, %w", id, err)
	}

	return Response{}, err
}

func (svc *Service) DeleteByIds(ids []int64) (Response, error) {
	request := DeleteByIdsRequest{IDs: ids}           // Создаем DTO для валидации
	var errValidate = svc.validator.Validate(request) // Валидируем запрос
	if errValidate != nil {
		// возвращаем кастомную ошибку в случае, если запрос не прошёл валидацию
		return Response{}, common.RequestValidationError{Message: errValidate.Error()}
	}

	var err = svc.repo.DeleteAllEmployeesByIds(ids)
	if err != nil {
		return Response{}, fmt.Errorf("error deleting employees by IDs: %d, %w", ids, err)
	}

	return Response{}, err
}

func (svc *Service) CreateEmployeeTx(request CreateRequest) (int64, error) {
	var err = svc.validator.Validate(request) // Валидируем запрос
	if err != nil {
		// возвращаем кастомную ошибку в случае, если запрос не прошёл валидацию (про кастомные ошибки - дальше)
		return 0, common.RequestValidationError{Message: err.Error()}
	}

	tx, err := svc.repo.BeginTransaction() // create Tx for using

	// отложенная функция завершения транзакции
	svc.CloseTx(tx, err, "Creating")

	if err != nil {
		return 0, fmt.Errorf("error create employee:  error creating transaction: %w", err)
	}

	// выполняем несколько запросов в базе данных
	isExist, err := svc.repo.FindByNameTx(tx, request.Name)
	if err != nil {
		return 0, fmt.Errorf("error finding Employee by Name: %s, %w", request.Name, err)
	}
	if isExist {
		return 0, common.AlreadyExistsError{
			fmt.Sprintf("employee with name %s already exists", request.Name),
		}
	}

	var entity = request.ToEntity()
	createdEmployeeId, err := svc.repo.CreateEntityTx(tx, entity)
	if err != nil {
		return 0, fmt.Errorf("error creating Employee whith Name: %s, %w", request.Name, err)
	}
	return createdEmployeeId, err
}

func (svc *Service) FindEmployeeByNameTx(name string) (isExists bool, err error) {
	tx, err := svc.repo.BeginTransaction() // create Tx for using

	// отложенная функция завершения транзакции
	svc.CloseTx(tx, err, "Finding")

	if err != nil {
		return false, fmt.Errorf("error finding transaction: %w", err)
	}

	isExists, err = svc.repo.FindByNameTx(tx, name)
	if err != nil {
		return isExists, fmt.Errorf("error checking existing Employee by Name: %v, %w", isExists, err)
	}
	return isExists, err
}

// Отложенная функция завершения транзакции
func (svc *Service) CloseTx(tx *sqlx.Tx, err error, value string) {
	// отложенная функция завершения транзакции
	defer func() {
		// проверяем, не было ли паники
		if r := recover(); r != nil {
			err = fmt.Errorf("%s employee panic: %v", value, r)
			errTx := tx.Rollback() // если была паника, то откатываем транзакцию
			if errTx != nil {
				err = fmt.Errorf("%s employee: rolling back transaction errors: %w, %w", value, err, errTx)
			}
		} else if err != nil {
			errTx := tx.Rollback() // если произошла другая ошибка (не паника), то откатываем транзакцию
			if errTx != nil {
				err = fmt.Errorf("%s employee: rolling back transaction errors: %w, %w", value, err, errTx)
			}
		} else {
			errTx := tx.Commit() // если ошибок нет, то коммитим транзакцию
			if errTx != nil {
				err = fmt.Errorf("%s employee: commiting transaction error: %w", value, errTx)
			}
		}
	}()
}
