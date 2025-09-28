package news

import (
	"sort"
	"strings"

	c "nned/internal/common"
	"nned/internal/ui/component/news/row"
	"nned/internal/ui/util"

	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	SetArticlesMsg    []c.Article
	SetFilterMsg      string
	UpdateArticlesMsg []c.Article
)

type Model struct {
	width     int
	cursor    int
	articles  []*c.Article
	rows      []*row.Model
	amap      map[uint64]c.Article
	paginator paginator.Model
}

func NewModel() *Model {
	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = 15
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")).Render("•")

	return &Model{
		width:     80,
		articles:  make([]*c.Article, 0),
		amap:      make(map[uint64]c.Article),
		cursor:    0,
		paginator: p,
	}
}
func (m *Model) Init() tea.Cmd { return nil }
func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case SetArticlesMsg:
		cmds := make([]tea.Cmd, 0)
		articles := make([]*c.Article, 0)

		for i := range msg {
			hash := util.GetHash(msg[i].Title + msg[i].Source)
			if _, ok := m.amap[hash]; !ok {
				articles = append(articles, &msg[i])
				m.amap[hash] = msg[i]
			}
		}

		sort.Sort(c.ByDate(articles))

		for i, article := range articles {
			if i < len(m.rows) {
				m.rows[i], cmd = m.rows[i].Update(row.UpdateArticleMsg(article))
				cmds = append(cmds, cmd)
			} else {
				m.rows = append(m.rows, row.New(row.Config{
					Article: article,
					Width:   m.width,
				}))
			}
		}

		m.articles = articles
		m.paginator.SetTotalPages(len(articles))
		return m, tea.Batch(cmds...)
	case SetFilterMsg:
		if msg == "" {
			for i := range m.rows {
				m.rows[i].Show = true
			}
			return m, nil
		}
		for i := range m.rows {
			if strings.Contains(strings.ToLower(m.rows[i].Config.Article.Title), strings.ToLower(string(msg))) {
				m.rows[i].Show = true
			} else {
				m.rows[i].Show = false
			}
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			prev_c := m.cursor
			m.cursor--
			if m.cursor < 0 {
				m.cursor = 0
			}
			if prev_c != m.cursor {
				m.rows[m.cursor], cmd = m.rows[m.cursor].Update(row.SetBoldMsg(true))
				m.rows[prev_c], cmd = m.rows[prev_c].Update(row.SetBoldMsg(false))
			}
			return m, nil
		case "down":
			prev_c := m.cursor
			m.cursor++
			if m.cursor > len(m.rows)-1 {
				m.cursor = len(m.rows) - 1
			}
			if prev_c != m.cursor {
				m.rows[m.cursor], cmd = m.rows[m.cursor].Update(row.SetBoldMsg(true))
				m.rows[prev_c], cmd = m.rows[prev_c].Update(row.SetBoldMsg(false))
			}
			return m, nil
		case "enter":
			m.rows[m.cursor], cmd = m.rows[m.cursor].Update(row.SetReadMsg{})
			return m, cmd
		case "m":
			m.rows[m.cursor], cmd = m.rows[m.cursor].Update(row.ToggleReadMsg{})
			return m, cmd
		case "h":
			if m.paginator.Page-1 >= 0 {
				m.cursor = (m.paginator.Page - 1) * m.paginator.PerPage
			}
			m.paginator, cmd = m.paginator.Update(msg)
			return m, cmd
		case "l":
			if m.paginator.Page+1 < m.paginator.TotalPages {
				m.cursor = (m.paginator.Page + 1) * m.paginator.PerPage
			}
			m.paginator, cmd = m.paginator.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		for i, a := range m.rows {
			m.rows[i], _ = a.Update(row.SetCellWidthMsg{
				Width: msg.Width,
			})
		}
		return m, nil
	case row.FrameMsg:
		var cmd tea.Cmd
		cmds := make([]tea.Cmd, 0)
		for i, r := range m.rows {
			m.rows[i], cmd = r.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}
	//	fmt.Printf("%#v\n", msg)
	return m, cmd
}

func (m *Model) SetWidth(width int) {
	m.width = width
}

func (m *Model) GetPageCount() int {
	return m.paginator.TotalPages
}

func (m *Model) GetRowArticle(i int) *c.Article {
	return m.rows[i].Config.Article
}

func (m *Model) GetPerPage() int {
	return m.paginator.PerPage
}

func (m *Model) GetCurrentPage() int {
	return m.paginator.Page
}

func (m *Model) SetCurrentPage(page int) {
	m.paginator.Page = page
}

func (m *Model) View() string {
	if m.width < 50 {
		return "Terminal window too narrow to render news feed.\nResize to fix"
	}
	var b strings.Builder
	start, end := m.paginator.GetSliceBounds(len(m.rows))
	for _, row := range m.rows[start:end] {
		b.WriteString(row.View() + "\n")
	}
	b.WriteString(m.paginator.View())

	//	for _, row := range m.rows {
	//		rows = append(rows, row.View())
	//	}
	//	return strings.Join(rows, "\n")
	return b.String()
}
