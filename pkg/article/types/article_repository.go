package article_types

import (
	"context"

	"github.com/google/uuid"
	"github.com/samber/mo"
)

type ArticleRepository interface {
	UpsertArticle(ctx context.Context, article Article) (Article, error)
	GetArticleById(ctx context.Context, id uuid.UUID) (mo.Option[Article], error)
	ListArticles(ctx context.Context,
		limit int,
		offset int,
		authorUserId mo.Option[uuid.UUID],
		favoritedByUserId mo.Option[uuid.UUID],
		tag mo.Option[string],
	) ([]Article, error)
	// ListArticleFeed returns articles created by authors the user follows
	ListArticleFeed(ctx context.Context,
		userId uuid.UUID,
		limit int,
		offset int,
	) ([]Article, error)
	DeleteArticle(ctx context.Context, id uuid.UUID) error
}
