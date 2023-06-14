package config_test

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/RabbyHub/derelay/config"
	"gopkg.in/yaml.v3"
)

func TestConfigOverwritten(t *testing.T) {
	// create the test config file
	defaultConfig := config.LoadConfig("")

	expectedConfig := defaultConfig
	expectedConfig.RelayServerConfig.Listen = ":9099"
	expectedConfig.WsServerConfig.HeartbeatInterval = 20
	expectedConfig.WsServerConfig.AllowedOrigins = []string{"debank.com", "ethereum.com"}
	expectedConfig.RedisServerConfig.ServerAddr = ":654321"

	tmpfile, err := ioutil.TempFile("", "tmpconfig.yml")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpfile.Name())

	if err := yaml.NewEncoder(tmpfile).Encode(expectedConfig); err != nil {
		panic(err)
	}

	loadedConfig := config.LoadConfig(tmpfile.Name())
	if !reflect.DeepEqual(loadedConfig, expectedConfig) {
		t.Errorf("load config test failed, loaded config = %v, want %v", loadedConfig, expectedConfig)
	}
}

func TestLoadIncompleteConfig(t *testing.T) {
	raw := `
wsserver_config:
  allowed_origins:
    - "debank.com"
    - "ethereum.com"
redis_config:
  server_addr: ":654321"
`

	// create the test config file
	defaultConfig := config.LoadConfig("")
	expectedConfig := defaultConfig
	expectedConfig.WsServerConfig.AllowedOrigins = []string{"debank.com", "ethereum.com"}
	expectedConfig.RedisServerConfig.ServerAddr = ":654321"

	tmpfile, err := ioutil.TempFile("", "tmpconfig.yml")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(raw); err != nil {
		panic(err)
	}
	if err := tmpfile.Close(); err != nil {
		panic(err)
	}

	loadedConfig := config.LoadConfig(tmpfile.Name())
	if !reflect.DeepEqual(loadedConfig, expectedConfig) {
		t.Errorf("load config test failed, loaded config = %v, want %v", loadedConfig, expectedConfig)
	}
}
