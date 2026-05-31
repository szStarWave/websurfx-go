package searx

import "github.com/szStarWave/websurfx-go/internal/engine/htmlengine"

func New(baseURL string) htmlengine.Engine {
	if baseURL == "" {
		baseURL = "https://searx.be"
	}
	return htmlengine.New(htmlengine.Config{
		Name:           "searx",
		BaseURL:        baseURL,
		Referer:        baseURL + "/",
		PathFormat:     "/search?q=%s&categories=general&pageno=%d",
		ResultSelector: ".result, article.result",
		TitleSelector:  "h3 a[href], .result_header a[href], a[href]",
		DescSelector:   ".content, p",
		EmptySelector:  ".no_results, #results:empty",
	})
}
