package libmqtt

import (
	"bytes"
	"testing"
	"time"
)

type extraHandler struct {
	onConnHandle      func(client Client, server string, code byte, err error) bool
	afterConnSuccess  func(client Client)
	afterPubSuccess   func(client Client)
	afterSubSuccess   func(client Client)
	afterUnSubSuccess func(client Client)
}

func testTLSOption() Option {
	return WithTLS(
		"./testdata/client-cert.pem",
		"./testdata/client-key.pem",
		"./testdata/ca-cert.pem",
		"MacBook-Air.local",
		true)
}

func baseClient(t *testing.T, handler *extraHandler) Client {
	var (
		c   Client
		err error
	)

	c, err = NewClient(
		WithLog(Verbose),
		WithConnHandleFunc(testConnHandler(c, t, handler)),
		WithDialTimeout(10),
		WithKeepalive(10, 1.2),
		WithAutoReconnect(true),
		WithBackoffStrategy(1*time.Second, 5*time.Second, 1.5),
		WithConnPacket(ConnPacket{
			Username:    "admin",
			Password:    "public",
			WillTopic:   "test",
			WillQos:     Qos0,
			WillRetain:  false,
			WillMessage: []byte("test data"),
		}),
	)
	if err != nil {
		panic("create baseClient failed")
	}

	initClient(c, handler, t)
	return c
}

func testConnHandler(c Client, t *testing.T, exH *extraHandler) ConnHandler {
	return func(server string, code byte, err error) {
		if exH != nil && exH.onConnHandle != nil {
			println("onConnHandle()")
			if exH.onConnHandle(c, server, code, err) {
				return
			}
		}

		if err != nil {
			t.Errorf("connect errored: %v", err)
		}

		if code != CodeSuccess {
			t.Errorf("connect failed with code: %d", code)
		}

		if exH != nil && exH.afterConnSuccess != nil {
			println("afterConnSuccess()")
			exH.afterConnSuccess(c)
		}
	}
}

func websocketPlainClient(t *testing.T, exH *extraHandler) (c Client, connFunc func()) {
	c = baseClient(t, exH)
	connFunc = func() {
		if err := c.ConnectServer("localhost:8083",
			WithWebSocketConnector(0, nil),
		); err != nil {
			t.Error(err)
		}
	}

	return
}

func websocketTLSClient(t *testing.T, exH *extraHandler) (c Client, connFunc func()) {
	c = baseClient(t, exH)
	connFunc = func() {
		if err := c.ConnectServer("localhost:8084/mqtt",
			testTLSOption(),
			WithWebSocketConnector(0, nil),
		); err != nil {
			t.Error(err)
		}
	}

	return
}

func tcpPlainClient(t *testing.T, exH *extraHandler) (c Client, connFunc func()) {
	c = baseClient(t, exH)
	connFunc = func() {
		if err := c.ConnectServer("localhost:1883"); err != nil {
			t.Error(err)
		}
	}

	return
}

func tcpTLSClient(t *testing.T, exH *extraHandler) (c Client, connFunc func()) {
	c = baseClient(t, exH)
	connFunc = func() {
		if err := c.ConnectServer("localhost:8883",
			testTLSOption(),
		); err != nil {
			t.Error(err)
		}
	}
	return
}

func initClient(c Client, exH *extraHandler, t *testing.T) {
	c.HandlePub(func(topic string, err error) {
		if err != nil {
			t.Error(err)
		}

		if exH != nil && exH.afterPubSuccess != nil {
			println("afterPubSuccess()")
			exH.afterPubSuccess(c)
		}
	})

	c.HandleSub(func(topics []*Topic, err error) {
		println("exH.Sub")
		if err != nil {
			t.Error(err)
		}

		if exH != nil && exH.afterSubSuccess != nil {
			println("afterSubSuccess()")
			exH.afterSubSuccess(c)
		}
	})

	c.HandleUnSub(func(topics []string, err error) {
		println("exH.UnSub")
		if err != nil {
			t.Error(err)
		}

		if exH != nil && exH.afterUnSubSuccess != nil {
			println("afterUnSubSuccess()")
			exH.afterUnSubSuccess(c)
		}
	})

	c.HandleNet(func(server string, err error) {
		if err != nil {
			t.Error(err)
		}
	})
}

func allClients(t *testing.T, handler *extraHandler) map[Client]func() {
	ret := make(map[Client]func())
	ws, wsC := websocketPlainClient(t, handler)
	ret[ws] = wsC
	wss, wssC := websocketTLSClient(t, handler)
	ret[wss] = wssC
	tcp, tcpC := tcpPlainClient(t, handler)
	ret[tcp] = tcpC
	tcps, tcpsC := tcpTLSClient(t, handler)
	ret[tcps] = tcpsC

	return ret
}

func handleTopicAndSub(c Client, t *testing.T) {
	for i := range testTopics {
		c.Handle(testTopics[i], func(topic string, maxQos byte, msg []byte) {
			if maxQos != testPubMsgs[i].Qos || bytes.Compare(testPubMsgs[i].Payload, msg) != 0 {
				t.Errorf("fail at sub topic = %v, content unexpected, payload = %v, target payload = %v",
					topic, string(msg), string(testPubMsgs[i].Payload))
			}
		})
	}

	c.Subscribe(testSubTopics...)
}
