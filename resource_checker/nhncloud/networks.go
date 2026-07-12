package nhncloud

import (
	"context"
	"fmt"
	"time"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/external"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/mtu"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/provider"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
)

// Network is a normalized view of a Neutron network.
type Network struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Status       string   `json:"status"`
	External     bool     `json:"external"`
	Shared       bool     `json:"shared"`
	Subnets      []string `json:"subnets"`
	MTU          int      `json:"mtu"`
	ProjectID    string   `json:"project_id"`
	ProviderType string   `json:"provider_network_type"`
	// Created is always the zero value: gophercloud's networks.Network
	// tags CreatedAt as json:"-" and never unmarshals it from the API
	// response. Kept (rather than dropped) so the README's documented
	// --long CREATED column has a field to render as "-".
	Created time.Time `json:"created"`
}

type networkWithExt struct {
	networks.Network
	external.NetworkExternalExt
	mtu.NetworkMTUExt
	provider.NetworkProviderExt
}

// ListNetworks returns every network visible to the current tenant.
func ListNetworks(ctx context.Context, client *gophercloud.ServiceClient) ([]Network, error) {
	pages, err := networks.List(client, networks.ListOpts{}).AllPages(ctx)
	if err != nil {
		return nil, fmt.Errorf("네트워크 조회 실패: %w", err)
	}

	var raw []networkWithExt
	if err := networks.ExtractNetworksInto(pages, &raw); err != nil {
		return nil, fmt.Errorf("네트워크 응답 파싱 실패: %w", err)
	}

	result := make([]Network, 0, len(raw))
	for _, n := range raw {
		result = append(result, Network{
			ID:           n.ID,
			Name:         n.Name,
			Status:       n.Status,
			External:     n.External,
			Shared:       n.Shared,
			Subnets:      n.Subnets,
			MTU:          n.MTU,
			ProjectID:    n.ProjectID,
			ProviderType: n.NetworkType,
			Created:      n.CreatedAt,
		})
	}
	return result, nil
}
