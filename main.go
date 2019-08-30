package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rifki192/nsqd-exporter/stats"

	logger "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	// Version is defined at build time - see VERSION file
	Version string

	knownTopics     []string
	knownChannels   []string
	buildInfoMetric *prometheus.GaugeVec
	nsqMetrics      = make(map[string]*prometheus.GaugeVec)
)

func main() {
	app := cli.NewApp()
	app.Version = Version
	app.Name = "nsqd-exporter"
	app.Usage = "Scrapes multiple nsqd stats and serves them up as Prometheus metrics"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "listenPort, lp",
			Value:  "12500",
			Usage:  "Port on which prometheus will expose metrics",
			EnvVar: "LISTEN_PORT",
		},
	}
	app.Action = func(c *cli.Context) {
		// Start HTTP server
		http.Handle("/metrics", promhttp.Handler())
		http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
			statsHandler(w, r)
		})

		err := http.ListenAndServe(":"+strconv.Itoa(c.GlobalInt("listenPort")), nil)
		if err != nil {
			logger.Fatal("Error starting Prometheus metrics server: " + err.Error())
		}
	}

	app.Run(os.Args)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	timeoutSeconds, err := getTimeout(r, 0.5)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse timeout from Prometheus header: %s", err), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeoutSeconds*float64(time.Second)))
	defer cancel()
	r = r.WithContext(ctx)

	probeSuccessGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "scrape_success",
		Help: "Displays whether or not the scrape to nsqd-stats was a success",
	})
	probeDurationGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "scrape_duration_seconds",
		Help: "Returns how long the scrape to nsqd-stats took to complete in seconds",
	})
	params := r.URL.Query()
	instance := params.Get("target")
	if instance == "" {
		http.Error(w, "Instance target parameter is missing", http.StatusBadRequest)
		return
	}

	logger.Debug("Getting nsqd stats from ", instance)
	start := time.Now()
	registry := prometheus.NewRegistry()
	registry.MustRegister(probeSuccessGauge)
	registry.MustRegister(probeDurationGauge)
	stats := stats.New(registry, instance)
	duration := time.Since(start).Seconds()
	probeDurationGauge.Set(duration)
	if stats == true {
		probeSuccessGauge.Set(1)
		logger.Debug("Stats Collected.")
	} else {
		logger.Error("Failed to get stats from ", instance)
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func getTimeout(r *http.Request, offset float64) (timeoutSeconds float64, err error) {
	// If a timeout is configured via the Prometheus header, add it to the request.
	if v := r.Header.Get("X-Prometheus-Scrape-Timeout-Seconds"); v != "" {
		var err error
		timeoutSeconds, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, err
		}
	}
	if timeoutSeconds == 0 {
		timeoutSeconds = 120
	}

	var maxTimeoutSeconds = timeoutSeconds - offset
	// if module.Timeout.Seconds() < maxTimeoutSeconds && module.Timeout.Seconds() > 0 {
	// 	timeoutSeconds = module.Timeout.Seconds()
	// } else {
	timeoutSeconds = maxTimeoutSeconds
	// }

	return timeoutSeconds, nil
}
