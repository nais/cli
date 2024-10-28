package main

import "github.com/nais/cli/cmd"
import "github.com/nais/cli/pkg/metrics"
import (
	"log"
)

func main() {
	_, err := metrics.InitMetrics()
	if err != nil {
		log.Fatalf("Error initializing metrics: %v", err)
	}
	metrics := metrics.GetMetrics()
	metrics.RecordError()
	metrics.PushMetrics(metrics.PushgatewayURL)

	cmd.Run()
}
