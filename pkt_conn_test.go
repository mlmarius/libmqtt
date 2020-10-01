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

	std "github.com/eclipse/paho.mqtt.golang/packets"
)

// conn test data
var (
	testConnProps = &ConnProps{
		SessionExpiryInterval: 100,
		MaxRecv:               100,
		MaxPacketSize:         100,
		MaxTopicAlias:         100,
		ReqRespInfo:           True,
		ReqProblemInfo:        True,
		UserProps:             testConstUserProps,
		AuthMethod:            "MQTT",
		AuthData:              []byte("MQTT"),
	}
	// nolint:deadcode,varcheck,unused
	testConnPropsBytes = []byte{
		propKeySessionExpiryInterval, 100, 100, 100, 100,
		propKeyMaxRecv, 100, 100,
		propKeyMaxPacketSize, 100, 100, 100, 100,
		propKeyMaxTopicAlias, 100, 100,
		propKeyReqRespInfo, 1,
		propKeyReqProblemInfo, 1,
		propKeyUserProps, 0, 2, 'M', 'Q', 0, 2, 'T', 'T',
	}
	testConnWillMsg = &ConnPacket{
		BasePacket:   BasePacket{ProtoVersion: testProtoVersion},
		Username:     testUsername,
		Password:     testPassword,
		ClientID:     testClientID,
		CleanSession: testCleanSession,
		IsWill:       testWill,
		WillQos:      testWillQos,
		WillRetain:   testWillRetain,
		WillTopic:    testWillTopic,
		WillMessage:  testWillMessage,
		Keepalive:    testKeepalive,
		Props:        testConnProps,
	}

	testConnMsg = &ConnPacket{
		BasePacket:   BasePacket{ProtoVersion: testProtoVersion},
		Username:     testUsername,
		Password:     testPassword,
		ClientID:     testClientID,
		CleanSession: testCleanSession,
		Keepalive:    testKeepalive,
		Props:        testConnProps,
	}

	testConnAckMsg = &ConnAckPacket{
		Present: testConnAckPresent,
		Code:    testConnAckCode,
		Props: &ConnAckProps{
			SessionExpiryInterval: 100,
			MaxRecv:               100,
			MaxQos:                Qos2,
			RetainAvail:           True,
			MaxPacketSize:         100,
			AssignedClientID:      "MQTT",
			MaxTopicAlias:         100,
			Reason:                "MQTT",
			UserProps:             testConstUserProps,
			WildcardSubAvail:      True,
			SubIDAvail:            True,
			SharedSubAvail:        True,
			ServerKeepalive:       100,
			RespInfo:              "MQTT",
			ServerRef:             "MQTT",
			AuthMethod:            "MQTT",
			AuthData:              []byte("MQTT"),
		},
	}

	testDisConnMsg = &DisconnPacket{
		BasePacket: BasePacket{ProtoVersion: testProtoVersion},
		Code:       CodeUnspecifiedError,
		Props: &DisconnProps{
			SessionExpiryInterval: 100,
			Reason:                "MQTT",
			ServerRef:             "MQTT",
			UserProps:             UserProps{"MQ": []string{"TT"}},
		},
	}
	testDisConnPropsBytes = []byte{
		propKeySessionExpiryInterval, 100, 100, 100, 100,
		propKeyReasonString, 0, 4, 'M', 'Q', 'T', 'T',
		propKeyServerRef, 0, 4, 'M', 'Q', 'T', 'T',
		propKeyUserProps, 0, 2, 'M', 'Q', 0, 2, 'T', 'T',
	}

	// mqtt 3.1.1
	testConnWillMsgBytesV311 []byte
	testConnMsgBytesV311     []byte
	testConnAckMsgBytesV311  []byte
	testDisConnMsgBytesV311  []byte

	// mqtt 5.0
	testConnWillMsgBytesV5 []byte
	testConnMsgBytesV5     []byte
	testConnAckMsgBytesV5  []byte
	testDisConnMsgBytesV5  []byte
)

