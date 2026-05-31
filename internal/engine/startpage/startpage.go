package startpage

import "github.com/szStarWave/websurfx-go/internal/engine/htmlengine"

func New() htmlengine.Engine {
	return htmlengine.New(htmlengine.Config{
		Name:            "startpage",
		BaseURL:         "https://www.startpage.com",
		Referer:         "https://www.startpage.com/",
		PathFormat:      "/sp/search?query=%s&page=%d",
		PageParamOffset: -1,
		ResultSelector:  ".w-gl__result, .result",
		TitleSelector:   "a.w-gl__result-title, .result-title a, a[href]",
		DescSelector:    ".w-gl__description, .description, p",
		EmptySelector:   ".no-results",
	})
}
