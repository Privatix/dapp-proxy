package monitor

import (
	"context"

	v2rayclient "github.com/privatix/dapp-proxy/v2ray-client"
)

// V2RayClientUsageGetter gets traffic usage using v2ray api.
type V2RayClientUsageGetter struct {
	client *v2rayclient.StatsClient
}

// NewV2RayClientUsageGetter creates an instance.
func NewV2RayClientUsageGetter(client *v2rayclient.StatsClient) *V2RayClientUsageGetter {
	return &V2RayClientUsageGetter{client}
}

// Get returns traffic usage for username.
func (getter *V2RayClientUsageGetter) Get(username string) (uint64, error) {
	return getter.client.GetUsage(context.Background(), username)
}
