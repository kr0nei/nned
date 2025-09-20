package common

import (
	"log"
	"time"

	"github.com/spf13/afero"
)

type Context struct {
	Config Config
	Logger *log.Logger
}

type Config struct {
	RefreshInterval int       `yaml:"interval"`
	NewsFeeds       []Feed    `yaml:"feeds"`
	Debug           bool      `yaml:"debug"`
	LastDate        time.Time `yaml:"last-date"`
}

type Dependencies struct {
	Fs afero.Fs
}

type Feed struct {
	Url   string `yaml:"url"`
	Title string `yaml:"title"`
	Color string `yaml:"color"`
}

type Article struct {
	Title       string
	Description string
	Link        string
	Date        *time.Time
	Source      string
	SourceTitle string
	SourceColor string
}

type MessageUpdate[T any] struct {
	Data          T
	ID            string
	Sequence      int64
	VersionVector int
}

type ByDate []*Article

func (a ByDate) Len() int           { return len(a) }
func (a ByDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDate) Less(i, j int) bool { return a[i].Date.After(*a[j].Date) }
