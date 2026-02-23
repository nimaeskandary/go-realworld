package helpers

import (
	article_types "github.com/nimaeskandary/go-realworld/pkg/article/types"

	"github.com/google/uuid"
)

func GenArticle(authorUserId uuid.UUID) article_types.Article {
	return article_types.Article{
		Id:              uuid.New(),
		AuthorUserId:    authorUserId,
		Title:           "Test Article",
		Description:     "This is a test article",
		CreatedAtMillis: now.UnixMilli(),
		UpdatedAtMillis: now.UnixMilli(),
	}
}
