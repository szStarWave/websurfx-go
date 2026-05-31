package qwant

import "github.com/szStarWave/websurfx-go/internal/engine/htmlengine"

func New() htmlengine.Engine {
	return htmlengine.New(htmlengine.Config{
		Name:           "qwant",
		BaseURL:        "https://www.qwant.com",
		Referer:        "https://www.qwant.com/",
		PathFormat:     "/?q=%s&t=web&p=%d",
		ResultSelector: "article, .webResult, [data-testid='webResult']",
		TitleSelector:  "h2 a[href], a[href]",
		DescSelector:   "p, .desc, .description",
		EmptySelector:  ".no-result, .noResults",
	})
}
