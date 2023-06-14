package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	promNamespace = "wc"
	promSubsystem = "relay"
)

var (
	defaultHandler http.Handler
)

func Handler() http.Handler {
	return defaultHandler
}

func init() {
	defaultHandler = promhttp.Handler()
}