func initTestData_Conn() {
	connWillPkt := std.NewControlPacket(std.Connect).(*std.ConnectPacket)
	connWillPkt.Username = testUsername
	connWillPkt.UsernameFlag = true
	connWillPkt.Password = []byte(testPassword)
	connWillPkt.PasswordFlag = true
	connWillPkt.ProtocolName = "MQTT"
	connWillPkt.ProtocolVersion = byte(testProtoVersion)
	connWillPkt.ClientIdentifier = testClientID
	connWillPkt.CleanSession = testCleanSession
	connWillPkt.WillFlag = testWill
	connWillPkt.WillQos = testWillQos
	connWillPkt.WillRetain = testWillRetain
	connWillPkt.WillTopic = testWillTopic
	connWillPkt.WillMessage = testWillMessage
	connWillPkt.Keepalive = testKeepalive
	connWillBuf := new(bytes.Buffer)
	_ = connWillPkt.Write(connWillBuf)
	testConnWillMsgBytesV311 = connWillBuf.Bytes()
	testConnWillMsgBytesV5 = newV5TestPacketBytes(CtrlConn, 0, nil, nil)

	connPkt := std.NewControlPacket(std.Connect).(*std.ConnectPacket)
	connPkt.Username = testUsername
	connPkt.UsernameFlag = true
	connPkt.Password = []byte(testPassword)
	connPkt.PasswordFlag = true
	connPkt.ProtocolName = "MQTT"
	connPkt.ProtocolVersion = byte(testProtoVersion)
	connPkt.ClientIdentifier = testClientID
	connPkt.CleanSession = testCleanSession
	connPkt.Keepalive = testKeepalive
	connBuf := new(bytes.Buffer)
	_ = connPkt.Write(connBuf)
	testConnMsgBytesV311 = connBuf.Bytes()
	testConnMsgBytesV5 = newV5TestPacketBytes(CtrlConn, 0, nil, nil)

	connAckPkt := std.NewControlPacket(std.Connack).(*std.ConnackPacket)
	connAckPkt.SessionPresent = testConnAckPresent
	connAckPkt.ReturnCode = testConnAckCode
	connAckBuf := new(bytes.Buffer)
	_ = connAckPkt.Write(connAckBuf)
	testConnAckMsgBytesV311 = connAckBuf.Bytes()
	testConnAckMsgBytesV5 = newV5TestPacketBytes(CtrlConnAck, 0, nil, nil)

	disConnPkt := std.NewControlPacket(std.Disconnect).(*std.DisconnectPacket)
	disConnBuf := new(bytes.Buffer)
	_ = disConnPkt.Write(disConnBuf)
	testDisConnMsgBytesV311 = disConnBuf.Bytes()
	testDisConnMsgBytesV5 = newV5TestPacketBytes(CtrlDisConn, 0,
		append([]byte{CodeUnspecifiedError}, testDisConnPropsBytes...), nil)
}

func TestConnPacket_Bytes(t *testing.T) {
	testPacketBytes(V311, testConnMsg, testConnMsgBytesV311, t)
	t.Skip("v5")
	testPacketBytes(V5, testConnMsg, testConnMsgBytesV5, t)
}

func TestConnWillPacket_Bytes(t *testing.T) {
	testPacketBytes(V311, testConnWillMsg, testConnWillMsgBytesV311, t)
	t.Skip("v5")
	testPacketBytes(V5, testConnWillMsg, testConnWillMsgBytesV5, t)
}

func TestConnProps_Props(t *testing.T) {

}

func TestConnProps_SetProps(t *testing.T) {

}

func TestConnAckPacket_Bytes(t *testing.T) {
	testPacketBytes(V311, testConnAckMsg, testConnAckMsgBytesV311, t)
	t.Skip("v5")
	testPacketBytes(V5, testConnAckMsg, testConnAckMsgBytesV5, t)
}

func TestConnAckProps_Props(t *testing.T) {

}

func TestConnAckProps_SetProps(t *testing.T) {

}

func TestDisConnPacket_Bytes(t *testing.T) {
	testPacketBytes(V311, testDisConnMsg, testDisConnMsgBytesV311, t)
	t.Skip("v5")
	testPacketBytes(V5, testDisConnMsg, testDisConnMsgBytesV5, t)
}

func TestDisConnProps_Props(t *testing.T) {

}

func TestDisConnProps_SetProps(t *testing.T) {

}
