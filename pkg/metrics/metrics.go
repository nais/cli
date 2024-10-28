package metrics

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	once     sync.Once
	instance *Metrics
)

func GetMetrics() *Metrics {
	return instance
}

type Metrics struct {
	PushgatewayURL    string
	errorCounter      prometheus.Counter
	subCommandCounter prometheus.CounterVec
}

func NewMetrics(pushgatewayURL string) *Metrics {
	errorCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   "nais_cli",
		Subsystem:   "subsystem_carl",
		Name:        "error_total",
		Help:        "Total number of errors encountered.",
		ConstLabels: map[string]string{},
	})

	subCommandCounter := // TODO: This should put the subcommand in the Subsystem field
		prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "nais_cli",
				Subsystem: "aiven",
				Name:      "subcommand_usage_total",
				Help:      "Total number of usages of the subcommand in the label",
			}, []string{"subcommand"}, // [subcommand, aiven, create]
			// [subcommand, gcp, auth]
		)

	return &Metrics{
		PushgatewayURL:    pushgatewayURL,
		errorCounter:      errorCounter,
		subCommandCounter: *subCommandCounter,
	}
}

func (m *Metrics) RecordError() {
	m.errorCounter.Inc()
}

func (m *Metrics) RecordSubcommandUsage(labels ...string) {
	for _, label := range labels {
		m.subCommandCounter.WithLabelValues(label).Inc()
	}
}

func (m *Metrics) PushMetrics(name string) error {
	registry := prometheus.NewRegistry()

	registry.MustRegister(m.errorCounter)

	if err := push.New(m.PushgatewayURL, name).
		Collector(m.errorCounter).
		Push(); err != nil {
		return fmt.Errorf("could not push metrics: %v", err)
	}

	return nil
}

func InitMetrics() (*Metrics, error) {
	var err error
	once.Do(func() {
		instance = NewMetrics("https://prometheus-pushgateway.prod-gcp.nav.cloud.nais.io")
		err = prometheus.Register(instance.errorCounter)
		if err != nil {
			instance = nil
		}
	})

	if err != nil {
		return nil, fmt.Errorf("failed to initialize metrics: %v", err)
	}

	return instance, nil
}
