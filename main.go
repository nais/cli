package main

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"

	"github.com/nais/cli/cmd"
	"github.com/nais/cli/pkg/metrics"
)

func main() {
	res, err := metrics.NewResource()
	if err != nil {
		panic(err)
	}

	meterProvider, err := metrics.NewMeterProvider(res)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := meterProvider.Shutdown(context.Background()); err != nil {
			log.Println(err)
		}
	}()

	otel.SetMeterProvider(meterProvider)

	cmd.Run()

}
