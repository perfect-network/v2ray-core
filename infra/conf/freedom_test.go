package conf_test

import (
	"testing"

	"github.com/perfect-network/v2ray-core/common/net"
	"github.com/perfect-network/v2ray-core/common/protocol"
	. "github.com/perfect-network/v2ray-core/infra/conf"
	"github.com/perfect-network/v2ray-core/proxy/freedom"
)

func TestFreedomConfig(t *testing.T) {
	creator := func() Buildable {
		return new(FreedomConfig)
	}

	runMultiTestCase(t, []TestCase{
		{
			Input: `{
				"domainStrategy": "AsIs",
				"timeout": 10,
				"redirect": "127.0.0.1:3366",
				"userLevel": 1
			}`,
			Parser: loadJSON(creator),
			Output: &freedom.Config{
				DomainStrategy: freedom.Config_AS_IS,
				Timeout:        10,
				DestinationOverride: &freedom.DestinationOverride{
					Server: &protocol.ServerEndpoint{
						Address: &net.IPOrDomain{
							Address: &net.IPOrDomain_Ip{
								Ip: []byte{127, 0, 0, 1},
							},
						},
						Port: 3366,
					},
				},
				UserLevel: 1,
			},
		},
	})
}
