package config

import (
	"reflect"

	"github.com/go-playground/validator/v10"

	"github.com/platformsh/cli/internal/version"
)

var _validator *validator.Validate

func getValidator() *validator.Validate {
	if _validator == nil {
		_validator = validator.New()
		initCustomValidators(_validator)
	}
	return _validator
}

func initCustomValidators(v *validator.Validate) {
	_ = v.RegisterValidation("version", func(fl validator.FieldLevel) bool {
		return fl.Field().Kind() == reflect.String && version.Validate(fl.Field().String())
	})
}
