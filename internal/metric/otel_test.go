package metric

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	collectorpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/protobuf/proto"
)

func TestBackgroundExportPreservesMetrics(t *testing.T) {
	t.Setenv("DO_NOT_TRACK", "0")
	path := filepath.Join(t.TempDir(), "payload")
	t.Setenv("NAIS_TEST_UPLOAD_FILE", path)
	provider := otel.GetMeterProvider()
	t.Cleanup(func() { otel.SetMeterProvider(provider) })
	flush := Initialize()
	CreateAndIncreaseCounter(t.Context(), "verify_audit_logging_error")
	histogram, err := otel.Meter(CliName).Int64Histogram(CliName+"_command_usage", metric.WithUnit("1"))
	if err != nil {
		t.Fatal(err)
	}
	histogram.Record(t.Context(), 1, metric.WithAttributes(attribute.String("command", "nais app list"), attribute.Bool("success", true)))
	flush(false)
	waitForFile(t, path+".done")
	payload := waitForFile(t, path)
	var request collectorpb.ExportMetricsServiceRequest
	if err := proto.Unmarshal(payload, &request); err != nil {
		t.Fatal(err)
	}
	if len(request.ResourceMetrics) != 1 {
		t.Fatalf("expected one resource, got %d", len(request.ResourceMetrics))
	}
	resource := request.ResourceMetrics[0]
	foundService := false
	for _, attr := range resource.Resource.Attributes {
		if attr.Key == "service.name" && attr.Value.GetStringValue() == CliName {
			foundService = true
		}
	}
	if !foundService {
		t.Error("service.name was not preserved")
	}
	foundCounter, foundHistogram := false, false
	for _, scope := range resource.ScopeMetrics {
		for _, m := range scope.Metrics {
			switch m.Name {
			case CliName + "_verify_audit_logging_error":
				foundCounter = m.GetSum().DataPoints[0].GetAsInt() == 1
			case CliName + "_command_usage":
				point := m.GetHistogram().DataPoints[0]
				foundHistogram = point.Count == 1 && point.GetSum() == 1
			}
		}
	}
	if !foundCounter || !foundHistogram {
		t.Fatalf("missing metrics: counter=%v, histogram=%v", foundCounter, foundHistogram)
	}
}

func TestInitializeOptOut(t *testing.T) {
	t.Setenv("DO_NOT_TRACK", "1")
	provider := otel.GetMeterProvider()
	flush := Initialize()
	if otel.GetMeterProvider() != provider {
		t.Fatal("opt out should not initialize a provider")
	}
	flush(false)
}

func TestTraceExportWaitsAndDoesNotRetry(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
	counter, err := provider.Meter(CliName).Int64Counter("test")
	if err != nil {
		t.Fatal(err)
	}
	counter.Add(t.Context(), 1)

	transport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = transport })
	started, release, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	calls := 0
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			close(started)
		}
		select {
		case <-release:
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
		return &http.Response{StatusCode: http.StatusServiceUnavailable, Header: make(http.Header), Body: http.NoBody}, nil
	})
	ctx, cancel := context.WithTimeout(t.Context(), uploadTimeout)
	defer cancel()
	go func() { done <- export(ctx, reader, true) }()
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("trace export did not start an HTTP request")
	}
	select {
	case err := <-done:
		t.Fatalf("trace export returned before upload completed: %v", err)
	default:
	}
	close(release)
	if err := <-done; err == nil {
		t.Fatal("expected upload failure")
	}
	if calls != 1 {
		t.Fatalf("expected one attempt, got %d", calls)
	}
}

func TestTraceReportsUploadFailure(t *testing.T) {
	t.Setenv("DO_NOT_TRACK", "0")
	provider, transport, stderr := otel.GetMeterProvider(), http.DefaultTransport, os.Stderr
	t.Cleanup(func() {
		otel.SetMeterProvider(provider)
		http.DefaultTransport, os.Stderr = transport, stderr
	})
	file, err := os.CreateTemp(t.TempDir(), "trace")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	os.Stderr = file
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusServiceUnavailable, Header: make(http.Header), Body: http.NoBody}, nil
	})
	flush := Initialize()
	CreateAndIncreaseCounter(t.Context(), "test")
	flush(true)
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Failed to upload metrics:") {
		t.Fatalf("expected upload failure diagnostics, got %q", data)
	}
}

func TestUploadHonorsDeadline(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		<-req.Context().Done()
		return nil, req.Context().Err()
	})}
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
	defer cancel()
	if err := upload(ctx, []byte("data"), client); err == nil {
		t.Fatal("expected deadline error")
	}
}
