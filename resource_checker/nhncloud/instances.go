package nhncloud

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
)

// Instance is a normalized view of a Nova server, with only the fields
// resource_checker/README.md's "## CLI Mode" section documents.
type Instance struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Status           string    `json:"status"`
	FlavorID         string    `json:"flavor_id"`
	ImageID          string    `json:"image_id"`
	PrivateIPs       []string  `json:"private_ips"`
	FloatingIPs      []string  `json:"floating_ips"`
	SecurityGroups   []string  `json:"security_groups"`
	AvailabilityZone string    `json:"availability_zone"`
	Created          time.Time `json:"created"`
}

// ListInstances returns instances. When all is false, only ACTIVE instances
// are returned (matches the README's documented default); when all is true,
// every state (SHUTOFF, ERROR, BUILD, ...) is included.
func ListInstances(ctx context.Context, client *gophercloud.ServiceClient, all bool) ([]Instance, error) {
	opts := servers.ListOpts{}
	if !all {
		opts.Status = "ACTIVE"
	}

	pages, err := servers.List(client, opts).AllPages(ctx)
	if err != nil {
		return nil, fmt.Errorf("인스턴스 조회 실패: %w", err)
	}
	raw, err := servers.ExtractServers(pages)
	if err != nil {
		return nil, fmt.Errorf("인스턴스 응답 파싱 실패: %w", err)
	}

	instances := make([]Instance, 0, len(raw))
	for _, s := range raw {
		privateIPs, floatingIPs := splitAddresses(s.Addresses)
		instances = append(instances, Instance{
			ID:               s.ID,
			Name:             s.Name,
			Status:           s.Status,
			FlavorID:         stringField(s.Flavor, "id"),
			ImageID:          stringField(s.Image, "id"),
			PrivateIPs:       privateIPs,
			FloatingIPs:      floatingIPs,
			SecurityGroups:   secGroupNames(s.SecurityGroups),
			AvailabilityZone: s.AvailabilityZone,
			Created:          s.Created,
		})
	}
	return instances, nil
}

// splitAddresses separates a Nova "addresses" map (keyed by network name,
// valued by a list of address objects) into fixed and floating IP strings.
func splitAddresses(addresses map[string]any) (private, floating []string) {
	for _, v := range addresses {
		entries, ok := v.([]any)
		if !ok {
			continue
		}
		for _, e := range entries {
			addr, ok := e.(map[string]any)
			if !ok {
				continue
			}
			ip, _ := addr["addr"].(string)
			if ip == "" {
				continue
			}
			ipType, _ := addr["OS-EXT-IPS:type"].(string)
			if ipType == "floating" {
				floating = append(floating, ip)
			} else {
				private = append(private, ip)
			}
		}
	}
	sort.Strings(private)
	sort.Strings(floating)
	return private, floating
}

func stringField(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, _ := m[key].(string)
	return v
}

func secGroupNames(groups []map[string]any) []string {
	names := make([]string, 0, len(groups))
	for _, g := range groups {
		if name, ok := g["name"].(string); ok && name != "" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// JoinOrDash formats a string slice for table/text display: comma-joined,
// or "-" when empty.
func JoinOrDash(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ",")
}
