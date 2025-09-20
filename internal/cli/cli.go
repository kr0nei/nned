package cli

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	c "nned/internal/common"

	"github.com/adrg/xdg"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Options struct {
	RefreshInterval int
	Feeds           string
	LastDate        time.Time
}

func Run(uiStartFn func() error) func(*cobra.Command, []string) {
	return func(_ *cobra.Command, _ []string) {
		err := uiStartFn()
		if err != nil {
			fmt.Println(fmt.Errorf("unable to start UI: %w", err).Error())
		}
	}
}

func Validate(config *c.Config, prevErr *error) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		if prevErr != nil && *prevErr != nil {
			return *prevErr
		}
		if len(config.NewsFeeds) == 0 {
			return errors.New("invalid config: no news feeds")
		}
		return nil
	}
}

func GetDependencies() c.Dependencies {
	return c.Dependencies{
		Fs: afero.NewOsFs(),
	}
}

func GetContext(d c.Dependencies, config c.Config) (c.Context, error) {
	var err error
	if err != nil {
		return c.Context{}, err
	}

	var logger *log.Logger
	if config.Debug {
		logger, err = getLogger(d)
		if err != nil {
			return c.Context{}, err
		}
	}

	context := c.Context{
		Config: config,
		Logger: logger,
	}
	return context, err
}

func readConfig(fs afero.Fs, configPathOption string) (c.Config, error) {
	var config c.Config
	configPath, err := getConfigPath(fs, configPathOption)
	if err != nil {
		return config, nil
	}

	handle, err := fs.Open(configPath)
	defer handle.Close()
	if err != nil {
		return config, fmt.Errorf("invalid config: %w", err)
	}

	err = yaml.NewDecoder(handle).Decode(&config)
	if err != nil {
		return config, fmt.Errorf("invalid config: %w", err)
	}
	return config, nil
}

func GetConfig(d c.Dependencies, configPath string) (c.Config, error) {
	config, err := readConfig(d.Fs, configPath)
	if err != nil {
		return c.Config{}, err
	}

	config.RefreshInterval = getRefreshInterval(0, config.RefreshInterval)
	config.Debug = false
	return config, nil
}

func getConfigPath(fs afero.Fs, configPathOption string) (string, error) {
	var err error

	if configPathOption != "" {
		return configPathOption, nil
	}
	home, _ := homedir.Dir()
	v := viper.New()
	v.SetFs(fs)
	v.SetConfigType("yaml")
	v.AddConfigPath(home)
	v.AddConfigPath(xdg.ConfigHome)
	v.AddConfigPath(xdg.ConfigHome + "/nned")
	v.SetConfigName(".nned")
	err = v.ReadInConfig()
	if err != nil {
		return "", fmt.Errorf("invalid config: %w", err)
	}
	return v.ConfigFileUsed(), nil
}

func getRefreshInterval(optionsRefreshInverval int, configRefreshInterval int) int {
	if optionsRefreshInverval > 0 {
		return optionsRefreshInverval
	}
	if configRefreshInterval > 0 {
		return configRefreshInterval
	}
	return 30
}

func getLogger(d c.Dependencies) (*log.Logger, error) {
	currentTime := time.Now()
	logFileName := fmt.Sprintf("nned-log-%s.log", currentTime.Format("2006-01-02"))
	logFile, err := d.Fs.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}
	return log.New(logFile, "", log.LstdFlags), nil
}
