package user_types

import (
	"errors"
	"fmt"
)

type DomainError interface {
	// unexported method keeps this sealed to the package
	sealed()
	error
}

func AsDomainError(err error) DomainError {
	if de, ok := errors.AsType[DomainError](err); ok {
		return de
	}
	return UnknownError{Err: err}
}

type UnknownError struct {
	Err error
}

func (e UnknownError) sealed() {}
func (e UnknownError) Error() string {
	return fmt.Errorf("UnknownError: unknown user domain error: %w", e.Err).Error()
}

type NotFoundError struct {
	Identifier string
}

func (e NotFoundError) sealed() {}
func (e NotFoundError) Error() string {
	return fmt.Sprintf("NotFoundError: could not find user with identifier: %v", e.Identifier)
}

type ConflictError struct {
	Msg string
}

func (e ConflictError) sealed() {}
func (e ConflictError) Error() string {
	return fmt.Sprintf("ConflictError: user conflict error: %v", e.Msg)
}

type BadParamsError struct {
	Err error
}

func (e BadParamsError) sealed() {}
func (e BadParamsError) Error() string {
	return fmt.Errorf("BadParamsError: user params validations error: %w", e.Err).Error()
}

type CannotFollowYourselfError struct{}

func (e CannotFollowYourselfError) sealed() {}
func (e CannotFollowYourselfError) Error() string {
	return "CannotFollowYourselfError: cannot follow yourself"
}
