package udp

import (
	"github.com/perfect-network/v2ray-core/common/buf"
	"github.com/perfect-network/v2ray-core/common/net"
)

// Packet is a UDP packet together with its source and destination address.
type Packet struct {
	Payload *buf.Buffer
	Source  net.Destination
	Target  net.Destination
}
