package libmqtt

import (
	"bytes"
	"testing"
	"time"
)

type extraHandler struct {
	afterPubSuccess   func()
	afterSubSuccess   func()
	afterUnSubSuccess func()
}

func websocketPlainClient(t *testing.T, exH *extraHandler) Client {
	c, err := NewClient(WithLog(Verbose))
	if err != nil {
		t.Error("create client failed")
	}

	if err != nil {
		t.Error(err)
	}
	initClient(c, exH, t)

	err = c.ConnectServer("localhost:8083/mqtt",
		func(server string, code byte, err error) {
			if err != nil {
				t.Errorf("connect errored: %v", err)
			}

			if code != CodeSuccess {
				t.Errorf("connect failed with code: %d", code)
			}

			c.Destroy(true)
		},
		WithWebSocketConnector(0, nil),
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
		t.Error(err)
	}

	return c
}

func websocketTLSClient(t *testing.T, exH *extraHandler) Client {
	c, err := NewClient(WithLog(Verbose))
	if err != nil {
		t.Error("create client failed")
	}

	if err != nil {
		t.Error(err)
	}
	initClient(c, exH, t)

	err = c.ConnectServer("localhost:8084/mqtt",
		func(server string, code byte, err error) {
			if err != nil {
				t.Errorf("connect errored: %v", err)
			}

			if code != CodeSuccess {
				t.Errorf("connect failed with code: %d", code)
			}

			c.Destroy(true)
		},
		WithTLS(
			"./testdata/client-cert.pem",
			"./testdata/client-key.pem",
			"./testdata/ca-cert.pem",
			"MacBook-Air.local",
			true),
		WithWebSocketConnector(0, nil),
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
		t.Error(err)
	}

	return c
}

func tcpPlainClient(t *testing.T, exH *extraHandler) Client {
	c, err := NewClient(
		WithLog(Verbose),
		WithServer("localhost:1883"),
		WithDialTimeout(10),
		WithKeepalive(10, 1.2),
		WithIdentity("admin", "public"),
		WithWill("test", Qos0, false, []byte("test data")),
		WithAutoReconnect(true),
		WithBackoffStrategy(1*time.Second, 5*time.Second, 1.5),
	)

	if err != nil {
		t.Error(err)
	}
	initClient(c, exH, t)
	return c
}

func tcpTLSClient(t *testing.T, exH *extraHandler) Client {
	c, err := NewClient(
		WithLog(Verbose),
		WithServer("localhost:8883"),
		WithTLS(
			"./testdata/client-cert.pem",
			"./testdata/client-key.pem",
			"./testdata/ca-cert.pem",
			"MacBook-Air.local",
			true),
		WithDialTimeout(10),
		WithKeepalive(10, 1.2),
		WithIdentity("admin", "public"),
		WithWill("test", Qos0, false, []byte("test data")),
		WithAutoReconnect(true),
		WithBackoffStrategy(1*time.Second, 5*time.Second, 1.5),
	)

	if err != nil {
		t.Error(err)
	}
	initClient(c, exH, t)
	return c
}

func initClient(c Client, exH *extraHandler, t *testing.T) {
	c.HandlePub(func(topic string, err error) {
		println("exH.Pub")
		if err != nil {
			t.Error(err)
		}

		if exH != nil && exH.afterPubSuccess != nil {
			println("afterPubSuccess()")
			exH.afterPubSuccess()
		}
	})

	c.HandleSub(func(topics []*Topic, err error) {
		println("exH.Sub")
		if err != nil {
			t.Error(err)
		}

		if exH != nil && exH.afterSubSuccess != nil {
			println("afterSubSuccess()")
			exH.afterSubSuccess()
		}
	})

	c.HandleUnSub(func(topics []string, err error) {
		println("exH.UnSub")
		if err != nil {
			t.Error(err)
		}

		if exH != nil && exH.afterUnSubSuccess != nil {
			println("afterUnSubSuccess()")
			exH.afterUnSubSuccess()
		}
	})

	c.HandleNet(func(server string, err error) {
		if err != nil {
			t.Error(err)
		}
	})
}

func conn(c Client, t *testing.T, afterConnSuccess func()) {
	c.Connect(func(server string, code byte, err error) {
		if err != nil {
			t.Errorf("connect errored: %v", err)
		}

		if code != CodeSuccess {
			t.Errorf("connect failed with code: %d", code)
		}

		if afterConnSuccess != nil {
			println("afterConnSuccess()")
			afterConnSuccess()
		}
	})
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
