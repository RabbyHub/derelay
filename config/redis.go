package config

type RedisConfig struct {
	ServerAddr string `yaml:"server_addr"`
	Password   string `yaml:"password"`
}
