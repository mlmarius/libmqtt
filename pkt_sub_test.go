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

// sub test data
var (
	testSubTopics   []*Topic
	testSubMsgs     []*SubscribePacket
	testSubAckMsgs  []*SubAckPacket
	testUnsubMsgs   []*UnsubPacket
	testUnsubAckMsg *UnsubAckPacket

	// mqtt 3.1.1
	testSubMsgBytesV311      [][]byte
	testSubAckMsgBytesV311   [][]byte
	testUnsubMsgBytesV311    [][]byte
	testUnsubAckMsgBytesV311 []byte

	// mqtt 5.0
	testSubMsgBytesV5      [][]byte
	testSubAckMsgBytesV5   [][]byte
	testUnsubMsgBytesV5    [][]byte
	testUnsubAckMsgBytesV5 []byte
)

func initTestData_Sub() {
	size := len(testTopics)
	testSubTopics = make([]*Topic, size)
	for i := range testSubTopics {
		testSubTopics[i] = &Topic{Name: testTopics[i], Qos: testTopicQos[i]}
	}

	testSubMsgs = make([]*SubscribePacket, size)
	testSubAckMsgs = make([]*SubAckPacket, size)
	testUnsubMsgs = make([]*UnsubPacket, size)

	testSubMsgBytesV311 = make([][]byte, size)
	testSubAckMsgBytesV311 = make([][]byte, size)
	testUnsubMsgBytesV311 = make([][]byte, size)

	testSubMsgBytesV5 = make([][]byte, size)
	testSubAckMsgBytesV5 = make([][]byte, size)
	testUnsubMsgBytesV5 = make([][]byte, size)

	for i := range testTopics {
		msgID := uint16(i + 1)
		testSubMsgs[i] = &SubscribePacket{
			Topics:   testSubTopics[:i+1],
			PacketID: msgID,
			Props: &SubscribeProps{
				SubID:     100,
				UserProps: testConstUserProps,
			},
		}

		subPkt := std.NewControlPacket(std.Subscribe).(*std.SubscribePacket)
		subPkt.Topics = testTopics[:i+1]
		subPkt.Qoss = testTopicQos[:i+1]
		subPkt.MessageID = msgID

		subBuf := new(bytes.Buffer)
		_ = subPkt.Write(subBuf)
		testSubMsgBytesV311[i] = subBuf.Bytes()
		testSubMsgBytesV5[i] = newV5TestPacketBytes(CtrlSubscribe, 0, nil, nil)

		testSubAckMsgs[i] = &SubAckPacket{
			PacketID: msgID,
			Codes:    testSubAckCodes[:i+1],
			Props: &SubAckProps{
				Reason:    "MQTT",
				UserProps: testConstUserProps,
			},
		}
		subAckPkt := std.NewControlPacket(std.Suback).(*std.SubackPacket)
		subAckPkt.MessageID = msgID
		subAckPkt.ReturnCodes = testSubAckCodes[:i+1]
		subAckBuf := new(bytes.Buffer)
		_ = subAckPkt.Write(subAckBuf)
		testSubAckMsgBytesV311[i] = subAckBuf.Bytes()
		testSubAckMsgBytesV5[i] = newV5TestPacketBytes(CtrlSubAck, 0, nil, nil)

		testUnsubMsgs[i] = &UnsubPacket{
			PacketID:   msgID,
			TopicNames: testTopics[:i+1],
			Props: &UnsubProps{
				UserProps: testConstUserProps,
			},
		}
		unsubPkt := std.NewControlPacket(std.Unsubscribe).(*std.UnsubscribePacket)
		unsubPkt.Topics = testTopics[:i+1]
		unsubPkt.MessageID = msgID
		unSubBuf := new(bytes.Buffer)
		_ = unsubPkt.Write(unSubBuf)
		testUnsubMsgBytesV311[i] = unSubBuf.Bytes()
		testUnsubMsgBytesV5[i] = newV5TestPacketBytes(CtrlUnsub, 0, nil, nil)
	}

	unSunAckBuf := new(bytes.Buffer)
	testUnsubAckMsg = &UnsubAckPacket{
		PacketID: 1,
		Props: &UnsubAckProps{
			Reason:    "MQTT",
			UserProps: testConstUserProps,
		},
	}
	unsubAckPkt := std.NewControlPacket(std.Unsuback).(*std.UnsubackPacket)
	unsubAckPkt.MessageID = 1
	_ = unsubAckPkt.Write(unSunAckBuf)
	testUnsubAckMsgBytesV311 = unSunAckBuf.Bytes()
	testUnsubAckMsgBytesV5 = newV5TestPacketBytes(CtrlUnsubAck, 0, nil, nil)
}

func TestSubscribePacket_Bytes(t *testing.T) {
	for i, p := range testSubMsgs {
		testPacketBytes(V311, p, testSubMsgBytesV311[i], t)
		//testPacketBytes(V5, p, testSubMsgBytesV5[i], t)
	}
}

func TestSubAckPacket_Bytes(t *testing.T) {
	for i, p := range testSubAckMsgs {
		testPacketBytes(V311, p, testSubAckMsgBytesV311[i], t)
		//testPacketBytes(V5, p, testSubAckMsgBytesV5[i], t)
	}
}

func TestUnsubPacket_Bytes(t *testing.T) {
	for i, p := range testUnsubMsgs {
		testPacketBytes(V311, p, testUnsubMsgBytesV311[i], t)
		//testPacketBytes(V5, p, testUnsubMsgBytesV5[i], t)
	}
}

func TestUnsubAckPacket_Bytes(t *testing.T) {
	testPacketBytes(V311, testUnsubAckMsg, testUnsubAckMsgBytesV311, t)

	t.Skip("v5")
	testPacketBytes(V5, testUnsubAckMsg, testUnsubAckMsgBytesV5, t)
}
