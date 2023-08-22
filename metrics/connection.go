package metrics

import (
	"strconv"
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

	countSendBlocking = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "send_blockings",
		Help:      "Number of send blocking connections",
	}, []string{"sendbuflen"})
)

func IncNewConnection() {
	countNewConnections.Inc()
}

func IncClosedConnection() {
	countClosedConnections.Inc()
}

func IncSendBlocking(sendbufLen int) {
	lenstr := strconv.Itoa(sendbufLen)
	countSendBlocking.With(prometheus.Labels{"sendbuflen": lenstr}).Inc()
}

func SetCurrentConnections(num int) {
	gaugeCurrentConnections.Set(float64(num))
}

func init() {
	prometheus.MustRegister(countNewConnections)
	prometheus.MustRegister(countClosedConnections)
	prometheus.MustRegister(gaugeCurrentConnections)
	prometheus.MustRegister(countSendBlocking)
}
