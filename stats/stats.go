package stats

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	PrometheusNamespace = "nsqd"
	DepthMetric         = "depth"
	BackendDepthMetric  = "backend_depth"
	InFlightMetric      = "in_flight_count"
	TimeoutCountMetric  = "timeout_count_total"
	RequeueCountMetric  = "requeue_count_total"
	DeferredCountMetric = "deferred_count_total"
	MessageCountMetric  = "message_count_total"
	ClientCountMetric   = "client_count"
	ChannelCountMetric  = "channel_count"
	InfoMetric          = "info"
)

func New(registry *prometheus.Registry, nsqdUrl string) bool {
	initMetrics(registry)
	success := fetchAndSetStats(nsqdUrl)
	return success
}

func initMetrics(reg *prometheus.Registry) {
	// Initialize Prometheus metrics
	var emptyMap map[string]string
	commonLabels := []string{"type", "topic", "paused", "channel"}
	buildInfoMetric = createGaugeVector(reg, "nsqd_prometheus_exporter_build_info", "", "",
		"nsqd-prometheus-exporter build info", emptyMap, []string{"version"})
	// buildInfoMetric.WithLabelValues(Version).Set(1)
	// # HELP nsqd_info nsqd info
	// # TYPE nsqd_info gauge
	nsqMetrics[InfoMetric] = createGaugeVector(reg, InfoMetric, PrometheusNamespace,
		"", "nsqd info", emptyMap, []string{"health", "start_time", "version"})
	// # HELP nsqd_depth Queue depth
	// # TYPE nsqd_depth gauge
	nsqMetrics[DepthMetric] = createGaugeVector(reg, DepthMetric, PrometheusNamespace,
		"", "Queue depth", emptyMap, commonLabels)
	// # HELP nsqd_backend_depth Queue backend depth
	// # TYPE nsqd_backend_depth gauge
	nsqMetrics[BackendDepthMetric] = createGaugeVector(reg, BackendDepthMetric, PrometheusNamespace,
		"", "Queue backend depth", emptyMap, commonLabels)
	// # HELP nsqd_in_flight_count In flight count
	// # TYPE nsqd_in_flight_count gauge
	nsqMetrics[InFlightMetric] = createGaugeVector(reg, InFlightMetric, PrometheusNamespace,
		"", "In flight count", emptyMap, commonLabels)
	// # HELP nsqd_timeout_count_total Timeout count
	// # TYPE nsqd_timeout_count_total gauge
	nsqMetrics[TimeoutCountMetric] = createGaugeVector(reg, TimeoutCountMetric, PrometheusNamespace,
		"", "Timeout count", emptyMap, commonLabels)
	// # HELP nsqd_requeue_count_total Requeue count
	// # TYPE nsqd_requeue_count_total gauge
	nsqMetrics[RequeueCountMetric] = createGaugeVector(reg, RequeueCountMetric, PrometheusNamespace,
		"", "Requeue count", emptyMap, commonLabels)
	// # HELP nsqd_deferred_count_total Deferred count
	// # TYPE nsqd_deferred_count_total gauge
	nsqMetrics[DeferredCountMetric] = createGaugeVector(reg, DeferredCountMetric, PrometheusNamespace,
		"", "Deferred count", emptyMap, commonLabels)
	// # HELP nsqd_message_count_total Total message count
	// # TYPE nsqd_message_count_total gauge
	nsqMetrics[MessageCountMetric] = createGaugeVector(reg, MessageCountMetric, PrometheusNamespace,
		"", "Total message count", emptyMap, commonLabels)
	// # HELP nsqd_client_count Number of clients
	// # TYPE nsqd_client_count gauge
	nsqMetrics[ClientCountMetric] = createGaugeVector(reg, ClientCountMetric, PrometheusNamespace,
		"", "Number of clients", emptyMap, commonLabels)
	// # HELP nsqd_channel_count Number of channels
	// # TYPE nsqd_channel_count gauge
	nsqMetrics[ChannelCountMetric] = createGaugeVector(reg, ChannelCountMetric, PrometheusNamespace,
		"", "Number of channels", emptyMap, commonLabels[:3])
}

// createGaugeVector creates a GaugeVec and registers it with Prometheus.
func createGaugeVector(registry *prometheus.Registry, name string, namespace string, subsystem string, help string,
	labels map[string]string, labelNames []string) *prometheus.GaugeVec {
	gaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        name,
		Help:        help,
		Namespace:   namespace,
		Subsystem:   subsystem,
		ConstLabels: prometheus.Labels(labels),
	}, labelNames)
	registry.MustRegister(gaugeVec)
	return gaugeVec
}
