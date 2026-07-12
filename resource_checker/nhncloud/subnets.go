package nhncloud

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/subnets"
)

// Subnet is a normalized view of a Neutron subnet.
type Subnet struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	CIDR            string   `json:"cidr"`
	GatewayIP       string   `json:"gateway_ip"`
	NetworkID       string   `json:"network_id"`
	EnableDHCP      bool     `json:"enable_dhcp"`
	AllocationPools []string `json:"allocation_pools"`
	DNSNameservers  []string `json:"dns_nameservers"`
	IPVersion       int      `json:"ip_version"`
}

// ListSubnets returns every subnet visible to the current tenant.
func ListSubnets(ctx context.Context, client *gophercloud.ServiceClient) ([]Subnet, error) {
	pages, err := subnets.List(client, subnets.ListOpts{}).AllPages(ctx)
	if err != nil {
		return nil, fmt.Errorf("서브넷 조회 실패: %w", err)
	}
	raw, err := subnets.ExtractSubnets(pages)
	if err != nil {
		return nil, fmt.Errorf("서브넷 응답 파싱 실패: %w", err)
	}

	result := make([]Subnet, 0, len(raw))
	for _, s := range raw {
		pools := make([]string, 0, len(s.AllocationPools))
		for _, p := range s.AllocationPools {
			pools = append(pools, fmt.Sprintf("%s-%s", p.Start, p.End))
		}
		result = append(result, Subnet{
			ID:              s.ID,
			Name:            s.Name,
			CIDR:            s.CIDR,
			GatewayIP:       s.GatewayIP,
			NetworkID:       s.NetworkID,
			EnableDHCP:      s.EnableDHCP,
			AllocationPools: pools,
			DNSNameservers:  s.DNSNameservers,
			IPVersion:       s.IPVersion,
		})
	}
	return result, nil
}
