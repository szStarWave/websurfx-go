package brave

import "github.com/szStarWave/websurfx-go/internal/engine/htmlengine"

func New() htmlengine.Engine {
	return htmlengine.New(htmlengine.Config{
		Name:            "brave",
		BaseURL:         "https://search.brave.com",
		Referer:         "https://search.brave.com/",
		PathFormat:      "/search?q=%s&offset=%d",
		PageParamOffset: -1,
		ResultSelector:  ".snippet, .web-result, [data-type='web']",
		TitleSelector:   "a.heading-serpresult, .title a, a[href]",
		DescSelector:    ".snippet-description, .description, .snippet-content",
		EmptySelector:   ".no-results",
	})
}
