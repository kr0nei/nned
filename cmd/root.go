package cmd

import (
	"fmt"
	"os"
	"time"

	cli "nned/internal/cli"
	c "nned/internal/common"
	ui "nned/internal/ui"

	"github.com/spf13/cobra"
)

var (
	Version    = "v0.0.1"
	configPath string
	ctx        c.Context
	config     c.Config
	options    cli.Options
	dep        c.Dependencies
	err        error
	rootCmd    = &cobra.Command{
		Version: Version,
		Use:     "nned",
		Short:   "A simple RSS feed reader",
		PreRun:  initContext,
		Args:    cli.Validate(&config, &err),
		Run:     cli.Run(ui.Start(&dep, &ctx)),
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().StringVar(&configPath, "config", "", "config file (default is $HOME/.nned.yaml)")
	rootCmd.Flags().StringVarP(&options.Feeds, "feeds", "n", "", "comma separated list of rss news feeds")
	rootCmd.Flags().IntVarP(&options.RefreshInterval, "interval", "i", 30, "refresh interval in seconds")
	rootCmd.Flags().TimeVarP(&options.LastDate, "last-date", "l", time.Now().Add(-time.Hour*time.Duration(24*7)), []string{"2006-01-02", "2006-01-02 15:04:05"}, "oldest date to fetch news")
}

func initConfig() {
	dep = cli.GetDependencies()
	config, err = cli.GetConfig(dep, configPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initContext(_ *cobra.Command, _ []string) {
	ctx, err = cli.GetContext(dep, config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
