package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RabbyHub/derelay/config"
	"github.com/RabbyHub/derelay/metrics"
	"github.com/RabbyHub/derelay/relay"
	"github.com/gorilla/mux"
)

func startMetricServer(config *config.MetricConfig) {
	r := mux.NewRouter()

	r.Handle("/metrics", metrics.Handler())

	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)

	r.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	r.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	r.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
	r.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	r.Handle("/debug/pprof/block", pprof.Handler("block"))
	r.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))

	s := &http.Server{
		Addr:           config.Listen,
		Handler:        r,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		s.ListenAndServe()
	}()
}

func parseCmdlineAndLoadConfig() config.Config {
	cmdlineConfig := config.Config{}
	configFilePath := flag.String("config", "", "config file")

	// define cmdline options
	flag.StringVar(&cmdlineConfig.RelayServerConfig.Listen, "relay.addr", "", "relay server listen address")
	flag.StringVar(&cmdlineConfig.RedisServerConfig.ServerAddr, "redis.server_addr", "", "redis server address")

	flag.Parse()

	// load file config
	fileConfig := config.LoadConfig(*configFilePath)

	// overwrite with cmdline config
	if listen := cmdlineConfig.RelayServerConfig.Listen; listen != "" {
		fileConfig.RelayServerConfig.Listen = listen
	}

	if serverAddr := cmdlineConfig.RedisServerConfig.ServerAddr; serverAddr != "" {
		fileConfig.RedisServerConfig.ServerAddr = serverAddr
	}

	return fileConfig
}

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	config := parseCmdlineAndLoadConfig()

	wsServer := relay.NewWSServer(&config)
	relayServer := relay.NewRelayServer(&config.RelayServerConfig, wsServer)

	// start websocket server
	go func() {
		wsServer.Run()
	}()

	// Start relay server
	go func() {
		relayServer.Run()
	}()

	// Start metric and pprof server
	if config.MetricServerConfig.Enable {
		startMetricServer(&config.MetricServerConfig)
	}

	sig := <-sigChan
	waitSeconds := config.RelayServerConfig.GracefulShutdownWaitSeconds
	log.Printf("Sig %v received, shutting down, graceful shutdown wait: %v seconds\n", sig, waitSeconds)

	relayServer.Shutdown()

	<-time.After(time.Duration(waitSeconds) * time.Second)

}
