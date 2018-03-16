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
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"time"

	"github.com/google/uuid"
	"github.com/lonnng/nano/acceptor"
	"github.com/lonnng/nano/cluster"
	"github.com/lonnng/nano/component"
	"github.com/lonnng/nano/internal/codec"
	"github.com/lonnng/nano/internal/message"
	"github.com/lonnng/nano/logger"
	"github.com/lonnng/nano/module"
	"github.com/lonnng/nano/protos"
	"github.com/lonnng/nano/route"
	"github.com/lonnng/nano/serialize"
	"github.com/lonnng/nano/serialize/protobuf"
	"github.com/lonnng/nano/session"
)

// App is the base app struct
type App struct {
	server           *cluster.Server
	debug            bool
	startAt          time.Time
	dieChan          chan bool
	acceptors        []acceptor.Acceptor
	heartbeat        time.Duration
	packetDecoder    codec.PacketDecoder
	packetEncoder    codec.PacketEncoder
	serializer       serialize.Serializer
	serviceDiscovery cluster.ServiceDiscovery
	rpcServer        cluster.RPCServer
	rpcClient        cluster.RPCClient
}

var (
	app = &App{
		server: &cluster.Server{
			ID:       uuid.New().String(),
			Type:     "game",
			Data:     map[string]string{},
			Frontend: true,
		},
		debug:         false,
		startAt:       time.Now(),
		dieChan:       make(chan bool),
		acceptors:     []acceptor.Acceptor{},
		heartbeat:     30 * time.Second,
		packetDecoder: codec.NewPomeloPacketDecoder(),
		packetEncoder: codec.NewPomeloPacketEncoder(),
		serializer:    protobuf.NewSerializer(),
	}
	log = logger.Log
)

// GetApp gets the app
func GetApp() *App {
	return app
}

// AddAcceptor adds a new acceptor to app
func AddAcceptor(ac acceptor.Acceptor) {
	app.acceptors = append(app.acceptors, ac)
}

// SetDebug toggles debug on/off
func SetDebug(debug bool) {
	app.debug = debug
}

// SetPacketDecoder changes the decoder used to parse messages received
func SetPacketDecoder(d codec.PacketDecoder) {
	app.packetDecoder = d
}

// SetPacketEncoder changes the encoder used to package outgoing messages
func SetPacketEncoder(e codec.PacketEncoder) {
	app.packetEncoder = e
}

// SetHeartbeatTime sets the heartbeat time
func SetHeartbeatTime(interval time.Duration) {
	app.heartbeat = interval
}

// SetRPCServer to be used
func SetRPCServer(s cluster.RPCServer) {
	app.rpcServer = s
}

// SetRPCClient to be used
func SetRPCClient(s cluster.RPCClient) {
	app.rpcClient = s
}

// SetServiceDiscoveryClient to be used
func SetServiceDiscoveryClient(s cluster.ServiceDiscovery) {
	app.serviceDiscovery = s
}

// SetSerializer customize application serializer, which automatically Marshal
// and UnMarshal handler payload
func SetSerializer(seri serialize.Serializer) {
	app.serializer = seri
}

// SetServerType sets the server type
// TODO need to specify in start
func SetServerType(t string) {
	app.server.Type = t
}

// SetServerData sets the server data that will be broadcasted using service discovery to other servers
// TODO need to specify in start
func SetServerData(data map[string]string) {
	app.server.Data = data
}

func startDefaultSD() {
	// initialize default service discovery
	// TODO remove this, force specifying
	var err error
	app.serviceDiscovery, err = cluster.NewEtcdServiceDiscovery(
		[]string{"localhost:2379"},
		time.Duration(5)*time.Second,
		"nano/",
		time.Duration(20)*time.Second,
		time.Duration(60)*time.Second,
		time.Duration(120)*time.Second,
		app.server,
	)
	if err != nil {
		log.Fatalf("error starting cluster service discovery component: %s", err.Error())
	}
}

func startDefaultRPCServer() {
	// initialize default rpc server
	// TODO remove this, force specifying
	var err error
	app.rpcServer = cluster.NewNatsRPCServer(
		"nats://localhost:4222",
		app.server,
	)
	if err != nil {
		log.Fatalf("error starting cluster rpc server component: %s", err.Error())
	}
}

