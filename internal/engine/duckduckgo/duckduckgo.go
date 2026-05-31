package duckduckgo

import "github.com/szStarWave/websurfx-go/internal/engine/htmlengine"

func New() htmlengine.Engine {
	return htmlengine.New(htmlengine.Config{
		Name:           "duckduckgo",
		BaseURL:        "https://html.duckduckgo.com",
		Referer:        "https://html.duckduckgo.com/",
		PathFormat:     "/html/?q=%s&s=%d",
		PageValue:      func(page int) int { return (page - 1) * 30 },
		ResultSelector: ".result",
		TitleSelector:  ".result__title a, a.result__a",
		DescSelector:   ".result__snippet",
		EmptySelector:  ".no-results",
	})
}
