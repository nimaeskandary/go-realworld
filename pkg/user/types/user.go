package user_types

import (
	"github.com/google/uuid"
	"github.com/samber/mo"
)

type User struct {
	Id              uuid.UUID `validate:"required"`
	Username        string    `validate:"required"`
	Email           string    `validate:"required,email"`
	Bio             mo.Option[string]
	Image           mo.Option[string]
	CreatedAtMillis int64 `validate:"required"`
	UpdatedAtMillis int64 `validate:"required"`
}
