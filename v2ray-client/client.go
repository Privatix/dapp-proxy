package client

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"v2ray.com/core/app/proxyman/command"
	"v2ray.com/core/common/protocol"
	"v2ray.com/core/common/serial"
	"v2ray.com/core/proxy/vmess"
)

// Client talks to v2ray api.
type Client struct {
	alterID    uint32
	inboundTag string
	handler    command.HandlerServiceClient
}

// Dial creates a client by establishing grpc connection with v2ray api.
func Dial(addr, inboundTag string, alterID uint32) (*Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("could not create v2ray api client: %v", err)
	}
	client := &Client{
		alterID:    alterID,
		inboundTag: inboundTag,
		handler:    command.NewHandlerServiceClient(conn),
	}
	return client, nil
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
