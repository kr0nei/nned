package monitor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	c "nned/internal/common"
	feedscraper "nned/internal/monitor/feed-scraper"
)

type Config struct {
	RefreshInterval int
	Feeds           []c.Feed
	LastDate        time.Time
}

type Monitor struct {
	Config               Config
	cancel               context.CancelFunc
	ctx                  context.Context
	mu                   sync.RWMutex
	chanUpdateArticle    chan c.MessageUpdate[c.Article]
	chanRequestArticle   chan []c.Feed
	chanError            chan error
	onUpdateArticle      func(article c.Article, versionVector int)
	articleVersionVector int
	scraper              *feedscraper.Scraper
}

type ConfigUpdateFunc struct {
	OnUpdateArticle func(article c.Article, versionVector int)
}

func NewMonitor(config Config) (*Monitor, error) {
	ctx, cancel := context.WithCancel(context.Background())
	chanError := make(chan error, 5)
	chanUpdateArticle := make(chan c.MessageUpdate[c.Article], 2)
	chanRequestArticle := make(chan []c.Feed, 2)

	feedScraper := feedscraper.NewScraper(feedscraper.Config{
		Ctx:                ctx,
		ChanUpdateArticle:  chanUpdateArticle,
		ChanRequestArticle: chanRequestArticle,
		ChanError:          chanError,
		LastDate:           config.LastDate,
	})

	return &Monitor{
		Config:             config,
		cancel:             cancel,
		ctx:                ctx,
		chanUpdateArticle:  chanUpdateArticle,
		chanRequestArticle: chanRequestArticle,
		chanError:          chanError,
		scraper:            feedScraper,
	}, nil
}

func (m *Monitor) Start() {
	go m.handleUpdates()
	m.scraper.Start()
	ticker := time.NewTicker(time.Duration(m.Config.RefreshInterval) * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				m.chanRequestArticle <- m.Config.Feeds
			case <-m.ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
	m.chanRequestArticle <- m.Config.Feeds
}

func (m *Monitor) SetOnUpdate(config ConfigUpdateFunc) error {
	if config.OnUpdateArticle == nil {
		return errors.New("onUpdateArticle must be set ")
	}
	m.onUpdateArticle = config.OnUpdateArticle
	return nil
}

func (m *Monitor) handleUpdates() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case update := <-m.chanUpdateArticle:
			if update.VersionVector != m.articleVersionVector {
				continue
			}
			go m.onUpdateArticle(update.Data, update.VersionVector)
		case err := <-m.chanError:
			fmt.Printf("error %+v\n", err)
		}
	}
}

func (m *Monitor) Stop() {
	m.cancel()
}
