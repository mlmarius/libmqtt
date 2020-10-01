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

	goleak.VerifyNone(t)
}

func testAllClient(t *testing.T, handler *extraHandler) {
	clients := allClients(t, handler)
	for client, startTest := range clients {
		startTest()
		// time.Sleep(10 * time.Millisecond)
		client.Wait()
	}
}

// conn
func TestClient_Connect(t *testing.T) {
	testAllClient(t, &extraHandler{
		afterConnSuccess: func(client Client) {
			client.Destroy(false)
		},
	})

	goleak.VerifyNone(t)
}

// conn -> pub
func TestClient_Publish(t *testing.T) {
	testAllClient(t, &extraHandler{
		afterConnSuccess: func(c Client) {
			c.Publish(testPubMsgs...)
		},
		afterPubSuccess: func(c Client) {
			c.Destroy(true)
		},
	})

	goleak.VerifyNone(t)
}

// conn -> sub -> pub
func TestClient_Subscribe(t *testing.T) {
	testAllClient(t, &extraHandler{
		afterConnSuccess: func(c Client) {
			handleTopicAndSub(c, t)
		},
		afterSubSuccess: func(c Client) {
			c.Publish(testPubMsgs...)
		},
		afterPubSuccess: func(c Client) {
			c.Destroy(true)
		},
	})

	goleak.VerifyNone(t)
}

// conn -> sub -> pub -> unSub
func TestClient_UnSubscribe(t *testing.T) {
	testAllClient(t, &extraHandler{
		afterConnSuccess: func(c Client) {
			handleTopicAndSub(c, t)
		},
		afterSubSuccess: func(c Client) {
			c.UnSubscribe(testTopics...)
		},
		afterUnSubSuccess: func(c Client) {
			c.Destroy(true)
		},
	})

	goleak.VerifyNone(t)
}
