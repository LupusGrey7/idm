package validator

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"idm/inner/domain"
	"strings"
)

type Validator struct {
	validate *validator.Validate
}

func NewValidator() *Validator {
	validate := validator.New()

	// Регистрируем кастомный валидатор "no_sql_injection"
	// (правильный вызов метода для *validator.Validate)
	_ = validate.RegisterValidation(
		"no_sql_injection",
		func(fl validator.FieldLevel) bool {
			value := fl.Field().String()
			forbiddenChars := []string{";", "'", "\"", "--", "/*", "*/", "\\"}
			for _, char := range forbiddenChars {
				if strings.Contains(value, char) {
					return false // Найден опасный символ
				}
			}
			return true
		},
	)
	return &Validator{validate: validate}
}

func (v *Validator) Validate(request any) error {
	err := v.validate.Struct(request)
	if err != nil {
		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			// Формируем читаемое сообщение об ошибке
			for _, e := range validateErrs {
				switch e.Tag() {
				case "required":
					return domain.RequestValidationError{Message: fmt.Sprintf("Field %s is required", e.Field())}
				case "min":
					return domain.RequestValidationError{Message: fmt.Sprintf("Field %s must be at least %s", e.Field(), e.Param())}
				case "max":
					return domain.RequestValidationError{Message: fmt.Sprintf("Field %s must not exceed %s", e.Field(), e.Param())}
				case "no_sql_injection": // Обработка нового тега
					return domain.RequestValidationError{Message: fmt.Sprintf("Field %s contains forbidden SQL characters", e.Field())}
				default:
					return domain.RequestValidationError{Message: fmt.Sprintf("Field %s is invalid", e.Field())} // Обработка других ошибок
				}
			}
		}
		return domain.RequestValidationError{Message: err.Error()}
	}
	return nil
}

//func (v Validator) Validate(request any) (err error) {
//	err = v.validate.Struct(request)
//	if err != nil {
//		var validateErrs validator.ValidationErrors
//		if errors.As(err, &validateErrs) {
//			return validateErrs
//		}
//	}
//	return err
//}

//func (v Validator) Validate(request any) error {
//	err := v.validate.Struct(request)
//	if err != nil {
//		var validateErrs validator.ValidationErrors
//		if errors.As(err, &validateErrs) {
//			// Преобразуем в кастомную ошибку валидации
//			return domain.RequestValidationError{
//				Message: "Invalid request parameters",
//			}
//		}
//		return err
//	}
//	return nil
//}
