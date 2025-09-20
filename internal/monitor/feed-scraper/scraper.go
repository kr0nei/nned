package feedscraper

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	c "nned/internal/common"

	"github.com/mmcdole/gofeed"
)

type Scraper struct {
	ctx                context.Context
	cancel             context.CancelFunc
	numWorkers         int
	mu                 sync.RWMutex
	started            bool
	chanError          chan error
	chanUpdateArticle  chan c.MessageUpdate[c.Article]
	chanRequestArticle chan []c.Feed
	lastDate           time.Time
}

type Config struct {
	Ctx                context.Context
	ChanUpdateArticle  chan c.MessageUpdate[c.Article]
	ChanRequestArticle chan []c.Feed
	ChanError          chan error
	LastDate           time.Time
}

func NewScraper(config Config) *Scraper {
	ctx, cancel := context.WithCancel(config.Ctx)
	return &Scraper{
		ctx:                ctx,
		numWorkers:         2,
		cancel:             cancel,
		chanError:          config.ChanError,
		chanUpdateArticle:  config.ChanUpdateArticle,
		chanRequestArticle: config.ChanRequestArticle,
		lastDate:           config.LastDate,
	}
}

func (s *Scraper) Start() error {
	if s.started {
		return errors.New("scraper already started")
	}
	go s.handleScraping()

	s.started = true
	return nil
}

func (s *Scraper) Stop() error {
	if !s.started {
		return errors.New("scraper not started")
	}
	s.cancel()
	s.started = false
	return nil
}

func (s *Scraper) handleScraping() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case feeds := <-s.chanRequestArticle:
			if len(feeds) == 0 {
				continue
			}
			numFeeds := len(feeds)
			jobs := make(chan c.Feed, numFeeds)
			results := make(chan []c.Article, numFeeds)
			var wg sync.WaitGroup
			wg.Add(numFeeds)

			for w := 0; w < s.numWorkers; w++ {
				go s.getNewArticles(jobs, results, s.chanError)
			}
			for i := 0; i < numFeeds; i++ {
				jobs <- feeds[i]
			}
			close(jobs)
			articles := make([]c.Article, 0)
			for a := 1; a <= numFeeds; a++ {
				new_articles := <-results
				articles = append(articles, new_articles...)
			}
			for _, article := range articles {
				s.chanUpdateArticle <- c.MessageUpdate[c.Article]{
					Data: article,
				}
			}
		}
	}
}

func (s *Scraper) getNewArticles(jobs <-chan c.Feed, results chan<- []c.Article, errors chan<- error) {
	fp := gofeed.NewParser()

	var articles []c.Article
	articles = make([]c.Article, 0)
	for job := range jobs {
		feed, err := fp.ParseURL(job.Url)
		if err != nil {
			results <- nil
			errors <- fmt.Errorf("error parsing feed %s", job.Url)
			continue
		}
		for _, item := range feed.Items {
			if item.PublishedParsed == nil {
				continue
			}
			articles = append(articles, c.Article{
				Title:       item.Title,
				Description: item.Description,
				Link:        item.Link,
				Date:        item.PublishedParsed,
				Source:      feed.Title,
				SourceTitle: job.Title,
				SourceColor: job.Color,
			})
		}
		{
			var tmp []c.Article
			tmp = make([]c.Article, 0)
			for _, article := range articles {
				if !article.Date.After(time.Now()) && article.Date.After(s.lastDate) {
					tmp = append(tmp, article)
				}
			}
			articles = tmp
		}
		results <- articles
	}
}
