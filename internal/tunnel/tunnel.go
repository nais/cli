package tunnel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type TunnelInfo struct {
	TunnelName       string
	TeamSlug         string
	EnvironmentName  string
	GatewayPublicKey wgtypes.Key
	GatewayEndpoint  string
	PrivateKey       wgtypes.Key
	WireGuardTunnel  *WireGuardTunnel
}

type Config struct {
	TeamSlug    string
	Environment string
	ListenAddr  string
	TargetHost  string
	TargetPort  int
}

func CreateAndConnect(ctx context.Context, cfg Config, progress func(string)) (*TunnelInfo, error) {
	privateKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("generate wireguard key: %w", err)
	}
	publicKey := privateKey.PublicKey()

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get graphql client: %w", err)
	}

	progress("Creating tunnel")

	createResp, err := gql.CreateTunnel(ctx, client, gql.CreateTunnelInput{
		TeamSlug:        cfg.TeamSlug,
		EnvironmentName: cfg.Environment,
		TargetHost:      cfg.TargetHost,
		TargetPort:      cfg.TargetPort,
		ClientPublicKey: publicKey.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("create tunnel: %w", err)
	}
	tunnelName := createResp.CreateTunnel.Tunnel.Name

	progress("Waiting for gateway")

	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var gatewayPublicKey string
	var forwarderEndpoint string
	var lastPollErr error

	for {
		select {
		case <-timeout:
			if lastPollErr != nil {
				return nil, fmt.Errorf("gateway did not become ready within 60s (last error: %v)", lastPollErr)
			}
			return nil, fmt.Errorf("gateway did not become ready within 60s — try again or check cluster status")
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			pollResp, err := gql.GetTunnel(ctx, client, cfg.TeamSlug, cfg.Environment, tunnelName)
			if err != nil {
				lastPollErr = err
				continue
			}
			lastPollErr = nil
			t := pollResp.Team.Environment.Tunnel
			if t.Id == "" {
				continue
			}
			switch phase := gql.TunnelPhase(strings.ToUpper(string(t.Phase))); phase {
			case gql.TunnelPhaseReady, gql.TunnelPhaseConnected:
				gatewayPublicKey = t.GatewayPublicKey
				forwarderEndpoint = t.ForwarderEndpoint
				progress("Gateway ready")
			case gql.TunnelPhaseFailed:
				return nil, fmt.Errorf("gateway failed to start: %s", t.Message)
			case gql.TunnelPhaseTerminated:
				return nil, fmt.Errorf("gateway terminated unexpectedly: %s", t.Message)
			default:
				progress(fmt.Sprintf("Gateway %s", t.Phase))
				continue
			}
		}
		if gatewayPublicKey != "" {
			break
		}
	}

	gwKey, err := wgtypes.ParseKey(gatewayPublicKey)
	if err != nil {
		return nil, fmt.Errorf("parse gateway public key: %w", err)
	}

	progress("Connecting WireGuard")

	wgTunnel, err := SetupWireGuard(privateKey, gwKey, forwarderEndpoint)
	if err != nil {
		return nil, fmt.Errorf("setup wireguard: %w", err)
	}

	return &TunnelInfo{
		TunnelName:       tunnelName,
		TeamSlug:         cfg.TeamSlug,
		EnvironmentName:  cfg.Environment,
		GatewayPublicKey: gwKey,
		GatewayEndpoint:  forwarderEndpoint,
		PrivateKey:       privateKey,
		WireGuardTunnel:  wgTunnel,
	}, nil
}

func DeleteTunnel(ctx context.Context, teamSlug, environmentName, tunnelName string) error {
	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return fmt.Errorf("get graphql client: %w", err)
	}
	_, err = gql.DeleteTunnel(ctx, client, teamSlug, environmentName, tunnelName)
	return err
}
