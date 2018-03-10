// Copyright (c) nano Author and TFG Co. All Rights Reserved.
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

package acceptor

import (
	"net"

	"github.com/lonnng/nano/logger"
)

// TCPAcceptor struct
type TCPAcceptor struct {
	addr     string
	connChan chan net.Conn
}

// NewTCPAcceptor creates a new instance of tcp acceptor
func NewTCPAcceptor(addr string) *TCPAcceptor {
	return &TCPAcceptor{
		addr:     addr,
		connChan: make(chan net.Conn),
	}
}

// GetAddr returns the addr the acceptor will listen on
func (a *TCPAcceptor) GetAddr() string {
	return a.addr
}

// GetConnChan gets a connection channel
func (a *TCPAcceptor) GetConnChan() chan net.Conn {
	return a.connChan
}

// ListenAndServe using tcp acceptor
func (a *TCPAcceptor) ListenAndServe() {
	listener, err := net.Listen("tcp", a.addr)
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Log.Error(err.Error())
			continue
		}

		a.connChan <- conn
	}
}
