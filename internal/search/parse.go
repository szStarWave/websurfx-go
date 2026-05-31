package search

import (
	"bytes"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func EncodeQuery(raw string) string {
	return url.QueryEscape(raw)
}

func Text(sel *goquery.Selection) string {
	return strings.Join(strings.Fields(sel.Text()), " ")
}

func TextFromHTML(raw string) string {
	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString("<div>" + raw + "</div>"))
	if err != nil {
		return strings.Join(strings.Fields(raw), " ")
	}
	return Text(doc.Find("div").First())
}

func AbsURL(base, href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}
	parsed, err := url.Parse(href)
	if err != nil {
		return href
	}
	if parsed.IsAbs() {
		return parsed.String()
	}
	root, err := url.Parse(base)
	if err != nil {
		return href
	}
	return root.ResolveReference(parsed).String()
}

func CanonicalURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if target := redirectTarget(parsed); target != "" {
		return CanonicalURL(target)
	}
	parsed.Fragment = ""
	parsed.Host = strings.ToLower(parsed.Host)
	if parsed.Path != "/" {
		parsed.Path = strings.TrimRight(parsed.Path, "/")
	}
	return parsed.String()
}

func redirectTarget(parsed *url.URL) string {
	for _, key := range []string{"url", "u", "target", "to"} {
		if value := parsed.Query().Get(key); value != "" {
			if decoded, err := url.QueryUnescape(value); err == nil {
				value = decoded
			}
			if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
				return value
			}
		}
	}
	return ""
}
