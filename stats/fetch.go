package stats

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	logger "github.com/sirupsen/logrus"
)

var (
	knownTopics     []string
	knownChannels   []string
	buildInfoMetric *prometheus.GaugeVec
	nsqMetrics      = make(map[string]*prometheus.GaugeVec)
)

// fetchAndCacheStats scrapes stats from nsqd and updates all the Prometheus metrics
// above. If a dead topic or channel is detected, the metrics reseted
func fetchAndSetStats(nsqdURL string) bool {
	// Fetch stats
	stats, err := getNsqdStats(nsqdURL)
	if err != nil {
		logger.Fatal("Error scraping stats from nsqd: " + err.Error())
		return false
	}

	// Build list of detected topics and channels - the list of channels is built
	// including the topic name that each belongs to, as it is possible to have
	// multiple channels with the same name on different topics.
	var detectedChannels []string
	var detectedTopics []string
	for _, topic := range stats.Topics {
		detectedTopics = append(detectedTopics, topic.Name)
		for _, channel := range topic.Channels {
			detectedChannels = append(detectedChannels, topic.Name+channel.Name)
		}
	}

	// Exit if a dead topic or channel is detected
	if deadTopicOrChannelExists(knownTopics, detectedTopics) {
		logger.Warning("At least one old topic no longer included in nsqd stats - rebuilding metrics")
		for _, metric := range nsqMetrics {
			metric.Reset()
		}
	}
	if deadTopicOrChannelExists(knownChannels, detectedChannels) {
		logger.Warning("At least one old channel no longer included in nsqd stats - rebuilding metrics")
		for _, metric := range nsqMetrics {
			metric.Reset()
		}
	}

	// Update list of known topics and channels
	knownTopics = detectedTopics
	knownChannels = detectedChannels

	// Update info metric with health, start time, and nsqd version
	nsqMetrics[InfoMetric].
		WithLabelValues(stats.Health, fmt.Sprintf("%d", stats.StartTime), stats.Version).Set(1)

	// Loop through topics and set metrics
	for _, topic := range stats.Topics {
		paused := "false"
		if topic.Paused {
			paused = "true"
		}
		nsqMetrics[DepthMetric].WithLabelValues("topic", topic.Name, paused, "").
			Set(float64(topic.Depth))
		nsqMetrics[BackendDepthMetric].WithLabelValues("topic", topic.Name, paused, "").
			Set(float64(topic.BackendDepth))
		nsqMetrics[ChannelCountMetric].WithLabelValues("topic", topic.Name, paused).
			Set(float64(len(topic.Channels)))

		// Loop through a topic's channels and set metrics
		for _, channel := range topic.Channels {
			paused = "false"
			if channel.Paused {
				paused = "true"
			}
			nsqMetrics[DepthMetric].WithLabelValues("channel", topic.Name, paused, channel.Name).
				Set(float64(channel.Depth))
			nsqMetrics[BackendDepthMetric].WithLabelValues("channel", topic.Name, paused, channel.Name).
				Set(float64(channel.BackendDepth))
			nsqMetrics[InFlightMetric].WithLabelValues("channel", topic.Name, paused, channel.Name).
				Set(float64(channel.InFlightCount))
			nsqMetrics[TimeoutCountMetric].WithLabelValues("channel", topic.Name, paused, channel.Name).
				Set(float64(channel.TimeoutCount))
			nsqMetrics[RequeueCountMetric].WithLabelValues("channel", topic.Name, paused, channel.Name).
				Set(float64(channel.RequeueCount))
			nsqMetrics[DeferredCountMetric].WithLabelValues("channel", topic.Name, paused, channel.Name).
				Set(float64(channel.DeferredCount))
			nsqMetrics[MessageCountMetric].WithLabelValues("channel", topic.Name, paused, channel.Name).
				Set(float64(channel.MessageCount))
			nsqMetrics[ClientCountMetric].WithLabelValues("channel", topic.Name, paused, channel.Name).
				Set(float64(len(channel.Clients)))
		}
	}

	return true
}

// deadTopicOrChannelExists loops through a list of known topic or channel names
// and compares them to a list of detected names. If a known name no longer exists,
// it is deemed dead, and the function returns true.
func deadTopicOrChannelExists(known []string, detected []string) bool {
	// Loop through all known names and check against detected names
	for _, knownName := range known {
		found := false
		for _, detectedName := range detected {
			if knownName == detectedName {
				found = true
				break
			}
		}
		// If a topic/channel isn't found, it's dead
		if !found {
			return true
		}
	}
	return false
}
