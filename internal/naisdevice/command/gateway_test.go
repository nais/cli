package command

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisdevice/command/flag"
	"github.com/nais/device/pkg/config"
	"github.com/nais/device/pkg/pb"
	"github.com/nais/naistrix"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gatewayTestAgent struct {
	pb.UnimplementedDeviceAgentServer
	status *pb.AgentStatus
	err    error
	calls  atomic.Int32
}

func (a *gatewayTestAgent) Status(_ *pb.AgentStatusRequest, stream grpc.ServerStreamingServer[pb.AgentStatus]) error {
	a.calls.Add(1)
	if a.err != nil {
		return a.err
	}
	return stream.Send(a.status)
}

func TestGatewayDescribeChoices(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the fake agent uses a Unix socket")
	}
	// Keep the socket path short enough for Unix socket limits, regardless of subtest names.
	configDir := t.TempDir()
	t.Setenv("HOME", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)
	agentDir, err := config.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(agentDir, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name, gateway, output, prefix, wantError string
		completion, afterValue, empty            bool
		disconnected, rpcError                   bool
		wantCalls                                int32
	}{
		{name: "existing gateway", gateway: "alpha", wantCalls: 2},
		{name: "JSON output", gateway: "alpha", output: "json", wantCalls: 2},
		{name: "YAML output", gateway: "alpha", output: "yaml", wantCalls: 2},
		{name: "unknown gateway", gateway: "unknown", wantError: "invalid value", wantCalls: 1},
		{name: "gateway names remain case sensitive", gateway: "ALPHA", wantError: "invalid value", wantCalls: 1},
		{name: "no available gateways", gateway: "alpha", empty: true, wantError: "no choices available", wantCalls: 1},
		{name: "disconnected agent", gateway: "alpha", disconnected: true, wantError: "not connected to naisdevice", wantCalls: 1},
		{name: "agent error preserved", gateway: "alpha", rpcError: true, wantError: "access denied", wantCalls: 1},
		{name: "gateway completion", completion: true, wantCalls: 1},
		{name: "completion prefix", completion: true, prefix: "be", wantCalls: 1},
		{name: "completion reports empty list", completion: true, empty: true, wantError: "no choices available", wantCalls: 1},
		{name: "completion reports disconnection", completion: true, disconnected: true, wantError: "not connected to naisdevice", wantCalls: 1},
		{name: "completion reports agent error", completion: true, rpcError: true, wantError: "access denied", wantCalls: 1},
		{name: "no lookup after final argument", completion: true, afterValue: true, gateway: "alpha"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			agent := &gatewayTestAgent{status: &pb.AgentStatus{
				ConnectionState: pb.AgentState_Connected,
				Gateways: []*pb.Gateway{
					{Name: "alpha", Endpoint: "alpha.example.com:51820", Healthy: true},
					{Name: "beta", RequiresPrivilegedAccess: true},
				},
			}}
			if tt.empty {
				agent.status.Gateways = nil
			}
			if tt.disconnected {
				agent.status.ConnectionState = pb.AgentState_Disconnected
			}
			if tt.rpcError {
				agent.err = status.Error(codes.PermissionDenied, "access denied")
			}
			listener, err := net.Listen("unix", filepath.Join(agentDir, "agent.sock"))
			if err != nil {
				t.Fatal(err)
			}
			server := grpc.NewServer()
			pb.RegisterDeviceAgentServer(server, agent)
			serveErr := make(chan error, 1)
			go func() { serveErr <- server.Serve(listener) }()
			t.Cleanup(func() {
				server.Stop()
				if err := <-serveErr; err != nil && !errors.Is(err, grpc.ErrServerStopped) {
					t.Errorf("serving fake agent: %v", err)
				}
			})
			var output bytes.Buffer
			app, global, err := naistrix.NewApplication("test", "Test gateway choices", "test", naistrix.ApplicationWithWriter(&output))
			if err != nil {
				t.Fatal(err)
			}
			global.Config = filepath.Join(configDir, "config.yaml")
			root := gatewaycmd(&flag.Device{GlobalFlags: &flags.GlobalFlags{GlobalFlags: global}})
			describe := root.SubCommands[1]
			if describe.AutoCompleteFunc != nil || describe.Args[0].ChoicesFunc == nil {
				t.Fatal("gateway describe must use automatic argument choices")
			}
			if err := app.AddCommand(root); err != nil {
				t.Fatal(err)
			}
			args := []string{"gateway", "describe"}
			if tt.output != "" {
				args = append(args, "--output", tt.output)
			}
			if tt.completion {
				args = append([]string{"__complete"}, args...)
				if tt.afterValue {
					args = append(args, tt.gateway)
				}
				args = append(args, tt.prefix)
			} else {
				args = append(args, tt.gateway)
			}
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			err = app.Run(naistrix.RunWithArgs(args), naistrix.RunWithContext(ctx))
			got := output.String()
			if tt.completion {
				if err != nil {
					t.Fatal(err)
				}
				if tt.wantError != "" {
					if !strings.HasPrefix(got, "_activeHelp_ Completion failed:") || !strings.Contains(got, tt.wantError) || strings.Count(got, "\n") != 2 {
						t.Errorf("completion = %q, want one active help error containing %q", got, tt.wantError)
					}
				} else if tt.afterValue {
					if got != ":4\n" {
						t.Errorf("completion = %q, want no suggestions", got)
					}
				} else {
					wantNames := "alpha\nbeta\n"
					if tt.prefix != "" {
						wantNames = "beta\n"
					}
					want := wantNames + "_activeHelp_ Name of the gateway to inspect\n:4\n"
					if got != want {
						t.Errorf("completion = %q, want %q", got, want)
					}
				}
				if !strings.HasSuffix(got, ":4\n") {
					t.Errorf("completion = %q, want NoFileComp", got)
				}
			} else if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("error = %v, want %q", err, tt.wantError)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				switch tt.output {
				case "json":
					var gateway struct{ Name, Endpoint string }
					if err := json.Unmarshal(output.Bytes(), &gateway); err != nil {
						t.Fatalf("invalid JSON output %q: %v", got, err)
					}
					if gateway.Name != "alpha" || gateway.Endpoint != "alpha.example.com:51820" {
						t.Errorf("incorrect gateway details: %+v", gateway)
					}
				case "yaml":
					if !strings.Contains(got, "name: alpha") || !strings.Contains(got, "endpoint: alpha.example.com:51820") {
						t.Errorf("incorrect YAML gateway details: %q", got)
					}
				default:
					if !strings.Contains(got, "alpha") || !strings.Contains(got, "alpha.example.com:51820") {
						t.Errorf("incorrect gateway details: %q", got)
					}
				}
			}
			if calls := agent.calls.Load(); calls != tt.wantCalls {
				t.Errorf("agent calls = %d, want %d", calls, tt.wantCalls)
			}
		})
	}
}
