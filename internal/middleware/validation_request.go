package middleware

import (
	"fmt"

	"github.com/wicaker/user/internal/domain"

	"gopkg.in/go-playground/validator.v9"
)

// Validate will validate an incoming request
func Validate(m interface{}) (bool, []domain.Validation) {
	var (
		response = []domain.Validation{}
		validate = validator.New()
		err      = validate.Struct(m)
	)

	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			errorResponse := domain.Validation{
				Field:     err.Field(),
				Tag:       err.Tag(),
				ActualTag: err.ActualTag(),
				Kind:      err.Kind().String(),
				Type:      err.Type().String(),
				Value:     fmt.Sprintf("%v", err.Value()),
				Param:     err.Param(),
				Message:   fmt.Sprintf("Field validation for %v failed, on the '%v' tag", err.Field(), err.Tag()),
			}
			response = append(response, errorResponse)
		}
		return false, response
	}

	return true, response
}
