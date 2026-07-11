package nhncloud

import (
	"context"
	"fmt"
	"time"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/ports"
)

// FloatingIP is a normalized view of a Neutron floating IP, with the owning
// instance's name resolved via its port (Neutron doesn't expose the
// instance directly on the floating IP resource).
type FloatingIP struct {
	ID                string `json:"id"`
	FloatingIP        string `json:"floating_ip"`
	FixedIP           string `json:"fixed_ip"`
	Status            string `json:"status"`
	AttachedInstance  string `json:"attached_instance"`
	FloatingNetworkID string `json:"floating_network_id"`
	PortID            string `json:"port_id"`
	// Created is always the zero value: gophercloud's floatingips.FloatingIP
	// tags CreatedAt as json:"-" and never unmarshals it from the API
	// response. See the equivalent note in networks.go.
	Created time.Time `json:"created"`
}

// ListFloatingIPs returns every floating IP visible to the current tenant,
// regardless of status (unattached/spare floating IPs are useful to see
// too, per resource_checker/README.md). It resolves each floating IP's
// attached instance name by cross-referencing PortID -> Port.DeviceID ->
// Instance.Name, using one bulk ports.List and one bulk ListInstances(all)
// call rather than a request per floating IP.
func ListFloatingIPs(ctx context.Context, computeClient, networkClient *gophercloud.ServiceClient) ([]FloatingIP, error) {
	fipPages, err := floatingips.List(networkClient, floatingips.ListOpts{}).AllPages(ctx)
	if err != nil {
		return nil, fmt.Errorf("floating IP 조회 실패: %w", err)
	}
	rawFIPs, err := floatingips.ExtractFloatingIPs(fipPages)
	if err != nil {
		return nil, fmt.Errorf("floating IP 응답 파싱 실패: %w", err)
	}

	portToDevice, err := portDeviceIDs(ctx, networkClient)
	if err != nil {
		return nil, err
	}

	instanceNames, err := instanceNamesByID(ctx, computeClient)
	if err != nil {
		return nil, err
	}

	result := make([]FloatingIP, 0, len(rawFIPs))
	for _, f := range rawFIPs {
		var attached string
		if deviceID, ok := portToDevice[f.PortID]; ok {
			attached = instanceNames[deviceID]
		}
		result = append(result, FloatingIP{
			ID:                f.ID,
			FloatingIP:        f.FloatingIP,
			FixedIP:           f.FixedIP,
			Status:            f.Status,
			AttachedInstance:  attached,
			FloatingNetworkID: f.FloatingNetworkID,
			PortID:            f.PortID,
			Created:           f.CreatedAt,
		})
	}
	return result, nil
}

func portDeviceIDs(ctx context.Context, networkClient *gophercloud.ServiceClient) (map[string]string, error) {
	pages, err := ports.List(networkClient, ports.ListOpts{}).AllPages(ctx)
	if err != nil {
		return nil, fmt.Errorf("포트 조회 실패: %w", err)
	}
	raw, err := ports.ExtractPorts(pages)
	if err != nil {
		return nil, fmt.Errorf("포트 응답 파싱 실패: %w", err)
	}

	byID := make(map[string]string, len(raw))
	for _, p := range raw {
		if p.DeviceID != "" {
			byID[p.ID] = p.DeviceID
		}
	}
	return byID, nil
}

func instanceNamesByID(ctx context.Context, computeClient *gophercloud.ServiceClient) (map[string]string, error) {
	instances, err := ListInstances(ctx, computeClient, true)
	if err != nil {
		return nil, err
	}
	names := make(map[string]string, len(instances))
	for _, inst := range instances {
		names[inst.ID] = inst.Name
	}
	return names, nil
}
