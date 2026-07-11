package nhncloud

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/security/groups"
)

// SecurityGroupRule is a normalized view of one rule within a security group.
type SecurityGroupRule struct {
	Direction    string `json:"direction"`
	EtherType    string `json:"ethertype"`
	Protocol     string `json:"protocol"`
	PortRangeMin int    `json:"port_range_min"`
	PortRangeMax int    `json:"port_range_max"`
	// Remote is either a CIDR (RemoteIPPrefix) or a security group ID
	// (RemoteGroupID), whichever the rule has set.
	Remote string `json:"remote"`
}

// SecurityGroup is a normalized view of a Neutron security group.
type SecurityGroup struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Rules       []SecurityGroupRule `json:"rules"`
}

// ListSecurityGroups returns every security group visible to the current
// tenant, including their rules.
func ListSecurityGroups(ctx context.Context, client *gophercloud.ServiceClient) ([]SecurityGroup, error) {
	pages, err := groups.List(client, groups.ListOpts{}).AllPages(ctx)
	if err != nil {
		return nil, fmt.Errorf("보안그룹 조회 실패: %w", err)
	}
	raw, err := groups.ExtractGroups(pages)
	if err != nil {
		return nil, fmt.Errorf("보안그룹 응답 파싱 실패: %w", err)
	}

	result := make([]SecurityGroup, 0, len(raw))
	for _, g := range raw {
		rules := make([]SecurityGroupRule, 0, len(g.Rules))
		for _, r := range g.Rules {
			remote := r.RemoteIPPrefix
			if remote == "" {
				remote = r.RemoteGroupID
			}
			rules = append(rules, SecurityGroupRule{
				Direction:    r.Direction,
				EtherType:    r.EtherType,
				Protocol:     r.Protocol,
				PortRangeMin: r.PortRangeMin,
				PortRangeMax: r.PortRangeMax,
				Remote:       remote,
			})
		}
		result = append(result, SecurityGroup{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			Rules:       rules,
		})
	}
	return result, nil
}

// RuleCountSummary formats "N inbound / M outbound" for a security group's
// default (non-long) table/text column.
func RuleCountSummary(rules []SecurityGroupRule) string {
	var inbound, outbound int
	for _, r := range rules {
		switch r.Direction {
		case "ingress":
			inbound++
		case "egress":
			outbound++
		}
	}
	return fmt.Sprintf("%d inbound / %d outbound", inbound, outbound)
}
