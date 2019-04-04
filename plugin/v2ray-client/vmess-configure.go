package v2rayclient

import (
	"context"

	"google.golang.org/grpc"
	"v2ray.com/core"
	"v2ray.com/core/app/proxyman/command"
	"v2ray.com/core/common/net"
	"v2ray.com/core/common/protocol"
	"v2ray.com/core/common/serial"
	"v2ray.com/core/proxy/freedom"
	"v2ray.com/core/proxy/vmess"
	"v2ray.com/core/proxy/vmess/outbound"
)

const (
	outboundVmessTag   = "outbound-vmess"
	outboundDefaultTag = "outbound-default"
)

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

// VmessOutbound params to configure vmess outbound.
type VmessOutbound struct {
	Address string
	AlterID uint32
	ID      string
	Port    uint32
}

// ConfigureVmess configures vmess outbound.
func (c *Configurer) ConfigureVmess(ctx context.Context, req *VmessOutbound) error {
	err := c.addVmessOutbound(ctx, req)
	if err != nil {
		return err
	}
	return c.removeOutbound(ctx, outboundDefaultTag)
}

// RemoveVmess removes vmess outbound.
func (c *Configurer) RemoveVmess(ctx context.Context) error {
	err := c.addDefaultOutbound(ctx)
	if err != nil {
		return err
	}
	return c.removeOutbound(ctx, outboundVmessTag)
}

func (c *Configurer) addVmessOutbound(ctx context.Context, req *VmessOutbound) error {
	_, err := c.handler.AddOutbound(ctx, &command.AddOutboundRequest{
		Outbound: &core.OutboundHandlerConfig{
			Tag: outboundVmessTag,
			ProxySettings: serial.ToTypedMessage(&outbound.Config{
				Receiver: []*protocol.ServerEndpoint{
					{
						Address: net.NewIPOrDomain(net.ParseAddress(req.Address)),
						Port:    req.Port,
						User: []*protocol.User{
							{
								Email: req.ID,
								Level: 255,
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

func (c *Configurer) addDefaultOutbound(ctx context.Context) error {
	_, err := c.handler.AddOutbound(ctx, &command.AddOutboundRequest{
		Outbound: &core.OutboundHandlerConfig{
			Tag:           outboundDefaultTag,
			ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
		},
	})
	return err
}

func (c *Configurer) removeOutbound(ctx context.Context, tag string) error {
	_, err := c.handler.RemoveOutbound(ctx, &command.RemoveOutboundRequest{
		Tag: tag,
	})
	return err
}
