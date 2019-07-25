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

// ConnHandleFunc is the handler which tend to the Connect result
// server is the server address provided by user in client creation call
// code is the ConnResult code
// err is the error happened when connect to server, if a error happened,
// the code value will max byte value (255)
type ConnHandleFunc func(client Client, server string, code byte, err error)

// TopicHandleFunc handles topic sub message
// topic is the client user provided topic
// code can be SubOkMaxQos0, SubOkMaxQos1, SubOkMaxQos2, SubFail
type TopicHandleFunc func(client Client, topic string, qos QosLevel, msg []byte)

// PubHandleFunc handles the error occurred when publish some message
// if err is not nil, that means a error occurred when sending pub msg
type PubHandleFunc func(client Client, topic string, err error)

// SubHandleFunc handles the error occurred when subscribe some topic
// if err is not nil, that means a error occurred when sending sub msg
type SubHandleFunc func(client Client, topics []*Topic, err error)

// UnsubHandleFunc handles the error occurred when publish some message
type UnsubHandleFunc func(client Client, topics []string, err error)

// NetHandleFunc handles the error occurred when net broken
type NetHandleFunc func(client Client, server string, err error)

// PersistHandleFunc handles err happened when persist process has trouble
type PersistHandleFunc func(client Client, packet Packet, err error)
