package internal

import article_types "github.com/nimaeskandary/go-realworld/pkg/article/types"

type articleServiceImpl struct {
	articleRepo article_types.ArticleRepository
}

func NewArticleServiceImpl(articleRepo article_types.ArticleRepository) article_types.ArticleService {
	return &articleServiceImpl{articleRepo: articleRepo}
}
