package ui

import (
	"fmt"
	"slices"
	"sync"
	"time"

	"nned/internal/ui/component/article"
	"nned/internal/ui/component/news"
	"nned/internal/ui/component/news/row"
	"nned/internal/ui/util"

	c "nned/internal/common"
	mon "nned/internal/monitor"

	grid "github.com/achannarasappa/term-grid"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	articles       []c.Article
	news           *news.Model
	article        *article.Model
	cursor         int
	ctx            c.Context
	viewport       viewport.Model
	ready          bool
	lastUpdateTime string
	headerHeight   int
	monitor        *mon.Monitor
	mu             sync.RWMutex
	versionVector  int
}

type SetArticleMsg struct {
	article       c.Article
	versionVector int
}

type tickMsg struct {
	versionVector int
}

var (
	styleLogo = util.NewStyle("#111111", "#ff8700", true)
	styleHelp = util.NewStyle("#4e4e4e", "", false)
)

const (
	footerHeight   = 1
	minFooterWidth = 80
)

func NewModel(dep c.Dependencies, ctx c.Context, monitor *mon.Monitor) *Model {
	return &Model{
		articles:     make([]c.Article, 0),
		cursor:       0,
		ctx:          ctx,
		ready:        false,
		news:         news.NewModel(),
		article:      article.NewModel(),
		headerHeight: 0,
		monitor:      monitor,
	}
}

func (m *Model) Init() tea.Cmd {
	(*m.monitor).Start()

	return tea.Batch(
		tick(0),
		func() tea.Msg {
			return nil
		},
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			fallthrough
		case "esc":
			fallthrough
		case "q":
			return m, tea.Quit
		case "up":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.news, cmd = m.news.Update(msg)
			return m, cmd
		case "down":
			m.cursor++
			if m.cursor > len(m.articles)-1 {
				m.cursor = len(m.articles) - 1
			}
			m.news, cmd = m.news.Update(msg)
			return m, cmd
		case "enter":
			m.article, cmd = m.article.Update(article.SetArticleMsg(&m.articles[m.cursor]))
			m.news, cmd = m.news.Update(msg)
			return m, cmd
		case "m":
			m.news, cmd = m.news.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		var cmd tea.Cmd
		viewportHeight := msg.Height - m.headerHeight - footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, viewportHeight)
			m.ready = true
			m.news.SetWidth(msg.Width / 2)
			m.article.SetDimensions(msg.Width/2, viewportHeight)
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = viewportHeight
		}
		return m, cmd

	case tickMsg:
		cmds := make([]tea.Cmd, 0)

		m.news, cmd = m.news.Update(news.SetArticlesMsg(m.articles))
		cmds = append(cmds, cmd)
		m.lastUpdateTime = getTime()

		if m.ready {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
		cmds = append(cmds, tick(m.versionVector))
		return m, tea.Batch(cmds...)
	case SetArticleMsg:
		if msg.versionVector != m.versionVector {
			return m, nil
		}

		m.articles = append(m.articles, msg.article)
		slices.SortFunc(m.articles, util.DateCmp)
		return m, nil

	case row.FrameMsg:
		var cmd tea.Cmd
		m.news, cmd = m.news.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *Model) View() string {
	if !m.ready {
		return "\n Fetching articles..."
	}
	content := "Loading content..."
	content = lipgloss.JoinHorizontal(lipgloss.Top, m.news.View(), m.article.View())
	m.viewport.SetContent(content)

	return m.viewport.View() + "\n" +
		footer(m.viewport.Width, m.lastUpdateTime)
}

func tick(t int) tea.Cmd {
	return tea.Tick(time.Second/5, func(time.Time) tea.Msg {
		return tickMsg{
			versionVector: t,
		}
	})
}

func footer(width int, time string) string {
	if width < minFooterWidth {
		return "nned"
	}
	help := "q: exit ↑: scroll up ↓: scroll down s: search m: mark read/unread enter: read article"
	return grid.Render(grid.Grid{
		Rows: []grid.Row{
			{
				Width: width,
				Cells: []grid.Cell{
					{Text: styleLogo.Render(" nned "), Width: 7},
					{Text: styleHelp.Render(help), Width: len(help)},
					{Text: styleHelp.Render("T: " + time), Align: grid.Right},
				},
			},
		},
	})
}

func getTime() string {
	t := time.Now()

	return fmt.Sprintf("%s %02d:%02d", t.Weekday().String(), t.Hour(), t.Minute())
}
