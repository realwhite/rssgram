// Тест создан с помощью AI
package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_Initialization(t *testing.T) {
	assert.NotNil(t, FeedsCount)
	assert.NotNil(t, FeedGetError)
	assert.NotNil(t, FeedGetSuccess)
	assert.NotNil(t, FeedGetTimeSec)
	assert.NotNil(t, NewItemsCount)
	assert.NotNil(t, ItemsEnrichTimeSec)
	assert.NotNil(t, ItemsReadyToSendCount)
	assert.NotNil(t, ItemsSentFailedCount)
	assert.NotNil(t, ItemsSentErrorCount)
	assert.NotNil(t, ItemsSentSuccessCount)
}

func TestMetrics_FeedsCount(t *testing.T) {
	FeedsCount.Set(0)
	FeedsCount.Set(5)
	value := testutil.ToFloat64(FeedsCount)
	assert.Equal(t, 5.0, value)
}

func TestMetrics_FeedGetError(t *testing.T) {
	FeedGetError.Reset()
	FeedGetError.WithLabelValues("test-feed").Inc()
	FeedGetError.WithLabelValues("test-feed").Inc()
	value := testutil.ToFloat64(FeedGetError.WithLabelValues("test-feed"))
	assert.Equal(t, 2.0, value)
}

func TestMetrics_FeedGetSuccess(t *testing.T) {
	FeedGetSuccess.Reset()
	FeedGetSuccess.WithLabelValues("test-feed").Inc()
	value := testutil.ToFloat64(FeedGetSuccess.WithLabelValues("test-feed"))
	assert.Equal(t, 1.0, value)
}

func TestMetrics_NewItemsCount(t *testing.T) {
	NewItemsCount.Reset()
	NewItemsCount.WithLabelValues("test-feed").Add(10)
	NewItemsCount.WithLabelValues("test-feed").Add(5)
	value := testutil.ToFloat64(NewItemsCount.WithLabelValues("test-feed"))
	assert.Equal(t, 15.0, value)
}

func TestMetrics_ItemsReadyToSendCount(t *testing.T) {
	ItemsReadyToSendCount.Set(0)
	ItemsReadyToSendCount.Set(25)
	value := testutil.ToFloat64(ItemsReadyToSendCount)
	assert.Equal(t, 25.0, value)
}

func TestMetrics_ItemsSentFailedCount(t *testing.T) {
	ItemsSentFailedCount.Set(0)
	ItemsSentFailedCount.Set(3)
	value := testutil.ToFloat64(ItemsSentFailedCount)
	assert.Equal(t, 3.0, value)
}

func TestMetrics_ItemsSentErrorCount(t *testing.T) {
	ItemsSentErrorCount.Reset()
	ItemsSentErrorCount.WithLabelValues("test-feed").Inc()
	ItemsSentErrorCount.WithLabelValues("test-feed").Inc()
	ItemsSentErrorCount.WithLabelValues("another-feed").Inc()
	value1 := testutil.ToFloat64(ItemsSentErrorCount.WithLabelValues("test-feed"))
	value2 := testutil.ToFloat64(ItemsSentErrorCount.WithLabelValues("another-feed"))
	assert.Equal(t, 2.0, value1)
	assert.Equal(t, 1.0, value2)
}

func TestMetrics_ItemsSentSuccessCount(t *testing.T) {
	ItemsSentSuccessCount.Reset()
	ItemsSentSuccessCount.WithLabelValues("test-feed").Inc()
	ItemsSentSuccessCount.WithLabelValues("test-feed").Inc()
	ItemsSentSuccessCount.WithLabelValues("test-feed").Inc()
	value := testutil.ToFloat64(ItemsSentSuccessCount.WithLabelValues("test-feed"))
	assert.Equal(t, 3.0, value)
}

func TestMetrics_Registry(t *testing.T) {
	registry := prometheus.NewRegistry()
	registry.MustRegister(FeedsCount)
	registry.MustRegister(FeedGetError)
	registry.MustRegister(FeedGetSuccess)
	registry.MustRegister(FeedGetTimeSec)
	registry.MustRegister(NewItemsCount)
	registry.MustRegister(ItemsEnrichTimeSec)
	registry.MustRegister(ItemsReadyToSendCount)
	registry.MustRegister(ItemsSentFailedCount)
	registry.MustRegister(ItemsSentErrorCount)
	registry.MustRegister(ItemsSentSuccessCount)
	metrics, err := registry.Gather()
	assert.NoError(t, err)
	assert.Greater(t, len(metrics), 0)
}
