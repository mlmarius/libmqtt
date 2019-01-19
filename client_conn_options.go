package libmqtt

import (
	"bufio"
	"context"
	"crypto/tls"
	"math"
	"net"
	"time"
)

// ConnectServer connect to server with connection specific options
// only return errors happened when applying options
func (c *AsyncClient) ConnectServer(server string, connOptions ...Option) error {
	c.log.v("AsyncClient.ConnectServer()")
	options := c.options.clone()

	for _, setOption := range connOptions {
		if err := setOption(c, &options); err != nil {
			return err
		}
	}

	c.workers.Add(1)
	go options.connect(c, server, options.protoVersion, options.firstDelay)

	return nil
}

type ConnOption func(*connectOptions) error

func defaultConnectOptions() connectOptions {
	return connectOptions{
		protoVersion:    V311,
		protoCompromise: false,

		maxDelay:         2 * time.Minute,
		firstDelay:       5 * time.Second,
		backOffFactor:    1.5,
		dialTimeout:      20 * time.Second,
		defaultTlsConfig: &tls.Config{},
		keepalive:        2 * time.Minute,
		keepaliveFactor:  1.5,
		connPacket:       &ConnPacket{},

		newConnection: func(ctx context.Context, address string, timeout time.Duration, tlsConfig *tls.Config) (conn net.Conn, e error) {
			return tcpConnect(ctx, address, timeout, 0, tlsConfig)
		},
	}
}

// connect options when connecting server (for conn packet)
type connectOptions struct {
	connHandler     ConnHandler
	dialTimeout     time.Duration
	protoVersion    ProtoVersion
	protoCompromise bool

	tlsConfig        *tls.Config // tls config with client side cert
	maxDelay         time.Duration
	firstDelay       time.Duration
	backOffFactor    float64
	autoReconnect    bool
	defaultTlsConfig *tls.Config

	connPacket      *ConnPacket
	keepalive       time.Duration // used by ConnPacket (time in second)
	keepaliveFactor float64       // used for reasonable amount time to close conn if no ping resp

	newConnection Connector
}

func (c connectOptions) connect(
	parent *AsyncClient,
	server string,
	version ProtoVersion,
	reconnectDelay time.Duration,
) {
	parent.log.v("connectOptions.connect()")
	var (
		conn net.Conn
		err  error
	)

	ctx, workers, log := parent.ctx, parent.workers, parent.log
	defer workers.Done()

	conn, err = c.newConnection(ctx, server, c.dialTimeout, c.tlsConfig)
	if err != nil {
		log.e("CLI connect server failed, err =", err, ", server =", server)
		if c.connHandler != nil {
			go c.connHandler(server, math.MaxUint8, err)
		}

		if c.autoReconnect && !parent.isClosing() {
			goto reconnect
		}

		return
	}

	defer func() { _ = conn.Close() }()

	{
		if parent.isClosing() {
			return
		}

		connImpl := &clientConn{
			protoVersion: version,
			parent:       parent,
			name:         server,
			conn:         conn,
			connRW:       bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
			keepaliveC:   make(chan struct{}, 1),
			logicSendC:   make(chan Packet, 10),
			netRecvC:     make(chan Packet, 10),
		}

		parent.connectedServers.Store(server, connImpl)

		connImpl.ctx, connImpl.exit = context.WithCancel(ctx)

		workers.Add(2)
		go connImpl.handleSend()
		go connImpl.handleNetRecv()

		connPkt := c.connPacket.clone()
		connPkt.ProtoVersion = version
		connImpl.send(connPkt)

		select {
		case <-ctx.Done():
			return
		case pkt, more := <-connImpl.netRecvC:
			if !more {
				if c.connHandler != nil {
					go c.connHandler(server, math.MaxUint8, ErrDecodeBadPacket)
				}
				close(connImpl.logicSendC)
				return
			}

			switch pkt.(type) {
			case *ConnAckPacket:
				p := pkt.(*ConnAckPacket)

				if p.Code != CodeSuccess {
					close(connImpl.logicSendC)

					if version > V311 && c.protoCompromise && p.Code == CodeUnsupportedProtoVersion {
						if !connImpl.parentExiting() {
							workers.Add(1)
							go c.connect(parent, server, version-1, reconnectDelay)
						}
						return
					}

					if c.connHandler != nil {
						go c.connHandler(server, p.Code, nil)
					}
					return
				}
			default:
				close(connImpl.logicSendC)
				if c.connHandler != nil {
					go c.connHandler(server, math.MaxUint8, ErrDecodeBadPacket)
				}
				return
			}
		}

		log.i("CLI connected to server =", server)
		if c.connHandler != nil {
			go c.connHandler(server, CodeSuccess, nil)
		}

		// start mqtt logic
		connImpl.logic()

		if parent.isClosing() || connImpl.parentExiting() {
			return
		}
	}

reconnect:

	reconnectTimer := time.NewTimer(reconnectDelay)
	defer reconnectTimer.Stop()

	select {
	case <-reconnectTimer.C:
		log.e("CLI reconnecting to server =", server, "delay =", reconnectDelay)
		reconnectDelay = time.Duration(float64(reconnectDelay) * c.backOffFactor)
		if reconnectDelay > c.maxDelay {
			reconnectDelay = c.maxDelay
		}

		workers.Add(1)
		go c.connect(parent, server, version-1, reconnectDelay)
	case <-ctx.Done():
		return
	}
}

func (c connectOptions) clone() connectOptions {
	return connectOptions{
		connHandler:      c.connHandler,
		dialTimeout:      c.dialTimeout,
		protoVersion:     c.protoVersion,
		protoCompromise:  c.protoCompromise,
		tlsConfig:        c.tlsConfig,
		maxDelay:         c.maxDelay,
		firstDelay:       c.firstDelay,
		backOffFactor:    c.backOffFactor,
		autoReconnect:    c.autoReconnect,
		defaultTlsConfig: c.defaultTlsConfig,
		connPacket:       c.connPacket,
		keepalive:        c.keepalive,
		keepaliveFactor:  c.keepaliveFactor,
		newConnection:    c.newConnection,
	}
}
