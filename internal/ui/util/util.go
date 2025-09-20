package util

import (
	"hash/fnv"
	"strings"

	te "github.com/muesli/termenv"
	"golang.org/x/net/html"
)

var p = te.ColorProfile()

func extractText(n *html.Node, sb *strings.Builder) {
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			sb.WriteString(text)
			sb.WriteString(" ")
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractText(c, sb)
	}
}

func GetStringFromHTML(s string) (string, error) {
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	extractText(doc, &sb)
	return sb.String(), nil
}

func StyleSource(s string, fg string, bg string, bold bool) string {
	return NewStyle(fg, bg, bold)(s)
}

func NewStyle(fg string, bg string, bold bool) func(string) string {
	s := te.Style{}.Foreground(p.Color(fg)).Background(p.Color(bg))
	if bold {
		s = s.Bold()
	}
	return s.Styled
}

func GetHash(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}
