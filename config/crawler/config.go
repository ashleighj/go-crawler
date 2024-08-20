package crawler

import (
	"os"
	"path/filepath"
	logger "webcrawler/logger"

	"gopkg.in/yaml.v3"
)

const configFile = "../../config/crawler/config.yml"

var config *Config

var defaultConfig = Config{
	Server: ServerConfig{
		Host: "localhost",
		Port: 8080,
		Timeout: TimeoutConfig{
			Server: 30,
			Write:  5,
			Read:   3,
			Idle:   5,
		},
	},
	Seeds:            []string{"https://www.wisdomforgoldfish.com"},
	DomainHitDelayMS: 2000,
	MaxDepth:         5,
	IgnoreIfContains: []string{".png", ".jpg", "javascript"},
}

// Config - configuration relating to the Crawler app
type Config struct {
	Server           ServerConfig `yaml:"server"`
	Seeds            []string     `yaml:"seeds"`
	BlacklistedURLs  []string     `yaml:"blacklisted_urls"`
	DomainHitDelayMS int          `yaml:"domain_delay_ms"`
	MaxDepth         int          `yaml:"max_depth"`
	IgnoreIfContains []string     `yaml:"ignore_if_contains"`
}

// ServerConfig - config relating to the web server
type ServerConfig struct {
	Host    string        `yaml:"host"`
	Port    int           `yaml:"port"`
	Timeout TimeoutConfig `yaml:"timeout"`
}

// TimeoutConfig - config relating to web server timeouts
type TimeoutConfig struct {
	// Server - general server timeout to use for graceful shutdowns
	Server int `yaml:"server"`

	// Write - time until an HTTP server write opperation is cancelled
	Write int `yaml:"write"`

	// Read - time until an HTTP server read operation is cancelled
	Read int `yaml:"read"`

	// Idle - time before an IDLE HTTP session is closed
	Idle int `yaml:"idle"`
}

// Get returns the config from file, or, if unavailable, default config
func Get() *Config {
	if config == nil {
		conf, e := NewFromFile()
		if e != nil {
			logger.Warnf("could not get config from file due to error [%s], using default instead", e)
			conf = NewDefault()
		}
		config = conf
	}
	return config
}

// NewFromFile creates and returns an instance of Config based on the contents of config.yml
func NewFromFile() (c *Config, e error) {
	configFilePath, e := filepath.Abs(configFile)

	yamlBytes, e := os.ReadFile(configFilePath)
	if e != nil {
		logger.Error(e)
		return
	}

	c = &defaultConfig
	if e = yaml.Unmarshal(yamlBytes, c); e != nil {
		logger.Error(e)
		return nil, e
	}

	e = c.validate()
	if e != nil {
		logger.Error(e)
		return nil, e
	}
	return
}

// NewDefault creates and returns the default instance of Config
func NewDefault() *Config {
	return &defaultConfig
}

func (c *Config) validate() (e error) {
	return nil
}
