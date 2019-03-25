package v2rayclient

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"v2ray.com/core"
	"v2ray.com/core/app/proxyman/command"
	"v2ray.com/core/common/net"
	"v2ray.com/core/common/protocol"
	"v2ray.com/core/common/serial"
	"v2ray.com/core/proxy/vmess"
	"v2ray.com/core/proxy/vmess/outbound"
)

const outboundVmessTag = "vmess-outbound"

// Configurer configures outbounds on client.
type Configurer struct {
	handler command.HandlerServiceClient
}

// NewConfigurer creates new configurer.
func NewConfigurer(conn *grpc.ClientConn) *Configurer {
	return &Configurer{
		handler: command.NewHandlerServiceClient(conn),
	}
}

// AddVmessRequest params to configure vmess outbound.
type AddVmessRequest struct {
	Address string
	AlterID uint32
	ID      string
	Port    uint32
}

// AddVmess configures vmess outbound.
func (c *Configurer) AddVmess(ctx context.Context, req *AddVmessRequest) error {
	addr, err := parseAddr(req.Address)
	if err != nil {
		return fmt.Errorf("could not parse ip address %s: %v", req.Address, err)
	}
	_, err = c.handler.AddOutbound(ctx, &command.AddOutboundRequest{
		Outbound: &core.OutboundHandlerConfig{
			Tag: outboundVmessTag,
			ProxySettings: serial.ToTypedMessage(&outbound.Config{
				Receiver: []*protocol.ServerEndpoint{
					{
						Address: net.NewIPOrDomain(addr),
						Port:    req.Port,
						User: []*protocol.User{
							{
								Account: serial.ToTypedMessage(&vmess.Account{
									Id:      req.ID,
									AlterId: req.AlterID,
								}),
							},
						},
					},
				},
			}),
		},
	})
	return err
}

// RemoveVmess removes vmess outbound.
func (c *Configurer) RemoveVmess(ctx context.Context) error {
	_, err := c.handler.RemoveOutbound(ctx, &command.RemoveOutboundRequest{
		Tag: outboundVmessTag,
	})
	return err
}

func parseAddr(addr string) (net.Address, error) {
	parts := strings.Split(addr, ".")
	err := errors.New("invalid ipv4")
	if len(parts) != 4 {
		return nil, err
	}
	ip := make([]byte, 4)
	for i, part := range parts {
		v, err := strconv.ParseUint(part, 10, 8)
		if err != nil || v >= 1<<8 {
			return nil, err
		}
		ip[i] = byte(v)
	}
	return net.IPAddress(ip), nil
}
