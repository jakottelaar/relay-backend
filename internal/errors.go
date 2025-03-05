package internal

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	JWTAuthHeader = "Authorization"
)

type ServiceError struct {
	Code    int
	Message string
	Err     error
}

func (e *ServiceError) Error() string {
	return e.Message
}

func NewNotFoundError(msg string) error {
	return &ServiceError{
		Code:    http.StatusNotFound,
		Message: msg,
		Err:     errors.New(msg),
	}
}

func NewDuplicateError(msg string) error {
	return &ServiceError{
		Code:    http.StatusConflict,
		Message: msg,
		Err:     errors.New(msg),
	}
}

func NewBadRequestError(msg string) error {
	return &ServiceError{
		Code:    http.StatusBadRequest,
		Message: msg,
		Err:     errors.New(msg),
	}
}

func NewUnprocessableEntityError(msg string) error {
	return &ServiceError{
		Code:    http.StatusUnprocessableEntity,
		Message: msg,
		Err:     errors.New(msg),
	}
}

func NewInternalServerError(msg string) error {
	return &ServiceError{
		Code:    http.StatusInternalServerError,
		Message: msg,
		Err:     errors.New(msg),
	}
}

func NewUnauthorizedError(msg string) error {
	return &ServiceError{
		Code:    http.StatusUnauthorized,
		Message: msg,
		Err:     errors.New(msg),
	}
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				// Check if it's a ServiceError
				if serviceErr, ok := e.Err.(*ServiceError); ok {
					c.JSON(serviceErr.Code, gin.H{
						"error": serviceErr.Message,
					})
					return
				}
			}

			// Fallback to generic internal server error
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		}
	}
}
