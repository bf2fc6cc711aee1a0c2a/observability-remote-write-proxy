package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var Registry = prometheus.NewRegistry()

var (
	IncomingRequestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "observability_proxy",
		Subsystem: "remote_write",
		Name:      "incoming_request_count",
		Help:      "number of received remote write requests",
	}, []string{})

	OutgoingRequestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "observability_proxy",
		Subsystem: "remote_write",
		Name:      "outgoing_request_count",
		Help:      "number of forwarded remote write requests",
	}, []string{})
)

func init() {
	Registry.MustRegister(IncomingRequestCount)
	Registry.MustRegister(OutgoingRequestCount)
	Registry.MustRegister(collectors.NewGoCollector(
		collectors.WithGoCollectorRuntimeMetrics(),
	))
}
