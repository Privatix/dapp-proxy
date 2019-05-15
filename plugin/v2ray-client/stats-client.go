package v2rayclient

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"v2ray.com/core/app/stats/command"
)

// UsageGetter handles usage requests to v2ray api.
type UsageGetter struct {
	queryNames []string
	key        string
	stats      command.StatsServiceClient
}

// NewInboundUsageGetter creates a client to get usages of an inbound.
func NewInboundUsageGetter(conn *grpc.ClientConn, inboundTag string) *UsageGetter {
	return &UsageGetter{
		queryNames: []string{
			fmt.Sprintf("inbound>>>%s>>>traffic>>>downlink", inboundTag),
			fmt.Sprintf("inbound>>>%s>>>traffic>>>downlink", inboundTag),
		},
		key:   inboundTag,
		stats: command.NewStatsServiceClient(conn),
	}
}

// NewUserUsageGetter creates a client to get usages of an inbound.
func NewUserUsageGetter(conn *grpc.ClientConn, v2rayEmail string) *UsageGetter {
	return &UsageGetter{
		queryNames: []string{
			fmt.Sprintf("user>>>%s>>>traffic>>>downlink", v2rayEmail),
			fmt.Sprintf("user>>>%s>>>traffic>>>downlink", v2rayEmail),
		},
		key:   v2rayEmail,
		stats: command.NewStatsServiceClient(conn),
	}
}

// GetUsage returns traffic usage.
func (c *UsageGetter) GetUsage(ctx context.Context) (uint64, error) {
	var usage uint64
	for _, name := range c.queryNames {
		res, err := c.stats.GetStats(ctx, &command.GetStatsRequest{Name: name})
		if err != nil {
			return 0, fmt.Errorf("could not get user's traffic usage: %v", err)
		}
		usage += uint64(res.Stat.GetValue())
	}
	return usage, nil
}

// RequestReset requests counter reset.
func (c *UsageGetter) RequestReset(ctx context.Context) error {
	for _, name := range c.queryNames {
		_, err := c.stats.GetStats(ctx, &command.GetStatsRequest{Name: name, Reset_: true})
		if err != nil {
			return fmt.Errorf("could not reset counter: %v", err)
		}
	}
	return nil
}
