package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	RelayServerConfig  RelayConfig  `yaml:"relay_config"`
	WsServerConfig     WsConfig     `yaml:"wsserver_config"`
	RedisServerConfig  RedisConfig  `yaml:"redis_config"`
	MetricServerConfig MetricConfig `yaml:"metric_config"`
}

type MetricConfig struct {
	Enable bool   `yaml:"enable"`
	Listen string `yaml:"listen"`
}

var defaultConfig = Config{
	RelayServerConfig: RelayConfig{
		Version:                     "1.0",
		Mode:                        "legacy",
		Listen:                      ":8080",
		GracefulShutdownWaitSeconds: 5,
	},
	WsServerConfig: WsConfig{
		HeartbeatInterval:          10,
		CheckSessionExpireInterval: 10,
		PendingSessionCacheTime:    1800, // in seconds
		MessageCacheTime:           1800,
		AllowedOrigins:             []string{"*"},
	},
	RedisServerConfig: RedisConfig{
		ServerAddr: "127.0.0.1:6379",
	},
	MetricServerConfig: MetricConfig{
		Enable: true,
		Listen: ":6060",
	},
}

func LoadConfig(configPath string) Config {
	if configPath == "" {
		// return a copy
		return defaultConfig
	}

	configFile, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("open config file error: %v\n", err)
	}
	defer configFile.Close()

	var config Config = defaultConfig
	parser := yaml.NewDecoder(configFile)
	err = parser.Decode(&config)
	if err != nil {
		log.Fatalf("parse config file error: %v\n", err)
	}
	return config
}
