package internal

import (
	"os"

	"rssgram/internal/outputs/telegram"

	"gopkg.in/yaml.v3"
)

type FeedConfig struct {
	Name            string   `yaml:"name"`
	URL             string   `yaml:"url"`
	Type            string   `yaml:"type"`
	Interval        string   `yaml:"interval"`
	Key             string   `yaml:"key"`
	DescriptionType string   `yaml:"description_type"`
	Tags            []string `yaml:"tags"`
}

type MetricsConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

type Config struct {
	Feeds      []FeedConfig                         `yaml:"feeds"`
	Telegram   telegram.TelegramChannelOutputConfig `yaml:"telegram"`
	EnableTags bool                                 `yaml:"enable_tags"`
	Metrics    MetricsConfig                        `yaml:"metrics"`
}

func ParseConfig() (*Config, error) {
	var cnf Config

	cnfBytes, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(cnfBytes, &cnf)
	if err != nil {
		return nil, err
	}

	return &cnf, nil
}
