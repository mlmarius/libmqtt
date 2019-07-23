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
)

var (
	testAuthMsg = &AuthPacket{
		BasePacket: BasePacket{ProtoVersion: V5},
		Code:       CodeReAuth,
		Props: &AuthProps{
			AuthMethod: "MQTT",
			AuthData:   []byte("MQTT"),
			Reason:     "MQTT",
			UserProps:  testConstUserProps,
		},
	}
	testAuthMsgBytes []byte

	testProps = map[byte][]byte{
		propKeyAuthMethod:   {0, 4, 'M', 'Q', 'T', 'T'},
		propKeyAuthData:     {0, 4, 'M', 'Q', 'T', 'T'},
		propKeyReasonString: {0, 4, 'M', 'Q', 'T', 'T'},
		propKeyUserProps:    testConstUserPropsBytes,
	}

	testAuthPropsBytes = append([]byte{
		propKeyAuthMethod, 0, 4, 'M', 'Q', 'T', 'T',
		propKeyAuthData, 0, 4, 'M', 'Q', 'T', 'T',
		propKeyReasonString, 0, 4, 'M', 'Q', 'T', 'T',
		propKeyUserProps}, testConstUserPropsBytes...)
)

func initTestData_Auth() {
	testAuthMsgBytes = newV5TestPacketBytes(CtrlAuth, 0, append([]byte{CodeReAuth}, testAuthPropsBytes...), nil)
}

func TestAuthPacket_Bytes(t *testing.T) {
	testPacketBytes(V5, testAuthMsg, testAuthMsgBytes, t)
}

func TestAuthProps_Props(t *testing.T) {
	propsBytes := testAuthMsg.Props.props()
	if !bytes.Equal(propsBytes, testAuthPropsBytes) {
		t.Errorf("auth props bytes not math:\ntarget: %v\ngenerated: %v", testAuthPropsBytes, propsBytes)
	}
}

func TestAuthProps_SetProps(t *testing.T) {
	emptyProps := &AuthProps{}
	emptyProps.setProps(testProps)

	if emptyProps.AuthMethod != testAuthMsg.Props.AuthMethod {
		t.Error("auth method set failed")
	}

	if bytes.Compare(emptyProps.AuthData, testAuthMsg.Props.AuthData) != 0 {
		t.Error("auth data set failed")
	}

	if emptyProps.Reason != testAuthMsg.Props.Reason {
		t.Error("auth reason set failed")
	}

	for k, v := range emptyProps.UserProps {
		if tv, ok := testAuthMsg.Props.UserProps[k]; ok {
			if len(v) == len(tv) {
				for i := range v {
					if v[i] != tv[i] {
						t.Error("auth user props set failed")
					}
				}
				continue
			} else {
				t.Error("auth user props length not equal")
			}
		}
	}
}
