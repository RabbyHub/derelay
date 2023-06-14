package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// total messages = direct forwarded messages + cached messages
	countTotalMessages = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "total_messages",
		Help:      "Number of total messages",
	})
	countCachedMessages = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "new_cached_messages",
		Help:      "Number of new cached messages",
	})
	// uncached messages are messages received by client laterly
	countUncachedMessages = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "new_uncached_messages",
		Help:      "Number of cached messages consumed",
	})

	countMessages = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "messages",
		Help:      "Number of messages",
	}, []string{"phase"})

	countSessions = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "sessions",
		Help:      "Number of new pending sessions",
	}, []string{"phase"})

	countNewRequestedSessions = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "new_sessions",
		Help:      "Number of new pending sessions",
	})
	countReceivedSessions = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "received_sessions",
		Help:      "Number of new received sessions",
	})
	countEstablishedSessions = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "established_sessions",
		Help:      "Number of new established sessions",
	})
	countExpiredSessions = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: promNamespace,
		Subsystem: promSubsystem,
		Name:      "expired_sessions",
		Help:      "Number of expired sessions",
	})
)

func IncTotalMessages() {
	countTotalMessages.Inc()
	countMessages.With(prometheus.Labels{"phase": "total"}).Inc()
}
func IncCachedMessages() {
	countCachedMessages.Inc()
	countMessages.With(prometheus.Labels{"phase": "pending"}).Inc()
}
func DecCachedMessages() {
	countUncachedMessages.Inc()
	countMessages.With(prometheus.Labels{"phase": "delay_delivered"}).Inc()
}

func IncNewRequestedSessions() {
	countNewRequestedSessions.Inc()
	countSessions.With(prometheus.Labels{"phase": "new"}).Inc()
}

func IncReceivedSessions() {
	countNewRequestedSessions.Inc()
	countSessions.With(prometheus.Labels{"phase": "received"}).Inc()
}

func IncEstablishedSessions() {
	countEstablishedSessions.Inc()
	countSessions.With(prometheus.Labels{"phase": "established"}).Inc()
}

func IncExpiredSessions() {
	countExpiredSessions.Inc()
	countSessions.With(prometheus.Labels{"phase": "expired"}).Inc()
}

func init() {
	prometheus.MustRegister(countTotalMessages)
	prometheus.MustRegister(countCachedMessages)
	prometheus.MustRegister(countUncachedMessages)

	prometheus.MustRegister(countMessages)
	prometheus.MustRegister(countSessions)
	prometheus.MustRegister(countNewRequestedSessions)
	prometheus.MustRegister(countEstablishedSessions)
	prometheus.MustRegister(countReceivedSessions)
	prometheus.MustRegister(countExpiredSessions)
}
