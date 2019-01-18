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
	"context"
	"crypto/tls"
	"strings"
	"sync"
)

// Client type for *AsyncClient
type Client = *AsyncClient

// NewClient create a new mqtt client
func NewClient(options ...Option) (Client, error) {
	c := defaultClient()

	for _, setOption := range options {
		err := setOption(c, &c.options)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

// AsyncClient mqtt client implementation
type AsyncClient struct {
	// Deprecated: use ConnectServer instead (will be removed in v1.0)
	servers []string
	// Deprecated: use ConnectServer instead (will be removed in v1.0)
	secureServers []string
	options       connectOptions // client connection options

	msgCh   chan *message       // error channel
	sendCh  chan Packet         // pub channel for sending publish packet to server
	recvCh  chan *PublishPacket // recv channel for server pub receiving
	idGen   *idGenerator        // Packet id generator
	router  TopicRouter         // Topic router
	persist PersistMethod       // Persist method
	workers *sync.WaitGroup     // Workers (goroutines)
	log     *logger             // client logger

	// success/error handlers
	pubHandler     PubHandler
	subHandler     SubHandler
	unSubHandler   UnSubHandler
	netHandler     NetHandler
	persistHandler PersistHandler

	ctx  context.Context    // closure of this channel will signal all client worker to stop
	exit context.CancelFunc // called when client exit
}

// create a client with default options
func defaultClient() *AsyncClient {
	ctx, exitFunc := context.WithCancel(context.Background())

	return &AsyncClient{
		servers:       make([]string, 0, 1),
		secureServers: make([]string, 0, 1),

		options: defaultConnectOptions(),
		msgCh:   make(chan *message, 10),
		sendCh:  make(chan Packet, 1),
		recvCh:  make(chan *PublishPacket, 1),
		ctx:     ctx,
		exit:    exitFunc,
		router:  NewTextRouter(),
		idGen:   newIDGenerator(),
		workers: &sync.WaitGroup{},
		persist: NonePersist,
	}
}

// Handle register subscription message route
func (c *AsyncClient) Handle(topic string, h TopicHandler) {
	if h != nil {
		c.log.d("HDL registered topic handler, topic =", topic)
		c.router.Handle(topic, h)
	}
}

// Connect to all designated servers
//
// Deprecated: use Client.ConnectServer instead (will be removed in v1.0)
func (c *AsyncClient) Connect(h ConnHandler) {
	c.log.d("CLI connect to server, handler =", h)

	for _, s := range c.servers {
		c.workers.Add(1)
		go c.options.connect(c, s, c.options.protoVersion, c.options.firstDelay, h)
	}

	for _, s := range c.secureServers {
		secureOptions := c.options.clone()
		secureOptions.tlsConfig = &tls.Config{
			ServerName: strings.SplitN(s, ":", 1)[0],
		}
		c.workers.Add(1)
		go secureOptions.connect(c, s, secureOptions.protoVersion, secureOptions.firstDelay, h)
	}

	c.workers.Add(2)
	go c.handleTopicMsg()
	go c.handleMsg()
}

// Publish message(s) to topic(s), one to one
func (c *AsyncClient) Publish(msg ...*PublishPacket) {
	if c.isClosing() {
		return
	}

	for _, m := range msg {
		if m == nil {
			continue
		}

		p := m
		if p.Qos > Qos2 {
			p.Qos = Qos2
		}

		if p.Qos != Qos0 {
			if p.PacketID == 0 {
				p.PacketID = c.idGen.next(p)
				if err := c.persist.Store(sendKey(p.PacketID), p); err != nil {
					notifyPersistMsg(c.msgCh, err)
				}
			}
		}
		c.sendCh <- p
	}
}

// Subscribe topic(s)
func (c *AsyncClient) Subscribe(topics ...*Topic) {
	if c.isClosing() {
		return
	}

	c.log.d("CLI subscribe, topic(s) =", topics)

	s := &SubscribePacket{Topics: topics}
	s.PacketID = c.idGen.next(s)

	c.sendCh <- s
}

// UnSubscribe topic(s)
func (c *AsyncClient) UnSubscribe(topics ...string) {
	if c.isClosing() {
		return
	}

	c.log.d("CLI unsubscribe topic(s) =", topics)

	u := &UnSubPacket{TopicNames: topics}
	u.PacketID = c.idGen.next(u)

	c.sendCh <- u
}

// Wait will wait for all connections to exit
func (c *AsyncClient) Wait() {
	if c.isClosing() {
		return
	}

	c.log.i("CLI wait for all workers")
	c.workers.Wait()
}

// Destroy will disconnect form all server
// If force is true, then close connection without sending a DisConnPacket
func (c *AsyncClient) Destroy(force bool) {
	c.log.d("CLI destroying client with force =", force)
	if force {
		c.exit()
	} else {
		c.sendCh <- &DisConnPacket{}
	}
}

func (c *AsyncClient) Disconnect(server string, packet *DisConnPacket) {
	if packet == nil {
		packet = &DisConnPacket{}
	}

	c.sendCh <- packet
}

func (c *AsyncClient) isClosing() bool {
	select {
	case <-c.ctx.Done():
		return true
	default:
		return false
	}
}

func (c *AsyncClient) handleTopicMsg() {
	defer c.workers.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case pkt, more := <-c.recvCh:
			if !more {
				return
			}

			c.router.Dispatch(pkt)
		}
	}
}

func (c *AsyncClient) handleMsg() {
	defer c.workers.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case m, more := <-c.msgCh:
			if !more {
				return
			}

			switch m.what {
			case pubMsg:
				if c.pubHandler != nil {
					go c.pubHandler(m.msg, m.err)
				}
			case subMsg:
				if c.subHandler != nil {
					go c.subHandler(m.obj.([]*Topic), m.err)
				}
			case unSubMsg:
				if c.unSubHandler != nil {
					go c.unSubHandler(m.obj.([]string), m.err)
				}
			case netMsg:
				if c.netHandler != nil {
					go c.netHandler(m.msg, m.err)
				}
			case persistMsg:
				if c.persistHandler != nil {
					go c.persistHandler(m.err)
				}
			}
		}
	}
}
