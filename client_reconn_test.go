// +build offline

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
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"go.uber.org/goleak"
)

func TestClient_Reconnect(t *testing.T) {
	var retryCount int32
	var once **sync.Once
	clients := allClients(t, &extraHandler{
		onConnHandle: func(c Client, server string, code byte, err error) bool {
			if err != nil {
				t.Log("connect to server error", err)
			}

			if code != CodeSuccess {
				t.Log("connect to server failed", code)
			}

			(*once).Do(func() {
				atomic.StoreInt32(&retryCount, 0)
				time.Sleep(7 * time.Second)
				t.Log("Destroy client")
				c.Destroy(true)
			})
			atomic.AddInt32(&retryCount, 1)
			return true
		},
	})

	onceP := unsafe.Pointer(once)
	for client, connect := range clients {
		newOnce := &sync.Once{}
		atomic.StorePointer(&onceP, unsafe.Pointer(&newOnce))
		startTime := time.Now()

		connect()
		client.Wait()

		elapsed := time.Now().Sub(startTime)
		t.Log("time used", elapsed)

		if atomic.LoadInt32(&retryCount) != 4 {
			t.Error("retryCount != 4")
		}
	}

	goleak.VerifyNoLeaks(t)
}
