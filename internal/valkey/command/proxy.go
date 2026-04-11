package command

import (
	"context"
	"fmt"
	"io"
	"net"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/tunnel"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/cli/internal/valkey"
	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
)

func proxy(parentFlags *flag.Valkey) *naistrix.Command {
	flags := &flag.Proxy{
		Valkey:     parentFlags,
		ListenAddr: "localhost:6379",
	}

	return &naistrix.Command{
		Name:        "proxy",
		Title:       "Create a proxy to a Valkey instance.",
		Description: "Allows your user to connect to Valkey instances and starts a proxy.",
		Flags:       flags,
		ValidateFunc: func(ctx context.Context, args *naistrix.Arguments) error {
			if err := validateSingleEnvironmentFlagUsage(); err != nil {
				return err
			}
			if err := validation.CheckEnvironment(string(flags.Environment)); err != nil {
				return err
			}
			if flags.Instance == "" {
				return fmt.Errorf("--instance flag is required")
			}
			return nil
		},
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			return autoCompleteValkeyNames(ctx, flags.Team, string(flags.Environment), true)
		},
		Examples: []naistrix.Example{
			{
				Description: "Create a proxy to a Valkey instance named my-valkey in environment dev.",
				Command:     "proxy --instance my-valkey --environment dev",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			creds, err := valkey.CreateCredentials(
				ctx,
				flags.Team,
				string(flags.Environment),
				flags.Instance,
				gql.CredentialPermissionReadwrite,
				"1h",
			)
			if err != nil {
				return fmt.Errorf("get valkey credentials: %w", err)
			}

			tunnelInfo, err := tunnel.CreateAndConnect(ctx, tunnel.Config{
				TeamSlug:     flags.Team,
				Environment:  string(flags.Environment),
				InstanceName: flags.Instance,
				ListenAddr:   flags.ListenAddr,
				TargetHost:   creds.Host,
				TargetPort:   creds.Port,
			}, func(msg string) { out.Infof("%s\n", msg) })
			if err != nil {
				return fmt.Errorf("create tunnel: %w", err)
			}
			defer tunnelInfo.WireGuardTunnel.Close()
			defer tunnel.DeleteTunnel(context.Background(), tunnelInfo.TunnelID) //nolint:errcheck

			ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
			defer stop()

			lc := net.ListenConfig{}
			listener, err := lc.Listen(ctx, "tcp", flags.ListenAddr)
			if err != nil {
				return fmt.Errorf("listen on %s: %w", flags.ListenAddr, err)
			}
			defer listener.Close()

			out.Infof("Listening on %s, forwarding to %s via WireGuard tunnel\n",
				listener.Addr().String(), flags.Instance)

			go func() {
				<-ctx.Done()
				listener.Close()
			}()

			gatewayAddr := fmt.Sprintf("10.0.0.2:%d", creds.Port)
			var wg sync.WaitGroup
			for ctx.Err() == nil {
				conn, err := listener.Accept()
				if err != nil {
					if ctx.Err() != nil {
						break
					}
					out.Infof("accept error: %v\n", err)
					continue
				}

				wg.Add(1)
				go func() {
					defer wg.Done()
					defer conn.Close()

					remote, err := tunnelInfo.WireGuardTunnel.DialTCP(gatewayAddr)
					if err != nil {
						out.Infof("dial through tunnel: %v\n", err)
						return
					}
					defer remote.Close()

					done := make(chan struct{}, 2)
					go func() {
						io.Copy(remote, conn) //nolint:errcheck
						done <- struct{}{}
					}()
					go func() {
						io.Copy(conn, remote) //nolint:errcheck
						done <- struct{}{}
					}()
					<-done
				}()
			}

			drainDone := make(chan struct{})
			go func() {
				wg.Wait()
				close(drainDone)
			}()
			select {
			case <-drainDone:
			case <-time.After(5 * time.Second):
			}
			return nil
		},
	}
}
