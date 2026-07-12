// Package nhncloud wraps github.com/gophercloud/gophercloud/v2 to talk to
// NHN Cloud's OpenStack-compatible Compute (Nova) and Network (Neutron)
// APIs. Every call in this package is read-only (GET), per
// .claude/rules/resource-checker.md.
package nhncloud

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"

	"github.com/opp-13/nhn_iac/resource_checker/config"
)

// identityEndpoint is NHN Cloud's OpenStack Identity v2 token endpoint. It
// matches the tenantId/username/password auth flow already documented in
// this repo's READMEs. NHN Cloud's per-region behavior against this single
// host is unverified against a real account — see .claude/plans (implementation
// plan) for the caveat.
const identityEndpoint = "https://api-identity-infrastructure.nhncloudservice.com/v2.0"

// Clients holds the service clients this package's callers need.
type Clients struct {
	Compute *gophercloud.ServiceClient
	Network *gophercloud.ServiceClient
}

// NewClients authenticates against NHN Cloud using cfg's credentials and
// returns ready-to-use Compute and Network service clients.
func NewClients(ctx context.Context, cfg *config.Config) (*Clients, error) {
	if cfg.Nhn.Auth.TenantID == "" || cfg.Nhn.Auth.Username == "" {
		return nil, fmt.Errorf("credential이 설정되지 않았습니다 — 먼저 'rescheck configure set'을 실행하세요")
	}

	authOpts := gophercloud.AuthOptions{
		IdentityEndpoint: identityEndpoint,
		Username:         cfg.Nhn.Auth.Username,
		Password:         cfg.Nhn.Auth.Password,
		TenantID:         cfg.Nhn.Auth.TenantID,
	}

	provider, err := openstack.AuthenticatedClient(ctx, authOpts)
	if err != nil {
		return nil, fmt.Errorf("nhn cloud 인증 실패: %w", err)
	}

	endpointOpts := gophercloud.EndpointOpts{Region: cfg.Nhn.Auth.Region}

	compute, err := openstack.NewComputeV2(provider, endpointOpts)
	if err != nil {
		return nil, fmt.Errorf("compute 클라이언트 생성 실패: %w", err)
	}
	network, err := openstack.NewNetworkV2(provider, endpointOpts)
	if err != nil {
		return nil, fmt.Errorf("network 클라이언트 생성 실패: %w", err)
	}

	return &Clients{Compute: compute, Network: network}, nil
}
