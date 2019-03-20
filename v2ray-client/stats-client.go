package client

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"v2ray.com/core/app/stats/command"
)

// StatsClient handles stats request to v2ray api.
type StatsClient struct {
	inboundTag string
	stats      command.StatsServiceClient
}

// NewStatsClient creates a client to make stats requests.
func NewStatsClient(conn *grpc.ClientConn, inboundTag string) *StatsClient {
	return &StatsClient{
		inboundTag: inboundTag,
		stats:      command.NewStatsServiceClient(conn),
	}
}

// GetUsage returns traffic usage for user.
func (c *StatsClient) GetUsage(ctx context.Context, username string) (uint64, error) {
	names := []string{
		fmt.Sprintf("user>>>%s>>>traffic>>>downlink", username),
		fmt.Sprintf("user>>>%s>>>traffic>>>uplink", username),
	}
	var usage uint64
	for _, name := range names {
		res, err := c.stats.GetStats(ctx, &command.GetStatsRequest{Name: name})
		if err != nil {
			return 0, fmt.Errorf("could not get user's traffic usage: %v", err)
		}
		usage += uint64(res.Stat.GetValue())
	}
	return usage, nil
}
