// +build !confonly

package dns

import (
	"context"

	"github.com/perfect-network/v2ray-core/common/net"
	"github.com/perfect-network/v2ray-core/features/dns/localdns"
)

// IPOption is an object for IP query options.
type IPOption struct {
	IPv4Enable bool
	IPv6Enable bool
}

// Client is the interface for DNS client.
type Client interface {
	// Name of the Client.
	Name() string

	// QueryIP sends IP queries to its configured server.
	QueryIP(ctx context.Context, domain string, option IPOption) ([]net.IP, error)
}

type LocalNameServer struct {
	client *localdns.Client
}

func (s *LocalNameServer) QueryIP(ctx context.Context, domain string, option IPOption) ([]net.IP, error) {
	if option.IPv4Enable && option.IPv6Enable {
		return s.client.LookupIP(domain)
	}

	if option.IPv4Enable {
		return s.client.LookupIPv4(domain)
	}

	if option.IPv6Enable {
		return s.client.LookupIPv6(domain)
	}

	return nil, newError("neither IPv4 nor IPv6 is enabled")
}

func (s *LocalNameServer) Name() string {
	return "localhost"
}

func NewLocalNameServer() *LocalNameServer {
	newError("DNS: created localhost client").AtInfo().WriteToLog()
	return &LocalNameServer{
		client: localdns.New(),
	}
}
