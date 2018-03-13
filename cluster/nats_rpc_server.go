// Copyright (c) TFG Co. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cluster

import (
	"fmt"

	nats "github.com/nats-io/go-nats"
)

// NatsRPCServer struct
type NatsRPCServer struct {
	connString string
	server     *Server
	conn       *nats.Conn
	inMsgChan  chan *nats.Msg
	running    bool
	stopChan   chan bool
}

// NewNatsRPCServer ctor
func NewNatsRPCServer(connectString string, server *Server) *NatsRPCServer {
	ns := &NatsRPCServer{
		connString: connectString,
		server:     server,
		inMsgChan:  make(chan *nats.Msg),
		running:    false,
		stopChan:   make(chan bool),
	}
	return ns
}

func (ns *NatsRPCServer) handleMessages() {
	for ns.running {
		select {
		case msg := <-ns.inMsgChan:
			log.Debugf("got msg from nats: %s", string(msg.Data))
		case <-ns.stopChan:
			ns.stop()
			break
		}
	}
}

// Init inits nats rpc server
func (ns *NatsRPCServer) Init() error {
	conn, err := setupNatsConn(ns.connString)
	if err != nil {
		return err
	}
	ns.conn = conn
	ns.running = true
	var c chan *nats.Msg
	if c, err = ns.subscribe(getChannel(ns.server.Type, ns.server.ID)); err != nil {
		return err
	}
	ns.inMsgChan = c

	go ns.handleMessages()

	return nil
}

// AfterInit runs after initialization
func (ns *NatsRPCServer) AfterInit() {}

// BeforeShutdown runs before shutdown
func (ns *NatsRPCServer) BeforeShutdown() {}

// Shutdown stops nats rpc server
func (ns *NatsRPCServer) Shutdown() error {
	ns.stopChan <- true
	return nil
}

func (ns *NatsRPCServer) subscribe(topic string) (chan *nats.Msg, error) {
	c := make(chan *nats.Msg)
	if _, err := ns.conn.ChanSubscribe(topic, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (ns *NatsRPCServer) stop() {
	ns.running = false
}

func getChannel(serverType, serverID string) string {
	return fmt.Sprintf("nano/servers/%s/%s", serverType, serverID)
}

func setupNatsConn(connectString string) (*nats.Conn, error) {
	nc, err := nats.Connect(connectString,
		nats.DisconnectHandler(func(_ *nats.Conn) {
			log.Warn("disconnected from nats!")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Warnf("reconnected to nats %s!", nc.ConnectedUrl)
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Warnf("nats connection closed. reason: %q", nc.LastError())
		}),
	)
	if err != nil {
		return nil, err
	}
	return nc, nil
}
