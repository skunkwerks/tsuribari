package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port string `mapstructure:"port"`
		Host string `mapstructure:"host"`
	} `mapstructure:"server"`

	CouchDB struct {
		URL      string `mapstructure:"url"`
		Database string `mapstructure:"database"`
	} `mapstructure:"couchdb"`

	RabbitMQ struct {
		URL      string `mapstructure:"url"`
		Exchange string `mapstructure:"exchange"`
		Queue    string `mapstructure:"queue"`
	} `mapstructure:"rabbitmq"`

	Security struct {
		TrustedIPs []string          `mapstructure:"trusted_ips"`
		Secrets    map[string]string `mapstructure:"secrets"`
	} `mapstructure:"security"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/usr/local/etc/tsuribari")

	// Set defaults
	viper.SetDefault("server.port", "4003")
	viper.SetDefault("server.host", "")
	viper.SetDefault("couchdb.url", "http://admin:passwd@127.0.0.1:5984")
	viper.SetDefault("couchdb.database", "koans")
	viper.SetDefault("rabbitmq.url", "amqp://guest:guest@localhost:5672/")
	viper.SetDefault("rabbitmq.exchange", "koans.topic")
	viper.SetDefault("rabbitmq.queue", "koans.workflow")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
