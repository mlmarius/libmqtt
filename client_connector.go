package libmqtt

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Connector func(ctx context.Context, address string, timeout time.Duration, tlsConfig *tls.Config) (net.Conn, error)

func WithTCPConnector(handshakeTimeout time.Duration) Option {
	return func(c *AsyncClient, options *connectOptions) error {
		options.newConnection = func(ctx context.Context, address string, timeout time.Duration, tlsConfig *tls.Config) (conn net.Conn, e error) {
			c.log.v("tcpConnect()")
			return tcpConnect(ctx, address, timeout, handshakeTimeout, tlsConfig)
		}
		return nil
	}
}

func WithWebSocketConnector(handshakeTimeout time.Duration, headers http.Header) Option {
	return func(c *AsyncClient, options *connectOptions) error {
		options.newConnection = func(ctx context.Context, address string, timeout time.Duration, tlsConfig *tls.Config) (conn net.Conn, e error) {
			c.log.v("websocketConnect()")
			return websocketConnect(ctx, address, timeout, handshakeTimeout, headers, tlsConfig)
		}
		return nil
	}
}

func WithCustomConnector(connector Connector) Option {
	return func(c *AsyncClient, options *connectOptions) error {
		options.newConnection = connector
		return nil
	}
}

type tlsTimeoutError struct{}

func (tlsTimeoutError) Error() string   { return "tls: timed out" }
func (tlsTimeoutError) Timeout() bool   { return true }
func (tlsTimeoutError) Temporary() bool { return true }

func tcpConnect(ctx context.Context, address string, timeout, handshakeTimeout time.Duration, tlsConfig *tls.Config) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout: timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, err
	}

	if tlsConfig != nil {
		// code copied from golang standard library tls.DialWithDialer
		errChannel := make(chan error, 2)

		if handshakeTimeout != 0 {
			time.AfterFunc(handshakeTimeout, func() {
				errChannel <- tlsTimeoutError{}
			})
		}

		colonPos := strings.LastIndex(address, ":")
		if colonPos == -1 {
			colonPos = len(address)
		}
		hostname := address[:colonPos]

		// If no ServerName is set, infer the ServerName
		// from the hostname we're connecting to.
		if tlsConfig.ServerName == "" {
			// Make a copy to avoid polluting argument or default.
			c := tlsConfig.Clone()
			c.ServerName = hostname
			tlsConfig = c
		}

		tlsConn := tls.Client(conn, tlsConfig)
		if timeout == 0 {
			err = tlsConn.Handshake()
		} else {
			go func() {
				errChannel <- tlsConn.Handshake()
			}()

			err = <-errChannel
		}

		if err != nil {
			_ = conn.Close()
			return nil, err
		}

		conn = tlsConn
	}

	return conn, err
}

type wsConn struct {
	conn *websocket.Conn
}

func (c *wsConn) Read(b []byte) (n int, err error) {
	_, reader, err := c.conn.NextReader()
	if err != nil {
		return 0, err
	}
	return io.ReadFull(reader, b)
}

func (c *wsConn) Write(b []byte) (n int, err error) {
	w, err := c.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return 0, err
	}
	defer func() {
		if e := w.Close(); e != nil && err == nil {
			err = e
		}
	}()
	return w.Write(b)
}

func (c *wsConn) Close() error {
	return c.conn.Close()
}

func (c *wsConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *wsConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *wsConn) SetDeadline(t time.Time) error {
	return c.conn.UnderlyingConn().SetDeadline(t)
}

func (c *wsConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *wsConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func websocketConnect(ctx context.Context, address string, dialTimeout, handShakeTimeout time.Duration, headers http.Header, tlsConfig *tls.Config) (net.Conn, error) {
	netDialer := &net.Dialer{
		Timeout: dialTimeout,
	}

	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: handShakeTimeout,
		NetDial:          netDialer.Dial,
		NetDialContext:   netDialer.DialContext,
		TLSClientConfig:  tlsConfig,
		Subprotocols:     []string{"mqtt"},
	}

	urlSchema := "ws"
	if tlsConfig != nil {
		urlSchema = "wss"
	}

	conn, _, err := dialer.DialContext(ctx, urlSchema+"://"+address, headers)
	if err != nil {
		return nil, err
	}

	return &wsConn{conn: conn}, nil
}

func quicConnector(address string, timeout, handshakeTimeout time.Duration, tlsConfig *tls.Config) (net.Conn, error) {
	// TODO: TBD
	return nil, nil
}
