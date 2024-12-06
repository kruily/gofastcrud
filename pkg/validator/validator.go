package validator

import (
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Validate 验证结构体
func Validate(obj interface{}) error {
	if err := validate.Struct(obj); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return err
		}
		return err
	}
	return nil
}

// ValidateVar 验证单个变量
func ValidateVar(field interface{}, tag string) error {
	return validate.Var(field, tag)
}
