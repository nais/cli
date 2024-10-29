package cmd

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"log"
	"os"
	"time"

	"github.com/nais/cli/cmd/aivencmd"
	"github.com/nais/cli/cmd/appstartercmd"
	"github.com/nais/cli/cmd/devicecmd"
	"github.com/nais/cli/cmd/kubeconfigcmd"
	"github.com/nais/cli/cmd/postgrescmd"
	"github.com/nais/cli/cmd/rootcmd"
	"github.com/nais/cli/cmd/validatecmd"
	"github.com/urfave/cli/v2"

	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var (
	// Is set during build
	version = "local"
	commit  = "uncommited"
)

func commands() []*cli.Command {
	return append(
		rootcmd.Commands(),
		aivencmd.Command(),
		appstartercmd.Command(),
		devicecmd.Command(),
		kubeconfigcmd.Command(),
		postgrescmd.Command(),
		validatecmd.Command(),
	)
}

func Run() {
	app := &cli.App{
		Name:                 "nais",
		Usage:                "A NAIS CLI",
		Description:          "A simple CLI application that developers in NAV can use",
		Version:              version + "-" + commit,
		EnableBashCompletion: true,
		HideHelpCommand:      true,
		Suggest:              true,
		Commands:             commands(),
	}

	var validSubcommands []string
	for _, command := range app.Commands {
		validSubcommands = append(validSubcommands, command.Name)
		for _, subcommand := range command.Subcommands {
			validSubcommands = append(validSubcommands, subcommand.Name)
		}
	}

	meterProv := New()
	defer meterProv.Shutdown(context.Background())
	commandHistogram, _ := meterProv.Meter("nais-cli").Int64Histogram("flag_usage", metric.WithDescription("Usage frequency of command flags"))
	ctx := context.Background()
	recordCommandUsage(ctx, commandHistogram, intersection(os.Args, validSubcommands))
	meterProv.ForceFlush(context.Background())

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func NewResource() (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName("nais_cli"),
			semconv.ServiceVersion(version+":"+commit),
		))
}

func NewMeterProvider(res *resource.Resource) *sdkmetric.MeterProvider {
	dnt := os.Getenv("DO_NOT_TRACK")
	var url string
	if dnt == "1" {
		url = "http://localhost"
	} else {
		url = "https://collector-internet.nav.cloud.nais.io"
	}
	metricExporter, _ := otlpmetrichttp.New(
		context.Background(),
		otlpmetrichttp.WithEndpointURL(url),
	)
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter,
			sdkmetric.WithInterval(1*time.Second))),
	)
	return meterProvider
}

func New() *sdkmetric.MeterProvider {
	res, _ := NewResource()
	meterProvider := NewMeterProvider(res)
	return meterProvider
}

func recordCommandUsage(ctx context.Context, histogram metric.Int64Histogram, flags []string) {
	for _, flag := range flags {
		histogram.Record(ctx, 1, metric.WithAttributes(attribute.String("flag", flag)))
	}
}

func intersection(list1, list2 []string) []string {
	elements := make(map[string]bool)
	for _, item := range list1 {
		elements[item] = true
	}
	var result []string
	for _, item := range list2 {
		if elements[item] {
			result = append(result, item)
		}
	}
	return result
}
