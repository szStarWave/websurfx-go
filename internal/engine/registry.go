package engine

import (
	"fmt"
	"strings"

	"github.com/szStarWave/websurfx-go/internal/engine/bing"
	"github.com/szStarWave/websurfx-go/internal/engine/so360"
	"github.com/szStarWave/websurfx-go/internal/engine/sogou"
	"github.com/szStarWave/websurfx-go/internal/engine/wikipedia"
	"github.com/szStarWave/websurfx-go/internal/search"
)

func Build(names []string) ([]search.Engine, error) {
	engines := make([]search.Engine, 0, len(names))
	for _, name := range names {
		switch strings.ToLower(strings.TrimSpace(name)) {
		case "bing":
			engines = append(engines, bing.New())
		case "so360", "360":
			engines = append(engines, so360.New())
		case "sogou":
			engines = append(engines, sogou.New())
		case "zhwikipedia", "zh-wikipedia", "wikipedia-zh":
			engines = append(engines, wikipedia.NewZH())
		default:
			return nil, fmt.Errorf("unknown engine %q", name)
		}
	}
	return engines, nil
}
