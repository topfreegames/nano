// Copyright (c) nano Author. All Rights Reserved.
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

package nano

import (
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"time"

	"github.com/google/uuid"
	"github.com/lonnng/nano/component"
	"github.com/lonnng/nano/internal/message"
)

// App is the base app struct
type App struct {
	serverType string
	serverID   string
	debug      bool
	startAt    time.Time
	dieChan    chan bool
	acceptors  []Acceptor
	heartbeat  time.Duration
}

var (
	app = &App{
		serverID:   uuid.New().String(),
		debug:      false,
		serverType: "game",
		startAt:    time.Now(),
		dieChan:    make(chan bool),
		acceptors:  []Acceptor{},
		heartbeat:  30 * time.Second,
	}
)

// GetApp gets the app
func GetApp() *App {
	return app
}

// AddAcceptor adds a new acceptor to app
func AddAcceptor(ac Acceptor) {
	app.acceptors = append(app.acceptors, ac)
}

// SetDebug toggles debug on/off
func SetDebug(debug bool) {
	app.debug = debug
}

// SetHeartbeatTime sets the heartbeat time
func SetHeartbeatTime(interval time.Duration) {
	app.heartbeat = interval
}

// SetServerType sets the server type
func SetServerType(t string) {
	app.serverType = t
}

// Start starts the app
func Start() {
	listen()
}

func listen() {
	hbdEncode()
	startupComponents()

	// create global ticker instance, timer precision could be customized
	// by SetTimerPrecision
	globalTicker = time.NewTicker(timerPrecision)

	logger.Infof("starting server %s:%s", app.serverType, app.serverID)

	// startup logic dispatcher
	go handler.dispatch()
	for _, acc := range app.acceptors {
		a := acc

		// gets connections from every acceptor and tell handlerservice to handle them
		go func() {
			for conn := range a.GetConnChan() {
				handler.handle(conn)
			}
		}()

		go func() {
			a.ListenAndServe()
		}()

		logger.Infof("listening with acceptor %s on addr %s", reflect.TypeOf(a), a.GetAddr())
	}

	sg := make(chan os.Signal)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)

	// stop server
	select {
	case <-app.dieChan:
		logger.Warn("The app will shutdown in a few seconds")
	case s := <-sg:
		logger.Warn("got signal", s)
	}

	logger.Warn("server is stopping...")

	// shutdown all components registered by application, that
	// call by reverse order against register
	shutdownComponents()
}

// SetDictionary set routes map, TODO(warning): set dictionary in runtime would be a dangerous operation!!!!!!
func SetDictionary(dict map[string]uint16) {
	message.SetDictionary(dict)
}

// Register register a component with options
func Register(c component.Component, options ...component.Option) {
	comps = append(comps, regComp{c, options})
}

// Shutdown send a signal to let 'nano' shutdown itself.
func Shutdown() {
	close(app.dieChan)
}
