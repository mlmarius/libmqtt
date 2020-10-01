/*
 * Copyright Go-IIoT (https://github.com/goiiot)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
		// WithVersion(V5, true),
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
		WithConnHandleFunc(testConnHandler(&c, t, handler)),
		WithPubHandleFunc(func(client Client, topic string, err error) {
			if err != nil {
				t.Error(err)
				t.FailNow()
			}

			if handler != nil && handler.afterPubSuccess != nil {
				println("afterPubSuccess()")
				handler.afterPubSuccess(c)
			}
		}),
		WithSubHandleFunc(func(client Client, topics []*Topic, err error) {
			println("exH.Sub")
			if err != nil {
				t.Error(err)
				t.FailNow()
			}

			if handler != nil && handler.afterSubSuccess != nil {
				println("afterSubSuccess()")
				handler.afterSubSuccess(c)
			}
		}),
		WithUnsubHandleFunc(func(client Client, topics []string, err error) {
			println("exH.UnSub")
			if err != nil {
				t.Error(err)
				t.FailNow()
			}

			if handler != nil && handler.afterUnSubSuccess != nil {
				println("afterUnSubSuccess()")
				handler.afterUnSubSuccess(c)
			}
		}),
		WithNetHandleFunc(func(client Client, server string, err error) {
			// if err != nil {
			// 	// t.Error(err)
			// 	// t.FailNow()
			// }
		}),
		WithPersistHandleFunc(func(client Client, packet Packet, err error) {}),
	)
	if err != nil {
		panic("create baseClient failed")
	}

	return c
}

func testConnHandler(c *Client, t *testing.T, exH *extraHandler) ConnHandleFunc {
	return func(client Client, server string, code byte, err error) {
		if exH != nil && exH.onConnHandle != nil {
			println("onConnHandle()")
			if exH.onConnHandle(*c, server, code, err) {
				return
			}
		}

		if err != nil {
			t.Errorf("connect errored: %v", err)
			t.FailNow()
		}

		if code != CodeSuccess {
			t.Errorf("connect failed with code: %d", code)
			t.FailNow()
		}

		if exH != nil && exH.afterConnSuccess != nil {
			println("afterConnSuccess()")
			exH.afterConnSuccess(*c)
		}
	}
}

func websocketPlainClient(t *testing.T, exH *extraHandler) (c Client, connFunc func()) {
	c = baseClient(t, exH)
	connFunc = func() {
		if err := c.ConnectServer("localhost:8083/mqtt",
			WithWebSocketConnector(0, nil),
		); err != nil {
			t.Error(err)
			t.FailNow()
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
			t.FailNow()
		}
	}

	return
}

func tcpPlainClient(t *testing.T, exH *extraHandler) (c Client, connFunc func()) {
	c = baseClient(t, exH)
	connFunc = func() {
		if err := c.ConnectServer("localhost:1883"); err != nil {
			t.Error(err)
			t.FailNow()
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
			t.FailNow()
		}
	}
	return
}

func allClients(t *testing.T, handler *extraHandler) map[Client]func() {
	ret := make(map[Client]func())
	// tcp based
	tcp, tcpC := tcpPlainClient(t, handler)
	ret[tcp] = tcpC
	tcps, tcpsC := tcpTLSClient(t, handler)
	ret[tcps] = tcpsC

	// websocket based
	ws, wsC := websocketPlainClient(t, handler)
	ret[ws] = wsC
	wss, wssC := websocketTLSClient(t, handler)
	ret[wss] = wssC

	return ret
}

func handleTopicAndSub(c Client, t *testing.T) {
	for i := range testTopics {
		c.HandleTopic(testTopics[i], func(client Client, topic string, maxQos byte, msg []byte) {
			if maxQos != testPubMsgs[i].Qos || !bytes.Equal(testPubMsgs[i].Payload, msg) {
				t.Errorf("fail at sub topic = %v, content unexpected, payload = %v, target payload = %v",
					topic, string(msg), string(testPubMsgs[i].Payload))
				t.FailNow()
			}
		})
	}

	c.Subscribe(testSubTopics...)
}
