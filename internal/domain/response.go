package domain

import (
	"errors"
	"net/http"

	"github.com/sirupsen/logrus"
)

var (
	// ErrInternalServerError will throw if any the Internal Server Error happen
	ErrInternalServerError = errors.New("Internal Server Error")
	// ErrUnauthorized will throw if the given request-header token is not valid
	ErrUnauthorized = errors.New("Unauthorized")
	// ErrStatusUnprocessableEntity will thrown if the given request-body is not valid
	ErrStatusUnprocessableEntity = errors.New("UnprocessableEntity")
	// ErrEmailAlreadyExist /
	ErrEmailAlreadyExist = errors.New("Email already exist! ")
	// ErrEmailNotFound /
	ErrEmailNotFound = errors.New("Email not found! ")
	// ErrWrongPassword /
	ErrWrongPassword = errors.New("Wrong password! ")
	// ErrUserNotFound /
	ErrUserNotFound = errors.New("User not found! ")
	// ErrUserAlreadyExist /
	ErrUserAlreadyExist = errors.New("User already exist! ")
)

// Response represent response structure of request
type Response struct {
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Errors  []Validation           `json:"errors,omitempty"`
}

// Validation represent structure of validation error
type Validation struct {
	Message   string `json:"message"`
	Field     string `json:"field"`
	Tag       string `json:"tag"`
	ActualTag string `json:"actual_tag"`
	Kind      string `json:"kind"`
	Type      string `json:"type"`
	Value     string `json:"value"`
	Param     string `json:"param"`
}

// GetStatusCode will return status code based on type of error
func GetStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	logrus.Error(err)

	switch err {
	case ErrEmailAlreadyExist:
		return http.StatusConflict
	case ErrUserAlreadyExist:
		return http.StatusConflict
	case ErrEmailNotFound:
		return http.StatusNotFound
	case ErrUserNotFound:
		return http.StatusNotFound
	case ErrWrongPassword:
		return http.StatusForbidden
	case ErrInternalServerError:
		return http.StatusInternalServerError
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrStatusUnprocessableEntity:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}
