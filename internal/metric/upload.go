package metric

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

const (
	uploadArg        = "__upload-metrics"
	uploadPayloadEnv = "_NAIS_METRICS_PAYLOAD"
	// Keep the handoff well below OS argument/environment limits. Oversized
	// telemetry is dropped rather than delaying or failing the user's command.
	maxPayloadSize = 8 * 1024
)

// RunUploader handles the private child entry point before CLI initialization.
// The encoded OTLP body travels in the environment, not the command line.
// The handoff contains metrics, never the original CLI arguments.
func RunUploader() bool {
	if len(os.Args) != 2 || os.Args[1] != uploadArg {
		return false
	}
	if os.Getenv("DO_NOT_TRACK") == "1" {
		return true
	}

	encoded := os.Getenv(uploadPayloadEnv)
	_ = os.Unsetenv(uploadPayloadEnv)
	if len(encoded) > base64.StdEncoding.EncodedLen(maxPayloadSize) {
		return true
	}
	payload, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(payload) == 0 || len(payload) > maxPayloadSize {
		return true
	}

	ctx, cancel := context.WithTimeout(context.Background(), uploadTimeout)
	defer cancel()
	_ = upload(ctx, payload, &http.Client{Timeout: uploadTimeout})
	return true
}

func upload(ctx context.Context, payload []byte, client *http.Client) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, collectorURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-protobuf")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("metrics upload: HTTP %d", resp.StatusCode)
	}
	return nil
}

// backgroundTransport lets the OTLP exporter do its normal encoding, but hands
// off the body instead of making an HTTP request in the foreground process.
type backgroundTransport struct{}

func (backgroundTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	defer func() { _ = req.Body.Close() }()
	payload, err := io.ReadAll(io.LimitReader(req.Body, maxPayloadSize+1))
	if err != nil {
		return nil, err
	}
	if err := startUploader(payload); err != nil {
		return nil, err
	}
	return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: http.NoBody, Request: req}, nil
}

func startUploader(payload []byte) error {
	if os.Getenv("DO_NOT_TRACK") == "1" {
		return nil
	}
	if len(payload) > maxPayloadSize {
		return fmt.Errorf("metrics payload exceeds %d bytes", maxPayloadSize)
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(executable, uploadArg)
	cmd.Env = append(os.Environ(), uploadPayloadEnv+"="+base64.StdEncoding.EncodeToString(payload))
	cmd.SysProcAttr = detachedProcessAttributes()
	// Nil standard streams connect to the null device: the child cannot hold
	// shell pipelines open or write to the user's terminal. Never Wait here.
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}
