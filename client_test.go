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
	"testing"

	"go.uber.org/goleak"
)

// test with emqx server (http://emqtt.io/ or https://github.com/emqx/emqx)
// the server is configured with default configuration

func TestNewClient(t *testing.T) {
	_, err := NewClient(
		WithTLS("foo", "bar", "foobar", "foo.bar", true),
	)
	if err == nil {
		t.Errorf("create new client with invalid tls should fail")
	}

	goleak.VerifyNoLeaks(t)
}

// conn
func TestClient_Connect(t *testing.T) {
	var c Client
	afterConn := func() {
		c.Destroy(true)
	}

	c = websocketPlainClient(t, nil)
	c.Wait()

	c = websocketTLSClient(t, nil)
	c.Wait()

	c = tcpPlainClient(t, nil)
	conn(c, t, afterConn)
	c.Wait()

	c = tcpTLSClient(t, nil)
	conn(c, t, afterConn)
	c.Wait()

	goleak.VerifyNoLeaks(t)
}

// conn -> pub
func TestClient_Publish(t *testing.T) {
	var c Client
	afterConn := func() {
		c.Publish(testPubMsgs...)
	}

	exH := &extraHandler{
		afterPubSuccess: func() {
			c.Destroy(true)
		},
	}

	c = websocketPlainClient(t, exH)
	c.Wait()

	c = websocketTLSClient(t, exH)
	c.Wait()

	c = tcpPlainClient(t, exH)
	conn(c, t, afterConn)
	c.Wait()

	c = tcpTLSClient(t, exH)
	conn(c, t, afterConn)
	c.Wait()

	goleak.VerifyNoLeaks(t)
}

// conn -> sub -> pub
func TestClient_Subscribe(t *testing.T) {
	var c Client
	afterConn := func() {
		handleTopicAndSub(c, t)
	}

	extH := &extraHandler{
		afterSubSuccess: func() {
			c.Publish(testPubMsgs...)
		},
		afterPubSuccess: func() {
			c.Destroy(true)
		},
	}

	c = websocketPlainClient(t, extH)
	c.Wait()

	c = websocketTLSClient(t, extH)
	c.Wait()

	c = tcpPlainClient(t, extH)
	conn(c, t, afterConn)
	c.Wait()

	c = tcpTLSClient(t, extH)
	conn(c, t, afterConn)
	c.Wait()

	goleak.VerifyNoLeaks(t)
}

// conn -> sub -> pub -> unSub
func TestClient_UnSubscribe(t *testing.T) {
	var c Client
	afterConn := func() {
		handleTopicAndSub(c, t)
	}

	extH := &extraHandler{
		afterSubSuccess: func() {
			c.UnSubscribe(testTopics...)
		},
		afterUnSubSuccess: func() {
			c.Destroy(true)
		},
	}

	c = websocketPlainClient(t, extH)
	c.Wait()

	c = websocketTLSClient(t, extH)
	c.Wait()

	c = tcpPlainClient(t, extH)
	conn(c, t, afterConn)
	c.Wait()

	c = tcpTLSClient(t, extH)
	conn(c, t, afterConn)
	c.Wait()

	goleak.VerifyNoLeaks(t)
}
