package relay

import (
	"context"
	"net/http"
	"time"

	"github.com/RabbyHub/derelay/config"
	"github.com/RabbyHub/derelay/log"
	"github.com/gorilla/mux"
)

type relayServer struct {
	httpServer *http.Server
	wsServer   *WsServer
}

func NewRelayServer(config *config.RelayConfig, wsServer *WsServer) *relayServer {
	r := mux.NewRouter()

	r.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	// handle websocket connection
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		wsServer.NewClientConn(w, r)
	})

	s := &http.Server{
		Addr:           config.Listen,
		Handler:        r,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return &relayServer{
		httpServer: s,
		wsServer:   wsServer,
	}
}

func (rs *relayServer) Run() {
	err := rs.httpServer.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// Shutdown Gracefully shutdown the relay server
func (rs *relayServer) Shutdown() {
	rs.httpServer.Shutdown(context.TODO())
	rs.wsServer.Shutdown()
}
