package news

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	c "nned/internal/common"
	"nned/internal/ui/component/news/row"
	"nned/internal/ui/util"

	tea "github.com/charmbracelet/bubbletea"
)

type SetArticlesMsg []c.Article

type UpdateArticlesMsg []c.Article

type Model struct {
	width    int
	cursor   int
	articles []*c.Article
	rows     []*row.Model
	amap     map[uint64]c.Article
	mu       sync.RWMutex
}

func NewModel() *Model {
	return &Model{
		width:    80,
		articles: make([]*c.Article, 0),
		amap:     make(map[uint64]c.Article),
		cursor:   0,
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
		return m, tea.Batch(cmds...)
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
	fmt.Printf("%#v\n", msg)
	return m, nil
}

func (m *Model) SetWidth(width int) {
	m.width = width
}

func (m *Model) View() string {
	if m.width < 50 {
		return "Terminal window too narrow to render news feed.\nResize to fix"
	}
	rows := make([]string, 0)
	for _, row := range m.rows {
		rows = append(rows, row.View())
	}
	return strings.Join(rows, "\n")
}
