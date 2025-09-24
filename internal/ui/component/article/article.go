package article

import (
	c "nned/internal/common"
	"nned/internal/ui/util"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	article c.Article
	width   int
	height  int
}

type SetArticleMsg *c.Article

func NewModel() *Model {
	return &Model{
		article: c.Article{},
		width:   80,
		height:  80,
	}
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SetArticleMsg:
		m.article = *msg
		return m, nil
	}
	return m, nil
}

func (m *Model) View() string {
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#BBBBBB")).Margin(1)
	descriptionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#DDDDDD")).Margin(1)
	sourceStyle := lipgloss.NewStyle().Background(lipgloss.Color(m.article.SourceColor)).Margin(0, 1)
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#999999"))
	var titleBlock, descriptionBlock, sourceBlock, dateBlock string

	if m.article.Title == "" {
		titleBlock = lipgloss.PlaceHorizontal(m.width, lipgloss.Left, "Select article to read")
		descriptionBlock = ""
		sourceBlock = ""
		dateBlock = ""
	} else {
		titleBlock = lipgloss.PlaceHorizontal(m.width, lipgloss.Left, titleStyle.Render(m.article.Title))
		description, err := util.GetStringFromHTML(m.article.Description)
		if err != nil {
			description = ""
		}
		descriptionBlock = lipgloss.PlaceHorizontal(m.width, lipgloss.Left, descriptionStyle.Render(description))
		sourceBlock = lipgloss.PlaceHorizontal(len(m.article.SourceTitle), lipgloss.Left, sourceStyle.Render(m.article.SourceTitle))
		dateBlock = lipgloss.PlaceHorizontal(m.width/2, lipgloss.Left, timeStyle.Render(m.article.Date.Format("Mon, Jan 2, 2006")))
	}
	content := lipgloss.JoinVertical(lipgloss.Left, titleBlock, lipgloss.JoinHorizontal(lipgloss.Top, sourceBlock, dateBlock), descriptionBlock)
	contentStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Width(m.width - 1).Height(m.height - 2)

	return contentStyle.Render(content)
}

func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}
