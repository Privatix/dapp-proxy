package v2rayclient

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"v2ray.com/core/app/proxyman/command"
	"v2ray.com/core/common/protocol"
	"v2ray.com/core/common/serial"
	"v2ray.com/core/proxy/vmess"
)

// UsersClient talks to v2ray api.
type UsersClient struct {
	alterID    uint32
	inboundTag string
	handler    command.HandlerServiceClient
}

// NewUsersClient creates a client to make user management requests.
func NewUsersClient(conn *grpc.ClientConn, inboundTag string, alterID uint32) *UsersClient {
	client := &UsersClient{
		alterID:    alterID,
		inboundTag: inboundTag,
		handler:    command.NewHandlerServiceClient(conn),
	}
	return client
}

// AddUser adds a user with id to inbound.
func (c *UsersClient) AddUser(ctx context.Context, id string) error {
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
		return fmt.Errorf("could not add user `%s`: %v", id, err)
	}
	return nil
}

// RemoveUser removes user with id from inbound.
func (c *UsersClient) RemoveUser(ctx context.Context, id string) error {
	_, err := c.handler.AlterInbound(ctx, &command.AlterInboundRequest{
		Tag: c.inboundTag,
		Operation: serial.ToTypedMessage(&command.RemoveUserOperation{
			Email: id,
		}),
	})
	if err != nil {
		return fmt.Errorf("could not remove user `%s`: %v", id, err)
	}
	return err
}
