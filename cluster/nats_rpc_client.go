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
	"time"

	nats "github.com/nats-io/go-nats"
)

// NatsRPCClient struct
type NatsRPCClient struct {
	connString string
	server     *Server
	conn       *nats.Conn
	reqTimeout time.Duration
	running    bool
}

// NewNatsRPCClient ctor
func NewNatsRPCClient(connectString string, server *Server) *NatsRPCClient {
	ns := &NatsRPCClient{
		connString: connectString,
		reqTimeout: time.Duration(5) * time.Second,
		server:     server,
		running:    false,
	}
	return ns
}

// Call calls a method remotally
// TODO oh my, this is hacky! will it perform good?
// even if it performs we need better concurrency control!!!
func (ns *NatsRPCClient) Call(server *Server, route string, data []byte) ([]byte, error) {
	topic := getChannel(server.Type, server.ID)
	msg, err := ns.conn.Request(topic, data, ns.reqTimeout)
	if err != nil {
		return nil, err
	}
	return msg.Data, nil
}

// Init inits nats rpc server
func (ns *NatsRPCClient) Init() error {
	ns.running = true
	conn, err := setupNatsConn(ns.connString)
	if err != nil {
		return err
	}
	ns.conn = conn
	return nil
}

// AfterInit runs after initialization
func (ns *NatsRPCClient) AfterInit() {}

// BeforeShutdown runs before shutdown
func (ns *NatsRPCClient) BeforeShutdown() {}

// Shutdown stops nats rpc server
func (ns *NatsRPCClient) Shutdown() error {
	return nil
}

func (ns *NatsRPCClient) subscribe(topic string) (chan *nats.Msg, error) {
	c := make(chan *nats.Msg)
	if _, err := ns.conn.ChanSubscribe(ns.getSubscribeChannel(), c); err != nil {
		return nil, err
	}
	return c, nil
}

func (ns *NatsRPCClient) stop() {
	ns.running = false
}

func (ns *NatsRPCClient) getSubscribeChannel() string {
	return fmt.Sprintf("nano/servers/%s/%s", ns.server.Type, ns.server.ID)
}
