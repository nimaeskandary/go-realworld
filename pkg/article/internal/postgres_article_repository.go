package internal

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	article_types "github.com/nimaeskandary/go-realworld/pkg/article/types"
	db_types "github.com/nimaeskandary/go-realworld/pkg/database/types"
	obs_types "github.com/nimaeskandary/go-realworld/pkg/observability/types"
	"github.com/samber/mo"

	"github.com/google/uuid"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/scan"
)

const (
	articlesTableName = "articles"
)

type postgresArticleRepo struct {
	db     db_types.PostgresRealWorldAppDb
	logger obs_types.Logger
}

func NewPostgresArticleRepository(db db_types.PostgresRealWorldAppDb, logger obs_types.Logger) article_types.ArticleRepository {
	return &postgresArticleRepo{db: db, logger: logger}
}

// UpsertArticle implements [article_types.ArticleRepository.UpsertArticle]
func (r *postgresArticleRepo) UpsertArticle(ctx context.Context, article article_types.Article) (article_types.Article, error) {
	data := map[string]any{
		"title":       article.Title,
		"description": article.Description,
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return article_types.Article{}, fmt.Errorf("error marshalling article data for upsert, article=%v: %w", article, err)
	}

	createCols := []string{"id", "author_user_id", "data", "created_at", "updated_at"}
	updateCols := []string{"data", "updated_at"}

	q := psql.Insert(
		im.Into(articlesTableName, createCols...),
		im.Values(
			psql.Arg(article.Id.String()),
			psql.Arg(article.AuthorUserId.String()),
			psql.Arg(dataBytes),
			psql.Arg(time.UnixMilli(article.CreatedAtMillis)),
			psql.Arg(time.UnixMilli(article.UpdatedAtMillis)),
		),
		im.OnConflict("id").DoUpdate(
			im.SetExcluded(updateCols...),
		),
		im.Returning("*"),
	)

	result, err := bob.One(ctx, bob.NewDB(r.db.GetDB()), q, scan.StructMapper[postgresArticle]())
	if err != nil {
		return article_types.Article{}, fmt.Errorf("error with article upsert query, article=%v: %w", article, err)
	}

	asArticle, err := fromPostgresArticle(result)
	if err != nil {
		return article_types.Article{}, fmt.Errorf("error converting postgres article to domain article: %w", err)
	}

	return asArticle, nil
}

// GetArticleById implements [types.ArticleRepository.GetArticleById].
func (r *postgresArticleRepo) GetArticleById(ctx context.Context, id uuid.UUID) (mo.Option[article_types.Article], error) {
	q := psql.Select(
		sm.Columns("*"),
		sm.From(articlesTableName),
		sm.Where(psql.Quote("id").EQ(psql.Arg(id.String()))),
	)

	result, err := bob.One(ctx, bob.NewDB(r.db.GetDB()), q, scan.StructMapper[postgresArticle]())
	if err != nil {
		if err == sql.ErrNoRows {
			return mo.None[article_types.Article](), nil
		}
		return mo.None[article_types.Article](), fmt.Errorf("error with get article by id query, id=%v: %w", id, err)
	}

	asArticle, err := fromPostgresArticle(result)
	if err != nil {
		return mo.None[article_types.Article](), fmt.Errorf("error converting postgres article to domain article: %w", err)
	}

	return mo.Some(asArticle), nil
}

// ListArticles implements [types.ArticleRepository.ListArticles].
func (r *postgresArticleRepo) ListArticles(ctx context.Context, limit int, offset int, authorUserId mo.Option[uuid.UUID], favoritedByUserId mo.Option[uuid.UUID], tag mo.Option[string]) ([]article_types.Article, error) {
	panic("unimplemented")
}

// ListArticleFeed implements [types.ArticleRepository.ListArticleFeed].
func (r *postgresArticleRepo) ListArticleFeed(ctx context.Context, userId uuid.UUID, limit int, offset int) ([]article_types.Article, error) {
	panic("unimplemented")
}

// DeleteArticle implements [types.ArticleRepository.DeleteArticle].
func (r *postgresArticleRepo) DeleteArticle(ctx context.Context, id uuid.UUID) error {
	q := psql.Delete(
		dm.From(articlesTableName),
		dm.Where(psql.Quote("id").EQ(psql.Arg(id.String()))),
	)

	_, err := bob.Exec(ctx, bob.NewDB(r.db.GetDB()), q)
	if err != nil {
		return fmt.Errorf("error with delete article query, id=%v: %w", id, err)
	}

	return nil
}

type postgresArticle struct {
	Id           uuid.UUID `db:"id"`
	AuthorUserId uuid.UUID `db:"author_user_id"`
	Data         []byte    `db:"data"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// fromPostgresArticle converts postgres article into an article_types.Article
func fromPostgresArticle(from postgresArticle) (article_types.Article, error) {
	var data struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if len(from.Data) > 0 {
		if err := json.Unmarshal(from.Data, &data); err != nil {
			return article_types.Article{}, fmt.Errorf("error unmarshalling article data json: %w", err)
		}
	}

	return article_types.Article{
		Id:              from.Id,
		AuthorUserId:    from.AuthorUserId,
		Title:           data.Title,
		Description:     data.Description,
		CreatedAtMillis: from.CreatedAt.UnixMilli(),
		UpdatedAtMillis: from.UpdatedAt.UnixMilli(),
	}, nil
}
