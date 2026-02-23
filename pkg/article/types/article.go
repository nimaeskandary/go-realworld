package article_types

import (
	"github.com/google/uuid"
)

type Article struct {
	Id              uuid.UUID
	AuthorUserId    uuid.UUID
	Title           string
	Description     string
	CreatedAtMillis int64
	UpdatedAtMillis int64
}
