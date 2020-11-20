// +build !confonly

package socks

import (
	"context"
	"time"

	"github.com/perfect-network/v2ray-core"
	"github.com/perfect-network/v2ray-core/common"
	"github.com/perfect-network/v2ray-core/common/buf"
	"github.com/perfect-network/v2ray-core/common/net"
	"github.com/perfect-network/v2ray-core/common/protocol"
	"github.com/perfect-network/v2ray-core/common/retry"
	"github.com/perfect-network/v2ray-core/common/session"
	"github.com/perfect-network/v2ray-core/common/signal"
	"github.com/perfect-network/v2ray-core/common/task"
	"github.com/perfect-network/v2ray-core/features/policy"
	"github.com/perfect-network/v2ray-core/transport"
	"github.com/perfect-network/v2ray-core/transport/internet"
)

// Client is a Socks5 client.
type Client struct {
	serverPicker  protocol.ServerPicker
	policyManager policy.Manager
}

// NewClient create a new Socks5 client based on the given config.
func NewClient(ctx context.Context, config *ClientConfig) (*Client, error) {
	serverList := protocol.NewServerList()
	for _, rec := range config.Server {
		s, err := protocol.NewServerSpecFromPB(rec)
		if err != nil {
			return nil, newError("failed to get server spec").Base(err)
		}
		serverList.AddServer(s)
	}
	if serverList.Size() == 0 {
		return nil, newError("0 target server")
	}

	v := core.MustFromContext(ctx)
	return &Client{
		serverPicker:  protocol.NewRoundRobinServerPicker(serverList),
		policyManager: v.GetFeature(policy.ManagerType()).(policy.Manager),
	}, nil
}

// Process implements proxy.Outbound.Process.
func (c *Client) Process(ctx context.Context, link *transport.Link, dialer internet.Dialer) error {
	outbound := session.OutboundFromContext(ctx)
	if outbound == nil || !outbound.Target.IsValid() {
		return newError("target not specified.")
	}
	destination := outbound.Target

	var server *protocol.ServerSpec
	var conn internet.Connection

	if err := retry.ExponentialBackoff(5, 100).On(func() error {
		server = c.serverPicker.PickServer()
		dest := server.Destination()
		rawConn, err := dialer.Dial(ctx, dest)
		if err != nil {
			return err
		}
		conn = rawConn

		return nil
	}); err != nil {
		return newError("failed to find an available destination").Base(err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			newError("failed to closed connection").Base(err).WriteToLog(session.ExportIDToError(ctx))
		}
	}()

	p := c.policyManager.ForLevel(0)

	request := &protocol.RequestHeader{
		Version: socks5Version,
		Command: protocol.RequestCommandTCP,
		Address: destination.Address,
		Port:    destination.Port,
	}
	if destination.Network == net.Network_UDP {
		request.Command = protocol.RequestCommandUDP
	}

	user := server.PickUser()
	if user != nil {
		request.User = user
		p = c.policyManager.ForLevel(user.Level)
	}

	if err := conn.SetDeadline(time.Now().Add(p.Timeouts.Handshake)); err != nil {
		newError("failed to set deadline for handshake").Base(err).WriteToLog(session.ExportIDToError(ctx))
	}
	udpRequest, err := ClientHandshake(request, conn, conn)
	if err != nil {
		return newError("failed to establish connection to server").AtWarning().Base(err)
	}

	if err := conn.SetDeadline(time.Time{}); err != nil {
		newError("failed to clear deadline after handshake").Base(err).WriteToLog(session.ExportIDToError(ctx))
	}

	ctx, cancel := context.WithCancel(ctx)
	timer := signal.CancelAfterInactivity(ctx, cancel, p.Timeouts.ConnectionIdle)

	var requestFunc func() error
	var responseFunc func() error
	if request.Command == protocol.RequestCommandTCP {
		requestFunc = func() error {
			defer timer.SetTimeout(p.Timeouts.DownlinkOnly)
			return buf.Copy(link.Reader, buf.NewWriter(conn), buf.UpdateActivity(timer))
		}
		responseFunc = func() error {
			defer timer.SetTimeout(p.Timeouts.UplinkOnly)
			return buf.Copy(buf.NewReader(conn), link.Writer, buf.UpdateActivity(timer))
		}
	} else if request.Command == protocol.RequestCommandUDP {
		udpConn, err := dialer.Dial(ctx, udpRequest.Destination())
		if err != nil {
			return newError("failed to create UDP connection").Base(err)
		}
		defer udpConn.Close()
		requestFunc = func() error {
			defer timer.SetTimeout(p.Timeouts.DownlinkOnly)
			return buf.Copy(link.Reader, &buf.SequentialWriter{Writer: NewUDPWriter(request, udpConn)}, buf.UpdateActivity(timer))
		}
		responseFunc = func() error {
			defer timer.SetTimeout(p.Timeouts.UplinkOnly)
			reader := &UDPReader{reader: udpConn}
			return buf.Copy(reader, link.Writer, buf.UpdateActivity(timer))
		}
	}

	var responseDonePost = task.OnSuccess(responseFunc, task.Close(link.Writer))
	if err := task.Run(ctx, requestFunc, responseDonePost); err != nil {
		return newError("connection ends").Base(err)
	}

	return nil
}

func init() {
	common.Must(common.RegisterConfig((*ClientConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewClient(ctx, config.(*ClientConfig))
	}))
}
