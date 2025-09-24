package util

import (
	"hash/fnv"
	"strings"

	c "nned/internal/common"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/net/html"
)

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

func NewStyle(fg string, bg string, bold bool) lipgloss.Style {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(fg)).
		Background(lipgloss.Color(bg)).
		Bold(!bold)
	return style
}

func GetHash(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

func DateCmp(a, b c.Article) int {
	if a.Date == b.Date {
		return 0
	} else if a.Date.Before(*b.Date) {
		return 1
	} else {
		return -1
	}
}
