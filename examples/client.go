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

package examples

import (
	"log"
	"time"

	"github.com/goiiot/libmqtt"
)

// ExampleClient example of client creation
func ExampleClient() {
	var (
		client libmqtt.Client
		err    error
	)

	client, err = libmqtt.NewClient(
		// try MQTT 5.0 and fallback to MQTT 3.1.1
		libmqtt.WithVersion(libmqtt.V5, true),
		// enable keepalive (10s interval) with 20% tolerance
		libmqtt.WithKeepalive(10, 1.2),
		// enable auto reconnect and set backoff strategy
		libmqtt.WithAutoReconnect(true),
		libmqtt.WithBackoffStrategy(time.Second, 5*time.Second, 1.2),
		// use RegexRouter for topic routing if not specified
		// will use TextRouter, which will match full text
		libmqtt.WithRouter(libmqtt.NewRegexRouter()),
		libmqtt.WithConnHandleFunc(connHandler),
		libmqtt.WithNetHandleFunc(netHandler),
		libmqtt.WithSubHandleFunc(subHandler),
		libmqtt.WithUnsubHandleFunc(unSubHandler),
		libmqtt.WithPubHandleFunc(pubHandler),
		libmqtt.WithPersistHandleFunc(persistHandler),
	)

	if err != nil {
		// handle client creation error
		panic("hmm, how could it failed")
	}

	// handle every subscribed message (just for example)
	client.HandleTopic(".*", func(client libmqtt.Client, topic string, qos libmqtt.QosLevel, msg []byte) {
		log.Printf("[%v] message: %v", topic, string(msg))
	})

	// connect to server

	// connect tcp server
	err = client.ConnectServer("test.mosquitto.org:1883")

	// connect a tcp (tls) server
	//
	// use `libmqtt.WithTLS` or `libmqtt.WithTLSReader` to use client tls certificate
	// and to use customized tls config, you should use `libmqtt.WithCustomTLS`
	err = client.ConnectServer("test.mosquitto.org:8883",
		libmqtt.WithCustomTLS(nil))

	// connect to a websocket server
	err = client.ConnectServer("test.mosquitto.org:8080",
		libmqtt.WithWebSocketConnector(0, nil))

	// connect to a websocket (tls) server
	err = client.ConnectServer("test.mosquitto.org:8081",
		libmqtt.WithCustomTLS(nil),
		libmqtt.WithWebSocketConnector(0, nil))

	client.Wait()
}

func connHandler(client libmqtt.Client, server string, code byte, err error) {
	if err != nil {
		log.Printf("connect to server [%v] failed: %v", server, err)
		return
	}

	if code != libmqtt.CodeSuccess {
		log.Printf("connect to server [%v] failed with server code [%v]", server, code)
		return
	}

	// connected
	go func() {
		// subscribe to some topics
		client.Subscribe([]*libmqtt.Topic{
			{Name: "foo", Qos: libmqtt.Qos0},
			{Name: "bar", Qos: libmqtt.Qos1},
		}...)

		// in this example, we publish packets right after subscribe succeeded
		// see `client.HandleSub`
	}()
}

func netHandler(client libmqtt.Client, server string, err error) {
	if err != nil {
		log.Printf("error happened to connection to server [%v]: %v", server, err)
	}
}

func persistHandler(client libmqtt.Client, packet libmqtt.Packet, err error) {
	if err != nil {
		log.Printf("session persist error: %v", err)
	}
}

func subHandler(client libmqtt.Client, topics []*libmqtt.Topic, err error) {
	if err != nil {
		for _, t := range topics {
			log.Printf("subscribe to topic [%v] failed: %v", t.Name, err)
		}
	} else {
		for _, t := range topics {
			log.Printf("subscribe to topic [%v] success: %v", t.Name, err)
		}

		// publish some packet (just for example)
		client.Publish([]*libmqtt.PublishPacket{
			{TopicName: "foo", Payload: []byte("bar"), Qos: libmqtt.Qos0},
			{TopicName: "bar", Payload: []byte("foo"), Qos: libmqtt.Qos1},
		}...)
	}
}

func unSubHandler(client libmqtt.Client, topic []string, err error) {
	if err != nil {
		// handle unsubscribe failure
		for _, t := range topic {
			log.Printf("unsubscribe to topic [%v] failed: %v", t, err)
		}
	} else {
		for _, t := range topic {
			log.Printf("unsubscribe to topic [%v] failed: %v", t, err)
		}
	}
}

func pubHandler(client libmqtt.Client, topic string, err error) {
	if err != nil {
		log.Printf("publish packet to topic [%v] failed: %v", topic, err)
	} else {
		log.Printf("publish packet to topic [%v] success: %v", topic, err)
	}
}
