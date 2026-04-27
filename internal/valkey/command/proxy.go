package command

import (
	"context"
	"fmt"
	"io"
	"net"
	"os/signal"
	"sync"
	"syscall"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/tunnel"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/cli/internal/valkey"
	"github.com/nais/cli/internal/valkey/command/flag"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
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

			spinner, _ := pterm.DefaultSpinner.Start("Creating tunnel")

			tunnelInfo, err := tunnel.CreateAndConnect(ctx, tunnel.Config{
				TeamSlug:    flags.Team,
				Environment: string(flags.Environment),
				ListenAddr:  flags.ListenAddr,
				TargetHost:  creds.Host,
				TargetPort:  creds.Port,
			}, func(msg string) { spinner.UpdateText(msg) })
			if err != nil {
				spinner.Fail()
				return fmt.Errorf("create tunnel: %w", err)
			}
			defer tunnelInfo.WireGuardTunnel.Close()
			defer tunnel.DeleteTunnel(context.Background(), tunnelInfo.TeamSlug, tunnelInfo.EnvironmentName, tunnelInfo.TunnelName) //nolint:errcheck

			ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
			defer stop()

			lc := net.ListenConfig{}
			listener, err := lc.Listen(ctx, "tcp", flags.ListenAddr)
			if err != nil {
				spinner.Fail()
				return fmt.Errorf("listen on %s: %w", flags.ListenAddr, err)
			}
			defer listener.Close()

			spinner.Success(fmt.Sprintf("Listening on %s, forwarding to %s via WireGuard tunnel",
				listener.Addr().String(), flags.Instance))

			go func() {
				<-ctx.Done()
				listener.Close() // #nosec G104 -- best-effort shutdown on context cancellation
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

				wg.Go(func() {
					defer conn.Close()

					remote, err := tunnelInfo.WireGuardTunnel.DialTCP(gatewayAddr)
					if err != nil {
						out.Infof("dial through tunnel: %v\n", err)
						return
					}
					defer remote.Close()

					var copyWg sync.WaitGroup
					copyWg.Go(func() { io.Copy(remote, conn) }) // #nosec G104 //nolint:errcheck
					copyWg.Go(func() { io.Copy(conn, remote) }) // #nosec G104 //nolint:errcheck
					copyWg.Wait()
				})
			}

			return nil
		},
	}
}
