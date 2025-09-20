package ui

import (
	c "nned/internal/common"
	mon "nned/internal/monitor"

	tea "github.com/charmbracelet/bubbletea"
)

func Start(dep *c.Dependencies, ctx *c.Context) func() error {
	return func() error {
		monitor, _ := mon.NewMonitor(mon.Config{
			RefreshInterval: ctx.Config.RefreshInterval,
			Feeds:           ctx.Config.NewsFeeds,
			LastDate:        ctx.Config.LastDate,
		})

		p := tea.NewProgram(
			NewModel(*dep, *ctx, monitor),
			tea.WithMouseCellMotion(),
			// tea.WithAltScreen(),
		)

		var err error

		err = monitor.SetOnUpdate(mon.ConfigUpdateFunc{
			OnUpdateArticle: func(article c.Article, versionVector int) {
				p.Send(SetArticleMsg{
					article:       article,
					versionVector: versionVector,
				})
			},
		})
		if err != nil {
			return err
		}

		_, err = p.Run()

		return err
	}
}
