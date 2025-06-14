package employee

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

type Service struct {
	repo Repo
}

type Repo interface {
	FindById(id int64) (Entity, error)
	FindAllEmployees() ([]Entity, error)
	FindAllEmployeesByIds(ids []int64) ([]Entity, error)
	FindByNameTx(tx *sqlx.Tx, name string) (isExists bool, err error)
	CreateEmployee(entity *Entity) (Entity, error)
	CreateEntityTx(tx *sqlx.Tx, entity *Entity) (Entity, error)
	UpdateEmployee(entity *Entity) error
	DeleteEmployeeById(id int64) error
	DeleteAllEmployeesByIds(ids []int64) error
	BeginTransaction() (tx *sqlx.Tx, err error)
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
	var entityRsl, err = svc.repo.CreateEmployee(entity)
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

func (svc *Service) CreateEmployeeTx(entity *Entity) (employee Response, err error) {
	tx, err := svc.repo.BeginTransaction() // create Tx for using

	// отложенная функция завершения транзакции
	svc.closeTx(tx, err, "Creating")

	if err != nil {
		return Response{}, fmt.Errorf("error creating transaction: %w", err)
	}

	var createdEmployee = Entity{}

	// выполняем несколько запросов в базе данных
	isExistsEmployee, err := svc.repo.FindByNameTx(tx, entity.Name)
	if err != nil {
		return Response{}, fmt.Errorf("error finding Employee by Name: %s, %w", entity.Name, err)
	}
	if !isExistsEmployee {
		createdEmployee, err = svc.repo.CreateEntityTx(tx, entity)
	}
	if err != nil {
		return Response{}, fmt.Errorf("error creating Employee whith Name: %s, %w", entity.Name, err)
	}
	return createdEmployee.ToResponse(), err
}

func (svc *Service) FindEmployeeByNameTx(name string) (isExists bool, err error) {
	tx, err := svc.repo.BeginTransaction() // create Tx for using

	// отложенная функция завершения транзакции
	svc.closeTx(tx, err, "Finding")

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
func (svc *Service) closeTx(tx *sqlx.Tx, err error, value string) {
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
