package monitor

import (
	"context"

	v2rayclient "github.com/privatix/dapp-proxy/plugin/v2ray-client"
)

// UsageGetterAdapter meant to hide timeout like logic from monitor.
type UsageGetterAdapter struct {
	getter *v2rayclient.UsageGetter
}

// NewUsageGetterAdapter creates an instance.
func NewUsageGetterAdapter(getter *v2rayclient.UsageGetter) *UsageGetterAdapter {
	return &UsageGetterAdapter{getter}
}

// Get returns traffic usage.
func (g *UsageGetterAdapter) Get() (uint64, error) {
	return g.getter.GetUsage(context.Background())
}
