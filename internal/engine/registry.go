package engine

import (
	"fmt"
	"strings"

	"github.com/szStarWave/websurfx-go/internal/engine/bing"
	"github.com/szStarWave/websurfx-go/internal/engine/brave"
	"github.com/szStarWave/websurfx-go/internal/engine/duckduckgo"
	"github.com/szStarWave/websurfx-go/internal/engine/qwant"
	"github.com/szStarWave/websurfx-go/internal/engine/searx"
	"github.com/szStarWave/websurfx-go/internal/engine/so360"
	"github.com/szStarWave/websurfx-go/internal/engine/sogou"
	"github.com/szStarWave/websurfx-go/internal/engine/startpage"
	"github.com/szStarWave/websurfx-go/internal/engine/wikipedia"
	"github.com/szStarWave/websurfx-go/internal/engine/yahoo"
	"github.com/szStarWave/websurfx-go/internal/search"
)

func AllNames() []string {
	return []string{
		"bing",
		"so360",
		"sogou",
		"zhwikipedia",
		"duckduckgo",
		"brave",
		"qwant",
		"startpage",
		"yahoo",
		"searx",
	}
}

func Build(names []string) ([]search.Engine, error) {
	engines := make([]search.Engine, 0, len(names))
	for _, name := range names {
		normalized := strings.ToLower(strings.TrimSpace(name))
		switch normalized {
		case "bing":
			engines = append(engines, bing.New())
		case "duckduckgo", "ddg":
			engines = append(engines, duckduckgo.New())
		case "brave":
			engines = append(engines, brave.New())
		case "qwant":
			engines = append(engines, qwant.New())
		case "startpage":
			engines = append(engines, startpage.New())
		case "yahoo":
			engines = append(engines, yahoo.New())
		case "searx":
			engines = append(engines, searx.New(""))
		case "so360", "360":
			engines = append(engines, so360.New())
		case "sogou":
			engines = append(engines, sogou.New())
		case "zhwikipedia", "zh-wikipedia", "wikipedia-zh":
			engines = append(engines, wikipedia.NewZH())
		default:
			if strings.HasPrefix(normalized, "searx:") {
				engines = append(engines, searx.New(strings.TrimSpace(name[len("searx:"):])))
				continue
			}
			return nil, fmt.Errorf("unknown engine %q", name)
		}
	}
	return engines, nil
}
