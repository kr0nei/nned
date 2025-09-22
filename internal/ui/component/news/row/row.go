package row

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	c "nned/internal/common"
	"nned/internal/ui/util"

	grid "github.com/achannarasappa/term-grid"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	HorGutter = 1
	VerGutter = 2
)

var (
	lastID           int64
	titleStyle       = util.NewStyle("#EBEBEB", "", false)
	unreadStyle      = util.NewStyle("#FF0000", "", true)
	descriptionStyle = util.NewStyle("#A0A0A0", "", false)
	timeStyle        = util.NewStyle("#666666", "", false)
	lineStyle        = util.NewStyle("#444444", "", false)
)

type Model struct {
	id     int
	width  int
	config Config
	bold   bool
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

type SetBoldMsg struct {
	Bold bool
}

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
		m.bold = msg.Bold
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
	rows = append(rows, grid.Row{
		Width: m.width,
		Cells: []grid.Cell{
			{Text: titleStyle.Bold(!m.bold).Render(m.config.Article.Title) + unreadStyle.Render(" • "), Width: len(m.config.Article.Title) + 4},
		},
	})
	time_s := timeAgo(m.config.Article.Date)
	rows = append(rows, grid.Row{
		Width: m.width,
		Cells: []grid.Cell{
			{Text: m.config.Article.SourceTitle[:min(len(m.config.Article.SourceTitle), 10)], Width: min(len(m.config.Article.SourceTitle), 10), Align: grid.Right},
			{Text: timeStyle.Render(time_s), Width: len(time_s), Align: grid.Left},
		},
	})

	description, err := util.GetStringFromHTML(m.config.Article.Description)
	if err != nil {
		description = ""
	}
	rows = append(rows, grid.Row{
		Width: m.width,
		Cells: []grid.Cell{
			{Text: descriptionStyle.Render(description), Align: grid.Left, Overflow: grid.WrapWord},
		},
	})
	rows = append(rows, grid.Row{
		Width: m.width,
		Cells: []grid.Cell{
			{Text: lineStyle.Render(strings.Repeat("━", m.width))},
		},
	})
	return grid.Render(grid.Grid{
		Rows:             rows,
		GutterHorizontal: 1,
	})
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
