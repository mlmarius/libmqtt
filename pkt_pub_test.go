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

// pub test data
var (
	testPubMsgs   []*PublishPacket
	testPubAckMsg = &PubAckPacket{
		PacketID: testPacketID,
		Code:     CodeUnspecifiedError,
		Props: &PubAckProps{
			Reason:    "MQTT",
			UserProps: testConstUserProps,
		},
	}
	testPubRecvMsg = &PubRecvPacket{
		PacketID: testPacketID,
		Code:     CodeUnspecifiedError,
		Props: &PubRecvProps{
			Reason:    "MQTT",
			UserProps: testConstUserProps,
		},
	}
	testPubRelMsg = &PubRelPacket{
		PacketID: testPacketID,
		Code:     CodeUnspecifiedError,
		Props: &PubRelProps{
			Reason:    "MQTT",
			UserProps: testConstUserProps,
		},
	}
	testPubCompMsg = &PubCompPacket{
		PacketID: testPacketID,
		Code:     CodeUnspecifiedError,
		Props: &PubCompProps{
			Reason:    "MQTT",
			UserProps: testConstUserProps,
		},
	}

	// mqtt 3.1.1
	testPubMsgBytesV311     [][]byte
	testPubAckMsgBytesV311  []byte
	testPubRecvMsgBytesV311 []byte
	testPubRelMsgBytesV311  []byte
	testPubCompMsgBytesV311 []byte

	// mqtt 5.0
	testPubMsgBytesV5     [][]byte
	testPubAckMsgBytesV5  []byte
	testPubRecvMsgBytesV5 []byte
	testPubRelMsgBytesV5  []byte
	testPubCompMsgBytesV5 []byte
)

// init pub test data
func initTestData_Pub() {
	size := len(testTopics)

	testPubMsgs = make([]*PublishPacket, size)
	testPubMsgBytesV311 = make([][]byte, size)
	testPubMsgBytesV5 = make([][]byte, size)

	for i := range testTopics {
		msgID := uint16(i + 1)
		testPubMsgs[i] = &PublishPacket{
			IsDup:     testPubDup,
			TopicName: testTopics[i],
			Qos:       testTopicQos[i],
			Payload:   []byte(testTopicMsgs[i]),
			PacketID:  msgID,
			Props: &PublishProps{
				PayloadFormat:         100,
				MessageExpiryInterval: 100,
				TopicAlias:            100,
				RespTopic:             "MQTT",
				CorrelationData:       []byte("MQTT"),
				UserProps:             testConstUserProps,
				SubIDs:                []int{1, 2, 3},
				ContentType:           "MQTT",
			},
		}

		// create standard publish packet and make bytes
		pkt := std.NewControlPacketWithHeader(std.FixedHeader{
			MessageType: std.Publish,
			Dup:         testPubDup,
			Qos:         testTopicQos[i],
		}).(*std.PublishPacket)
		pkt.TopicName = testTopics[i]
		pkt.Payload = []byte(testTopicMsgs[i])
		pkt.MessageID = msgID

		buf := &bytes.Buffer{}
		_ = pkt.Write(buf)
		testPubMsgBytesV311[i] = buf.Bytes()
		testPubMsgBytesV5[i] = newV5TestPacketBytes(CtrlPublish, 0, nil, nil)
	}

	// puback
	pubAckPkt := std.NewControlPacket(std.Puback).(*std.PubackPacket)
	pubAckPkt.MessageID = testPacketID
	pubAckBuf := &bytes.Buffer{}
	_ = pubAckPkt.Write(pubAckBuf)
	testPubAckMsgBytesV311 = pubAckBuf.Bytes()
	testPubAckMsgBytesV5 = newV5TestPacketBytes(CtrlPubAck, 0, nil, nil)

	// pubrecv
	pubRecvBuf := &bytes.Buffer{}
	pubRecPkt := std.NewControlPacket(std.Pubrec).(*std.PubrecPacket)
	pubRecPkt.MessageID = testPacketID
	_ = pubRecPkt.Write(pubRecvBuf)
	testPubRecvMsgBytesV311 = pubRecvBuf.Bytes()
	testPubRecvMsgBytesV5 = newV5TestPacketBytes(CtrlPubRecv, 0, nil, nil)

	// pubrel
	pubRelBuf := &bytes.Buffer{}
	pubRelPkt := std.NewControlPacket(std.Pubrel).(*std.PubrelPacket)
	pubRelPkt.MessageID = testPacketID
	_ = pubRelPkt.Write(pubRelBuf)
	testPubRelMsgBytesV311 = pubRelBuf.Bytes()
	testPubRelMsgBytesV5 = newV5TestPacketBytes(CtrlPubRel, 0, nil, nil)

	// pubcomp
	pubCompBuf := &bytes.Buffer{}
	pubCompPkt := std.NewControlPacket(std.Pubcomp).(*std.PubcompPacket)
	pubCompPkt.MessageID = testPacketID
	_ = pubCompPkt.Write(pubCompBuf)
	testPubCompMsgBytesV311 = pubCompBuf.Bytes()
	testPubCompMsgBytesV5 = newV5TestPacketBytes(CtrlPubComp, 0, nil, nil)
}

func TestPublishPacket_Bytes(t *testing.T) {
	for i, p := range testPubMsgs {
		testPacketBytes(V311, p, testPubMsgBytesV311[i], t)
		testPacketBytes(V5, p, testPubMsgBytesV5[i], t)
	}
}

func TestPubProps_Props(t *testing.T) {

}

func TestPubProps_SetProps(t *testing.T) {

}

func TestPubAckPacket_Bytes(t *testing.T) {
	testPacketBytes(V311, testPubAckMsg, testPubAckMsgBytesV311, t)
	testPacketBytes(V5, testPubAckMsg, testPubAckMsgBytesV5, t)
}

func TestPubAckProps_Props(t *testing.T) {

}

func TestPubAckProps_SetProps(t *testing.T) {

}

func TestPubRecvPacket_Bytes(t *testing.T) {
	testPacketBytes(V311, testPubRecvMsg, testPubRecvMsgBytesV311, t)
	testPacketBytes(V5, testPubRecvMsg, testPubRecvMsgBytesV5, t)
}

func TestPubRecvProps_Props(t *testing.T) {

}

func TestPubRecvProps_SetProps(t *testing.T) {

}

func TestPubRelPacket_Bytes(t *testing.T) {
	testPacketBytes(V311, testPubRelMsg, testPubRelMsgBytesV311, t)
	testPacketBytes(V5, testPubRelMsg, testPubRelMsgBytesV5, t)
}

func TestPubRelProps_Props(t *testing.T) {

}

func TestPubRelProps_SetProps(t *testing.T) {

}

func TestPubCompPacket_Bytes(t *testing.T) {
	testPacketBytes(V311, testPubCompMsg, testPubCompMsgBytesV311, t)
	testPacketBytes(V5, testPubCompMsg, testPubCompMsgBytesV5, t)
}

func TestPubCompProps_Props(t *testing.T) {

}

func TestPubCompProps_SetProps(t *testing.T) {

}
