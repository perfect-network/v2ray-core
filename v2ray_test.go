package core_test

import (
	"testing"

	"github.com/golang/protobuf/proto"
	. "github.com/perfect-network/v2ray-core"
	"github.com/perfect-network/v2ray-core/app/dispatcher"
	"github.com/perfect-network/v2ray-core/app/proxyman"
	"github.com/perfect-network/v2ray-core/common"
	"github.com/perfect-network/v2ray-core/common/net"
	"github.com/perfect-network/v2ray-core/common/protocol"
	"github.com/perfect-network/v2ray-core/common/serial"
	"github.com/perfect-network/v2ray-core/common/uuid"
	"github.com/perfect-network/v2ray-core/features/dns"
	"github.com/perfect-network/v2ray-core/features/dns/localdns"
	_ "github.com/perfect-network/v2ray-core/main/distro/all"
	"github.com/perfect-network/v2ray-core/proxy/dokodemo"
	"github.com/perfect-network/v2ray-core/proxy/vmess"
	"github.com/perfect-network/v2ray-core/proxy/vmess/outbound"
	"github.com/perfect-network/v2ray-core/testing/servers/tcp"
)

func TestV2RayDependency(t *testing.T) {
	instance := new(Instance)

	wait := make(chan bool, 1)
	instance.RequireFeatures(func(d dns.Client) {
		if d == nil {
			t.Error("expected dns client fulfilled, but actually nil")
		}
		wait <- true
	})
	instance.AddFeature(localdns.New())
	<-wait
}

func TestV2RayClose(t *testing.T) {
	port := tcp.PickPort()

	userID := uuid.New()
	config := &Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&dispatcher.Config{}),
			serial.ToTypedMessage(&proxyman.InboundConfig{}),
			serial.ToTypedMessage(&proxyman.OutboundConfig{}),
		},
		Inbound: []*InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(port),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address: net.NewIPOrDomain(net.LocalHostIP),
					Port:    uint32(0),
					NetworkList: &net.NetworkList{
						Network: []net.Network{net.Network_TCP, net.Network_UDP},
					},
				}),
			},
		},
		Outbound: []*OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&outbound.Config{
					Receiver: []*protocol.ServerEndpoint{
						{
							Address: net.NewIPOrDomain(net.LocalHostIP),
							Port:    uint32(0),
							User: []*protocol.User{
								{
									Account: serial.ToTypedMessage(&vmess.Account{
										Id: userID.String(),
									}),
								},
							},
						},
					},
				}),
			},
		},
	}

	cfgBytes, err := proto.Marshal(config)
	common.Must(err)

	server, err := StartInstance("protobuf", cfgBytes)
	common.Must(err)
	server.Close()
}
