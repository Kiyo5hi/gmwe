package gmwe

import "net/http"

type GMWEException interface {
	Status() int
	Error() string
}

type NotFoundException struct {
	Message string
}

type ResourceConflictException struct {
	Message string
}

type InvalidArgumentException struct {
	Message string
}

func (e NotFoundException) Error() string {
	return e.Message
}

func (NotFoundException) Status() int {
	return http.StatusNotFound
}

func (e ResourceConflictException) Error() string {
	return e.Message
}

func (ResourceConflictException) Status() int {
	return http.StatusConflict
}

func (e InvalidArgumentException) Error() string {
	return e.Message
}

func (InvalidArgumentException) Status() int {
	return http.StatusBadRequest
}
