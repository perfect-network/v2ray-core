package outbound_test

import (
	"context"
	"testing"

	"github.com/perfect-network/v2ray-core"
	"github.com/perfect-network/v2ray-core/app/policy"
	. "github.com/perfect-network/v2ray-core/app/proxyman/outbound"
	"github.com/perfect-network/v2ray-core/app/stats"
	"github.com/perfect-network/v2ray-core/common/net"
	"github.com/perfect-network/v2ray-core/common/serial"
	"github.com/perfect-network/v2ray-core/features/outbound"
	"github.com/perfect-network/v2ray-core/proxy/freedom"
	"github.com/perfect-network/v2ray-core/transport/internet"
)

func TestInterfaces(t *testing.T) {
	_ = (outbound.Handler)(new(Handler))
	_ = (outbound.Manager)(new(Manager))
}

const v2rayKey core.V2rayKey = 1

func TestOutboundWithoutStatCounter(t *testing.T) {
	config := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&stats.Config{}),
			serial.ToTypedMessage(&policy.Config{
				System: &policy.SystemPolicy{
					Stats: &policy.SystemPolicy_Stats{
						InboundUplink: true,
					},
				},
			}),
		},
	}

	v, _ := core.New(config)
	v.AddFeature((outbound.Manager)(new(Manager)))
	ctx := context.WithValue(context.Background(), v2rayKey, v)
	h, _ := NewHandler(ctx, &core.OutboundHandlerConfig{
		Tag:           "tag",
		ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
	})
	conn, _ := h.(*Handler).Dial(ctx, net.TCPDestination(net.DomainAddress("localhost"), 13146))
	_, ok := conn.(*internet.StatCouterConnection)
	if ok {
		t.Errorf("Expected conn to not be StatCouterConnection")
	}
}

func TestOutboundWithStatCounter(t *testing.T) {
	config := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&stats.Config{}),
			serial.ToTypedMessage(&policy.Config{
				System: &policy.SystemPolicy{
					Stats: &policy.SystemPolicy_Stats{
						OutboundUplink:   true,
						OutboundDownlink: true,
					},
				},
			}),
		},
	}

	v, _ := core.New(config)
	v.AddFeature((outbound.Manager)(new(Manager)))
	ctx := context.WithValue(context.Background(), v2rayKey, v)
	h, _ := NewHandler(ctx, &core.OutboundHandlerConfig{
		Tag:           "tag",
		ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
	})
	conn, _ := h.(*Handler).Dial(ctx, net.TCPDestination(net.DomainAddress("localhost"), 13146))
	_, ok := conn.(*internet.StatCouterConnection)
	if !ok {
		t.Errorf("Expected conn to be StatCouterConnection")
	}
}
