package transport

import (
	"github.com/perfect-network/v2ray-core/transport/internet"
)

// Apply applies this Config.
func (c *Config) Apply() error {
	if c == nil {
		return nil
	}
	return internet.ApplyGlobalTransportSettings(c.TransportSettings)
}
