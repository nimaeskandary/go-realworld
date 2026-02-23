package internal_test

import (
	"testing"
	"time"

	"github.com/nimaeskandary/go-realworld/pkg/test_utils/fixtures"
	"github.com/nimaeskandary/go-realworld/pkg/test_utils/helpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_PostgresArticleRepo(t *testing.T) {
	t.Parallel()

	f := fixtures.SetupStandardFixture(t)

	t.Run("UpsertArticle", func(t *testing.T) {
		t.Parallel()
		user := helpers.CreateUsers(t, f.UserService, 1)[0]

		t.Run("should insert a new article", func(t *testing.T) {
			t.Parallel()

			article := helpers.GenArticle(user.Id)
			created, err := f.ArticleRepo.UpsertArticle(t.Context(), article)
			assert.NoError(t, err)
			assert.Equal(t, article, created)
		})

		t.Run("should update an existing article", func(t *testing.T) {
			t.Parallel()
			article := helpers.GenArticle(user.Id)
			_, err := f.ArticleRepo.UpsertArticle(t.Context(), article)
			assert.NoError(t, err)

			updatedArticle := article
			updatedArticle.Title = "Updated Title"
			updatedArticle.Description = "Updated Description"
			// this should not update
			updatedArticle.CreatedAtMillis = time.Now().UnixMilli()

			expectedUpdatedArticle := updatedArticle
			expectedUpdatedArticle.CreatedAtMillis = article.CreatedAtMillis

			actual, err := f.ArticleRepo.UpsertArticle(t.Context(), updatedArticle)
			assert.NoError(t, err)
			assert.Equal(t, expectedUpdatedArticle, actual)
		})
	})

	t.Run("GetArticleById", func(t *testing.T) {
		t.Parallel()
		user := helpers.CreateUsers(t, f.UserService, 1)[0]

		t.Run("should get article by id", func(t *testing.T) {
			t.Parallel()

			article := helpers.GenArticle(user.Id)
			created, err := f.ArticleRepo.UpsertArticle(t.Context(), article)
			assert.NoError(t, err)

			fromDb, err := f.ArticleRepo.GetArticleById(t.Context(), article.Id)
			assert.NoError(t, err)
			assert.True(t, fromDb.IsSome())
			assert.Equal(t, created, fromDb.MustGet())
		})

		t.Run("should return none for non-existent id", func(t *testing.T) {
			t.Parallel()

			fromDb, err := f.ArticleRepo.GetArticleById(t.Context(), uuid.New())
			assert.NoError(t, err)
			assert.True(t, fromDb.IsNone())
		})
	})

	t.Run("DeleteArticle", func(t *testing.T) {
		t.Parallel()
		user := helpers.CreateUsers(t, f.UserService, 1)[0]
		article := helpers.GenArticle(user.Id)
		_, err := f.ArticleRepo.UpsertArticle(t.Context(), article)
		require.NoError(t, err)

		t.Run("should delete existing article", func(t *testing.T) {
			t.Parallel()

			err := f.ArticleRepo.DeleteArticle(t.Context(), article.Id)
			assert.NoError(t, err)

			fromDb, err := f.ArticleRepo.GetArticleById(t.Context(), article.Id)
			assert.NoError(t, err)
			assert.True(t, fromDb.IsNone())
		})

		t.Run("should return no error when deleting non-existing article", func(t *testing.T) {
			t.Parallel()

			err := f.ArticleRepo.DeleteArticle(t.Context(), uuid.New())
			assert.NoError(t, err)
		})
	})
}
