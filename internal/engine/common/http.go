package common

import (
	"context"
	"io"
	"net/http"

	"github.com/szStarWave/websurfx-go/internal/search"
)

func Get(ctx context.Context, client *http.Client, url, referer string) ([]byte, *search.EngineError) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, &search.EngineError{Type: search.ErrorUnexpected, Detail: err.Error()}
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0 Safari/537.36")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, &search.EngineError{Type: search.ErrorRequest, Detail: err.Error()}
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &search.EngineError{Type: search.ErrorRequest, Detail: resp.Status}
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &search.EngineError{Type: search.ErrorRequest, Detail: err.Error()}
	}
	return body, nil
}

func WithEngine(engine string, err *search.EngineError) *search.EngineError {
	if err == nil {
		return nil
	}
	err.Engine = engine
	return err
}
