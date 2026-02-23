package auth_types

import (
	"errors"
	"fmt"
)

type DomainError interface {
	sealed()
	Error() string
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
	return fmt.Errorf("unknown auth domain error: %w", e.Err).Error()
}

type InvalidTokenError struct {
	Reason string
}

func (e InvalidTokenError) sealed() {}
func (e InvalidTokenError) Error() string {
	return "invalid auth token: " + e.Reason
}

type ExpiredTokenError struct {
}

func (e ExpiredTokenError) sealed() {}
func (e ExpiredTokenError) Error() string {
	return "expired auth token"
}
