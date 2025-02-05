package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const metricNamespace = "rssgram"

var FeedsCount = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: metricNamespace,
	Name:      "feeds_count",
})

var NewItemsCount = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: metricNamespace,
		Name:      "new_items_count",
	},
	[]string{"feed_name"},
)

var ItemsReadyToSendCount = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: metricNamespace,
	Name:      "items_ready_to_send_count",
})

var ItemsSentFailedCount = promauto.NewGauge(prometheus.GaugeOpts{
	Namespace: metricNamespace,
	Name:      "items_sent_failed_count",
})

var ItemsSentSuccessCount = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: metricNamespace,
		Name:      "items_sent_success_count",
	},
	[]string{"feed_name"},
)

var ItemsSentErrorCount = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: metricNamespace,
		Name:      "items_sent_error_count",
	},
	[]string{"feed_name"},
)

var FeedGetTimeSec = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: metricNamespace,
		Name:      "feed_get_time_sec",
	},
	[]string{"feed_name"},
)

var ItemsEnrichTimeSec = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: metricNamespace,
		Name:      "items_enrich_time_sec",
	},
	[]string{"feed_name"},
)

var FeedGetSuccess = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: metricNamespace,
		Name:      "feed_get_success",
	},
	[]string{"feed_name"},
)

var FeedGetError = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: metricNamespace,
		Name:      "feed_get_error",
	},
	[]string{"feed_name"},
)
