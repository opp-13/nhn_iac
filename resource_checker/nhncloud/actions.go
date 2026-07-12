package nhncloud

// This file holds the only mutating operations allowed in resource_checker:
// starting and stopping instances (see .claude/rules/resource-checker.md).
// Everything else in this package must stay read-only (GET).

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
)

// StartInstance starts the instance with the given ID (Nova os-start).
func StartInstance(ctx context.Context, client *gophercloud.ServiceClient, id string) error {
	if err := servers.Start(ctx, client, id).ExtractErr(); err != nil {
		return fmt.Errorf("인스턴스 시작 실패 (%s): %w", id, err)
	}
	return nil
}

// StopInstance stops the instance with the given ID (Nova os-stop).
func StopInstance(ctx context.Context, client *gophercloud.ServiceClient, id string) error {
	if err := servers.Stop(ctx, client, id).ExtractErr(); err != nil {
		return fmt.Errorf("인스턴스 정지 실패 (%s): %w", id, err)
	}
	return nil
}
