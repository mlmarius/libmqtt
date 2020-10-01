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

	"github.com/stretchr/testify/assert"
)

// common test data
const (
	testUsername       = "foo"
	testPassword       = "bar"
	testClientID       = "1"
	testCleanSession   = true
	testWill           = true
	testWillQos        = Qos2
	testWillRetain     = true
	testWillTopic      = "foo"
	testKeepalive      = uint16(1)
	testProtoVersion   = V311
	testPubDup         = false
	testPacketID       = 0
	testConnAckPresent = true
	testConnAckCode    = byte(CodeSuccess)
)

var (
	_True  = true
	_False = false
	True   = &_True
	False  = &_False // nolint:unused
)

var (
	testConstUserProps      = UserProps{"MQ": []string{"TT"}}
	testConstUserPropsBytes = []byte{0, 2, 'M', 'Q', 0, 2, 'T', 'T'}
)

var (
	testWillMessage = []byte("bar")
	testSubAckCodes = []byte{SubOkMaxQos0, SubOkMaxQos1, SubFail}
	testTopics      = []string{"/test", "/test/foo", "/test/bar"}
	testTopicQos    = []QosLevel{Qos0, Qos1, Qos2}
	testTopicMsgs   = []string{"test data qos0", "foo data qos1", "bar data qos2"}
)

func testPacketBytes(version ProtoVersion, pkt Packet, target []byte, t *testing.T) {
	task := func() string {
		switch version {
		case V311:
			return "V311"
		case V5:
			return "V5"
		default:
			return "Unknown"
		}
	}()

	t.Run(task, func(t *testing.T) {
		t.Parallel()
		pkt.SetVersion(version)
		assert.Equal(t, target, pkt.Bytes())
	})
}

// nolint:unparam
func newV5TestPacketBytes(typ CtrlType, flags byte, props []byte, payload []byte) []byte {
	packetBuf := new(bytes.Buffer)
	packetBuf.WriteByte((typ << 4) | flags)

	propsLenBuf := new(bytes.Buffer)
	// get props length
	_ = writeVarInt(len(props), propsLenBuf)
	// set props
	varHeader := append(propsLenBuf.Bytes(), props...)
	// set variable header
	_ = writeVarInt(len(varHeader)+len(payload), packetBuf)

	packetBuf.Write(varHeader)
	packetBuf.Write(payload)

	return packetBuf.Bytes()
}

func init() {
	initTestData_Auth()
	initTestData_Ping()
	initTestData_Conn()
	initTestData_Sub()
	initTestData_Pub()
}
