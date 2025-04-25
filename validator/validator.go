package validator

import (
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/kruily/gofastcrud/errors"
)

// Validate 验证结构体
func Validate(obj interface{}) error {
	var validate = validator.New()
	if err := validate.Struct(obj); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return err
		}
		return err
	}
	return nil
}

// ValidateMap 验证map
func ValidateMap(fields map[string]interface{}, entity any) error {
	// 获取实体类型的反射信息
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() == reflect.Ptr {
		entityType = entityType.Elem()
	}

	// 验证字段
	for fieldName, value := range fields {
		// 检查字段是否存在
		field, exists := entityType.FieldByName(fieldName)
		if !exists {
			return errors.New(errors.ErrInvalidParam, "invalid field: "+fieldName)
		}

		// 获取字段的验证标签
		validateTag := field.Tag.Get("validate")
		if validateTag != "" {
			if err := ValidateVar(value, validateTag); err != nil {
				return err
			}
		}
	}
	return nil
}

// ValidateVar 验证单个变量
func ValidateVar(field interface{}, tag string) error {
	var validate = validator.New()
	return validate.Var(field, tag)
}
