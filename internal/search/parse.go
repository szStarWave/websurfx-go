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
