package client

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"v2ray.com/core/app/proxyman/command"
	statscommand "v2ray.com/core/app/stats/command"
	"v2ray.com/core/common/protocol"
	"v2ray.com/core/common/serial"
	"v2ray.com/core/proxy/vmess"
)

// Client talks to v2ray api.
type Client struct {
	alterID    uint32
	inboundTag string
	handler    command.HandlerServiceClient
	stats      statscommand.StatsServiceClient
}

// NewClient creates a client by establishing grpc connection with v2ray api.
func NewClient(conn *grpc.ClientConn, inboundTag string, alterID uint32) *Client {
	client := &Client{
		alterID:    alterID,
		inboundTag: inboundTag,
		handler:    command.NewHandlerServiceClient(conn),
		stats:      statscommand.NewStatsServiceClient(conn),
	}
	return client
}

// AddUser adds a user with id to inbound.
func (c *Client) AddUser(ctx context.Context, id string) error {
	_, err := c.handler.AlterInbound(ctx, &command.AlterInboundRequest{
		Tag: c.inboundTag,
		Operation: serial.ToTypedMessage(&command.AddUserOperation{
			User: &protocol.User{
				Email: id,
				Account: serial.ToTypedMessage(&vmess.Account{
					Id:      id,
					AlterId: c.alterID,
				}),
			},
		}),
	})
	if err != nil {
		return fmt.Errorf("could not add user: %v", err)
	}
	return nil
}

// RemoveUser removes user with id from inbound.
func (c *Client) RemoveUser(ctx context.Context, id string) error {
	_, err := c.handler.AlterInbound(ctx, &command.AlterInboundRequest{
		Tag: c.inboundTag,
		Operation: serial.ToTypedMessage(&command.RemoveUserOperation{
			Email: id,
		}),
	})
	if err != nil {
		return fmt.Errorf("could not remove user: %v", err)
	}
	return err
}

// GetUsage returns traffic usage for user.
func (c *Client) GetUsage(ctx context.Context, username string) (uint64, error) {
	names := []string{
		fmt.Sprintf("user>>>%s>>>traffic>>>downlink", username),
		fmt.Sprintf("user>>>%s>>>traffic>>>uplink", username),
	}
	var usage uint64
	for _, name := range names {
		res, err := c.stats.GetStats(ctx, &statscommand.GetStatsRequest{Name: name})
		if err != nil {
			return 0, fmt.Errorf("could not get user's traffic usage: %v", err)
		}
		usage += uint64(res.Stat.GetValue())
	}
	return usage, nil
}
