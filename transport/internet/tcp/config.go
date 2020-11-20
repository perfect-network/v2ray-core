// +build !confonly

package tcp

import (
	"github.com/perfect-network/v2ray-core/common"
	"github.com/perfect-network/v2ray-core/transport/internet"
)

const protocolName = "tcp"

func init() {
	common.Must(internet.RegisterProtocolConfigCreator(protocolName, func() interface{} {
		return new(Config)
	}))
}
