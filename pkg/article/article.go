package article

import (
	"github.com/nimaeskandary/go-realworld/pkg/article/internal"
	"github.com/nimaeskandary/go-realworld/pkg/article/types"
	"github.com/nimaeskandary/go-realworld/pkg/util"

	"go.uber.org/fx"
)

func NewArticleModule() fx.Option {
	return util.NewFxModule[article_types.ArticleService](
		"article_service",
		internal.NewArticleServiceImpl,
		fx.Provide(internal.NewPostgresArticleRepository),
	)
}
