package domain

//Структуры кастомных ошибок - Доменные ошибки
//Доменные ошибки отдельно от транспортных

// RequestValidationError - ошибка валидации запроса
type RequestValidationError struct {
	Message string
}

func (err RequestValidationError) Error() string {
	return err.Message
}

// AlreadyExistsError - ошибка, когда объект уже существует
type AlreadyExistsError struct {
	Message string
}

func (err AlreadyExistsError) Error() string {
	return err.Message
}
