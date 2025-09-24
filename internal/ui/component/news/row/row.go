package row

import (
	"fmt"
	"sync/atomic"
	"time"

	c "nned/internal/common"
	"nned/internal/ui/util"

	grid "github.com/achannarasappa/term-grid"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	HorGutter = 1
	VerGutter = 2
)

var (
	lastID      int64
	titleStyle  = util.NewStyle("#EBEBEB", "", false)
	unreadStyle = util.NewStyle("#FF0000", "", true)
	timeStyle   = util.NewStyle("#666666", "", false)
)

type Model struct {
	id     int
	width  int
	config Config
	bold   bool
	unread bool
}

type Config struct {
	ID      int
	Article *c.Article
	Width   int
}

type UpdateArticleMsg *c.Article

type FrameMsg int

type SetCellWidthMsg struct {
	Width int
}

type (
	ToggleReadMsg struct{}
	SetReadMsg    struct{}
)

type (
	SetBoldMsg bool
)

func New(config Config) *Model {
	var id int

	if config.ID != 0 {
		id = config.ID
	} else {
		id = nextID()
	}
	return &Model{
		id:     id,
		width:  config.Width,
		config: config,
		bold:   false,
		unread: true,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SetCellWidthMsg:
		m.width = msg.Width
		return m, nil
	case SetBoldMsg:
		m.bold = bool(msg)
	case SetReadMsg:
		m.unread = false
	case ToggleReadMsg:
		m.unread = !m.unread
		return m, nil
	case UpdateArticleMsg:
		m.config.Article = msg
		return m, nil
	case FrameMsg:
		return m, nil
	}
	return m, nil
}

func (m *Model) View() string {
	rows := []grid.Row{}
	readStr := "•"
	if !m.unread {
		readStr = ""
	}
	rows = append(rows, grid.Row{
		Width: m.width,
		Cells: []grid.Cell{
			{Text: titleStyle.Bold(!m.bold).Render(m.config.Article.Title), Width: m.width - 4, Overflow: grid.Hidden},
			{Text: unreadStyle.Render(readStr), Width: 4},
		},
	})
	time_s := timeAgo(m.config.Article.Date)
	rows = append(rows, grid.Row{
		Width: m.width,
		Cells: []grid.Cell{
			{Text: lipgloss.NewStyle().Background(lipgloss.Color(m.config.Article.SourceColor)).Render(m.config.Article.SourceTitle), Width: 10, Align: grid.Left, Overflow: grid.Hidden},
			{Text: timeStyle.Render(time_s), Width: m.width - 10, Align: grid.Left},
		},
	})

	border := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#444444")).
		BorderBottom(true)
	rendered_row := grid.Render(grid.Grid{
		Rows: rows,
	})
	return border.Render(rendered_row)
}

func timeAgo(date *time.Time) string {
	if date == nil {
		return "No time"
	}
	diff := time.Since(*date)
	if diff.Minutes() < 60 {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	} else if diff.Hours() < 24 {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	} else {
		return fmt.Sprintf("%d day ago", int(diff.Hours()/24))
	}
}

func nextID() int {
	return int(atomic.AddInt64(&lastID, 1))
}
