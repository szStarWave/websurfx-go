package yahoo

import "github.com/szStarWave/websurfx-go/internal/engine/htmlengine"

func New() htmlengine.Engine {
	return htmlengine.New(htmlengine.Config{
		Name:           "yahoo",
		BaseURL:        "https://search.yahoo.com",
		Referer:        "https://search.yahoo.com/",
		PathFormat:     "/search?p=%s&b=%d",
		PageValue:      func(page int) int { return (page-1)*10 + 1 },
		ResultSelector: ".algo, .Sr",
		TitleSelector:  "h3 a[href], a[href]",
		DescSelector:   ".compText, .fc-falcon, .lh-16",
		EmptySelector:  ".NoResults",
	})
}
