package common

//Структуры кастомных ошибок

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
