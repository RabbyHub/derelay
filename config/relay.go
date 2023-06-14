package config

type RelayConfig struct {
	Version string `yaml:"version"`
	Mode    string `yaml:"mode"`

	Listen                      string `yaml:"listen"`
	GracefulShutdownWaitSeconds int    `yaml:"graceful_shutdown_wait_seconds"`
}
