package tunnel

import (
	"context"
	"fmt"
	"time"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type TunnelInfo struct {
	TunnelID         string
	GatewayPublicKey wgtypes.Key
	GatewayEndpoint  string
	PrivateKey       wgtypes.Key
	WireGuardTunnel  *WireGuardTunnel
}

type Config struct {
	TeamSlug     string
	Environment  string
	InstanceName string
	ListenAddr   string
	TargetHost   string
	TargetPort   int
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

	progress("Creating tunnel...")

	createResp, err := gql.CreateTunnel(ctx, client, gql.CreateTunnelInput{
		TeamSlug:        cfg.TeamSlug,
		EnvironmentName: cfg.Environment,
		InstanceName:    cfg.InstanceName,
		TargetHost:      cfg.TargetHost,
		TargetPort:      cfg.TargetPort,
		ClientPublicKey: publicKey.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("create tunnel: %w", err)
	}
	tunnelID := createResp.CreateTunnel.Tunnel.Id

	progress("Gateway starting...")

	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var gatewayPublicKey string
	var gatewaySTUNEndpoint string

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("gateway did not become ready within 60s — try again or check cluster status")
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			pollResp, err := gql.GetTunnel(ctx, client, cfg.TeamSlug, cfg.Environment, tunnelID)
			if err != nil {
				return nil, fmt.Errorf("poll tunnel status: %w", err)
			}
			t := pollResp.Team.Environment.Tunnel
			if t.Id == "" {
				continue
			}
			switch t.Phase {
			case gql.TunnelPhaseReady, gql.TunnelPhaseConnected:
				gatewayPublicKey = t.GatewayPublicKey
				gatewaySTUNEndpoint = t.GatewaySTUNEndpoint
				progress("Gateway ready!")
			case gql.TunnelPhaseFailed:
				return nil, fmt.Errorf("gateway failed to start: %s", t.Message)
			case gql.TunnelPhaseTerminated:
				return nil, fmt.Errorf("gateway terminated unexpectedly: %s", t.Message)
			default:
				progress(fmt.Sprintf("Gateway phase: %s...", t.Phase))
				continue
			}
		}
		if gatewayPublicKey != "" {
			break
		}
	}

	progress("Discovering STUN endpoint...")
	stunEndpoint, stunConn, err := DiscoverSTUNEndpoint(0)
	if err != nil {
		return nil, fmt.Errorf("STUN discovery failed: %w", err)
	}
	stunConnOwned := true
	defer func() {
		if stunConnOwned {
			_ = stunConn.Close()
		}
	}()

	_, err = gql.UpdateTunnelSTUNEndpoint(ctx, client, tunnelID, stunEndpoint)
	if err != nil {
		return nil, fmt.Errorf("update STUN endpoint: %w", err)
	}

	gwKey, err := wgtypes.ParseKey(gatewayPublicKey)
	if err != nil {
		return nil, fmt.Errorf("parse gateway public key: %w", err)
	}

	wgTunnel, err := SetupWireGuardWithConn(privateKey, gwKey, gatewaySTUNEndpoint, stunConn)
	if err != nil {
		return nil, fmt.Errorf("setup wireguard: %w", err)
	}
	stunConnOwned = false

	return &TunnelInfo{
		TunnelID:         tunnelID,
		GatewayPublicKey: gwKey,
		GatewayEndpoint:  gatewaySTUNEndpoint,
		PrivateKey:       privateKey,
		WireGuardTunnel:  wgTunnel,
	}, nil
}

func DeleteTunnel(ctx context.Context, tunnelID string) error {
	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return fmt.Errorf("get graphql client: %w", err)
	}
	_, err = gql.DeleteTunnel(ctx, client, tunnelID)
	return err
}