func startDefaultRPCClient() {
	// initialize default rpc client
	// TODO remove this, force specifying
	var err error
	app.rpcClient = cluster.NewNatsRPCClient(
		"nats://localhost:4222",
		app.server,
	)
	if err != nil {
		log.Fatalf("error starting cluster rpc client component: %s", err.Error())
	}
}

// Start starts the app
// TODO fix non cluster mode
func Start(isFrontend bool) {
	app.server.Frontend = isFrontend

	if app.serviceDiscovery == nil {
		log.Warn("creating default service discovery because cluster mode is enabled, if you want to specify yours, use nano.SetServiceDiscoveryClient")
		startDefaultSD()
	}
	if app.rpcServer == nil {
		log.Warn("creating default rpc server because cluster mode is enabled, if you want to specify yours, use nano.SetRPCServer")
		startDefaultRPCServer()
	}
	if app.rpcClient == nil {
		log.Warn("creating default rpc client because cluster mode is enabled, if you want to specify yours, use nano.SetRPCClient")
		startDefaultRPCClient()
		RegisterModule(app.serviceDiscovery, "serviceDiscovery")
		RegisterModule(app.rpcServer, "rpcServer")
		RegisterModule(app.rpcClient, "rpcClient")
	}

	listen()

	sg := make(chan os.Signal)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)

	// stop server
	select {
	case <-app.dieChan:
		log.Warn("The app will shutdown in a few seconds")
	case s := <-sg:
		log.Warn("got signal", s)
	}

	log.Warn("server is stopping...")

	shutdownModules()
	// shutdown all components registered by application, that
	// call by reverse order against register
	shutdownComponents()

}

func listen() {
	hbdEncode()
	startupComponents()

	// create global ticker instance, timer precision could be customized
	// by SetTimerPrecision
	globalTicker = time.NewTicker(timerPrecision)

	log.Infof("starting server %s:%s", app.server.Type, app.server.ID)

	// startup logic dispatcher
	go handler.dispatch()
	for _, acc := range app.acceptors {
		a := acc

		// gets connections from every acceptor and tell handlerservice to handle them
		go func() {
			for conn := range a.GetConnChan() {
				go handler.handle(conn)
			}
		}()

		go func() {
			a.ListenAndServe()
		}()

		log.Infof("listening with acceptor %s on addr %s", reflect.TypeOf(a), a.GetAddr())
	}
	startModules()

	// this handles remote messages
	// TODO probably this shouldnt be here :/
	if app.rpcServer != nil {
		// TODO config concurrency, should this be done this way?
		processMsgConcurrency := 100
		for i := 0; i < processMsgConcurrency; i++ {
			go processRemoteMessages(i)
		}
	}
}

// TODO own file?
func remoteCall(rpcType protos.RPCType, route *route.Route, session *session.Session, msg *message.Message) ([]byte, error) {
	svType := route.SvType
	//TODO this logic should be elsewhere, routing should be changeable
	serversOfType, err := app.serviceDiscovery.GetServersByType(svType)
	if err != nil {
		return nil, err
	}

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	server := serversOfType[r.Intn(len(serversOfType))]

	res, err := app.rpcClient.Call(rpcType, route, session, msg, server)
	if err != nil {
		return nil, err
	}
	return res, err
}

// SetDictionary set routes map, TODO(warning): set dictionary in runtime would be a dangerous operation!!!!!!
func SetDictionary(dict map[string]uint16) {
	message.SetDictionary(dict)
}

// Register register a component with options
func Register(c component.Component, options ...component.Option) {
	comps = append(comps, regComp{c, options})
}

// RegisterModule register a module
func RegisterModule(m module.Module, name string) error {
	if _, ok := modules[name]; ok {
		return fmt.Errorf(
			"a module names %s was already registered", name,
		)
	}
	modules[name] = m
	return nil
}

// Shutdown send a signal to let 'nano' shutdown itself.
func Shutdown() {
	close(app.dieChan)
}
