package metric

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// The detached child is this test executable. Simulate a slow upload without
// network access, and deliberately write to both streams to detect pipe leaks.
func TestMain(m *testing.M) {
	if len(os.Args) == 2 && os.Args[1] == uploadArg {
		path := os.Getenv("NAIS_TEST_UPLOAD_FILE")
		if path == "" {
			os.Exit(2) // Never upload real telemetry from tests.
		}
		payload, err := base64.StdEncoding.DecodeString(os.Getenv(uploadPayloadEnv))
		if err != nil {
			os.Exit(3)
		}
		stdin, err := io.ReadAll(os.Stdin)
		if err != nil || len(stdin) != 0 {
			os.Exit(4)
		}
		fmt.Println("child stdout must not escape")
		fmt.Fprintln(os.Stderr, "child stderr must not escape")
		if err := os.WriteFile(path, payload, 0o600); err != nil {
			os.Exit(5)
		}
		if release := os.Getenv("NAIS_TEST_UPLOAD_RELEASE"); release != "" {
			deadline := time.Now().Add(10 * time.Second)
			for time.Now().Before(deadline) {
				if _, err := os.Stat(release); err == nil {
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
		_ = os.WriteFile(path+".done", nil, 0o600)
		os.Exit(0)
	}
	if os.Getenv("NAIS_TEST_UPLOAD_PARENT") == "1" {
		if err := startUploader([]byte("test payload")); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestDetachedUploaderDoesNotHoldPipelinesOpen(t *testing.T) {
	dir := t.TempDir()
	path, release := filepath.Join(dir, "payload"), filepath.Join(dir, "release")
	t.Setenv("DO_NOT_TRACK", "0")
	t.Setenv("NAIS_TEST_UPLOAD_FILE", path)
	t.Setenv("NAIS_TEST_UPLOAD_RELEASE", release)
	t.Setenv("NAIS_TEST_UPLOAD_PARENT", "1")
	t.Cleanup(func() {
		_ = os.WriteFile(release, nil, 0o600)
		waitForFile(t, path+".done")
	})

	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable)
	cmd.WaitDelay = time.Second
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("parent did not exit independently of uploader: %v; output: %s", err, out)
	}
	if len(out) != 0 {
		t.Fatalf("uploader leaked output: %s", out)
	}
	if got := string(waitForFile(t, path)); got != "test payload" {
		t.Fatalf("payload = %q", got)
	}
	if _, err := os.Stat(path + ".done"); !os.IsNotExist(err) {
		t.Fatal("uploader should still be waiting after parent exits")
	}
}

func TestStartUploaderOptOutAndSizeLimit(t *testing.T) {
	t.Setenv("DO_NOT_TRACK", "1")
	if err := startUploader([]byte("unused")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DO_NOT_TRACK", "0")
	if err := startUploader(make([]byte, maxPayloadSize+1)); err == nil {
		t.Fatal("expected oversized payload to be rejected")
	}
}

func TestRunUploader(t *testing.T) {
	for _, tc := range []struct {
		name    string
		args    []string
		payload string
		optOut  string
		handled bool
		calls   int
	}{
		{name: "normal command", args: []string{"nais", "app"}},
		{name: "extra arguments", args: []string{"nais", uploadArg, "extra"}},
		{name: "missing payload", args: []string{"nais", uploadArg}, handled: true},
		{name: "invalid payload", args: []string{"nais", uploadArg}, payload: "!", handled: true},
		{name: "oversized payload", args: []string{"nais", uploadArg}, payload: base64.StdEncoding.EncodeToString(make([]byte, maxPayloadSize+1)), handled: true},
		{name: "opt out", args: []string{"nais", uploadArg}, payload: base64.StdEncoding.EncodeToString([]byte("data")), optOut: "1", handled: true},
		{name: "upload", args: []string{"nais", uploadArg}, payload: base64.StdEncoding.EncodeToString([]byte("data")), handled: true, calls: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args, transport := os.Args, http.DefaultTransport
			t.Cleanup(func() { os.Args, http.DefaultTransport = args, transport })
			os.Args = tc.args
			t.Setenv(uploadPayloadEnv, tc.payload)
			t.Setenv("DO_NOT_TRACK", tc.optOut)
			calls := 0
			http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				calls++
				if req.URL.String() != collectorURL || req.Method != http.MethodPost {
					t.Errorf("unexpected request: %s %s", req.Method, req.URL)
				}
				if req.Header.Get("Content-Type") != "application/x-protobuf" {
					t.Error("missing protobuf content type")
				}
				deadline, ok := req.Context().Deadline()
				if !ok || time.Until(deadline) > uploadTimeout {
					t.Error("missing upload deadline")
				}
				body, _ := io.ReadAll(req.Body)
				if string(body) != "data" {
					t.Errorf("body = %q", body)
				}
				// An error must not trigger a retry or escape RunUploader.
				return nil, errors.New("upload failed")
			})
			if got := RunUploader(); got != tc.handled {
				t.Errorf("handled = %v, want %v", got, tc.handled)
			}
			if calls != tc.calls {
				t.Errorf("upload calls = %d, want %d", calls, tc.calls)
			}
		})
	}
}

func TestBackgroundTransportRejectsOversizedBody(t *testing.T) {
	t.Setenv("DO_NOT_TRACK", "0")
	req, err := http.NewRequest(http.MethodPost, collectorURL, bytes.NewReader(make([]byte, maxPayloadSize+1)))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (backgroundTransport{}).RoundTrip(req); err == nil {
		t.Fatal("expected oversized payload error")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func waitForFile(t *testing.T, path string) []byte {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(path)
		if err == nil {
			return data
		}
		if !os.IsNotExist(err) {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", path)
	return nil
}
