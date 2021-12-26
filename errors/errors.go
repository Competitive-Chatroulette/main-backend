package errors

import (
	"fmt"
	"net/http"
)

type CError interface {
	GetStatusCode() int
	Error() string
}

type NotUnique struct {
	StatusCode int
	Field      string
}

func (e NotUnique) GetStatusCode() int {
	return e.StatusCode
}
func (e NotUnique) Error() string {
	return fmt.Sprintf("%s already taken", e.Field)
}
func NewNotUnique(field string) NotUnique {
	return NotUnique{
		StatusCode: http.StatusConflict,
		Field:      field,
	}
}

type Internal struct {
	StatusCode int
}

func (e Internal) GetStatusCode() int {
	return e.StatusCode
}
func (e Internal) Error() string {
	return fmt.Sprintf("Unexpected error occured")
}
func NewInternal() NotUnique {
	return NotUnique{
		StatusCode: http.StatusInternalServerError,
	}
}

type NotFound struct {
	StatusCode int
	Field      string
}

func (e NotFound) GetStatusCode() int {
	return e.StatusCode
}
func (e NotFound) Error() string {
	return fmt.Sprintf("This %s doesn't exist", e.Field)
}
func NewNotFound(field string) NotFound {
	return NotFound{
		StatusCode: http.StatusNotFound,
		Field:      field,
	}
}
