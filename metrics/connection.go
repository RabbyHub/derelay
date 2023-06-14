package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	countNewConnections = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "new_connections",
		Help:      "Number of new connections",
	})
	countClosedConnections = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "closed_connections",
		Help:      "Number of closed connections",
	})
	gaugeCurrentConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "current_connections",
		Help:      "Number of current connections",
	})
)

func IncNewConnection() {
	countNewConnections.Inc()
}

func IncClosedConnection() {
	countClosedConnections.Inc()
}

func SetCurrentConnections(num int) {
	gaugeCurrentConnections.Set(float64(num))
}

func init() {
	prometheus.MustRegister(countNewConnections)
	prometheus.MustRegister(countClosedConnections)
	prometheus.MustRegister(gaugeCurrentConnections)
}
