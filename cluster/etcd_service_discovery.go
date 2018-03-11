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
	"context"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
	"github.com/lonnng/nano/logger"
)

// EtcdServiceDiscovery struct
type EtcdServiceDiscovery struct {
	cli                 *clientv3.Client
	heartbeatInterval   time.Duration
	syncServersInterval time.Duration
	heartbeatTTL        time.Duration
	leaseID             clientv3.LeaseID
}

// NewEtcdServiceDiscovery ctor
// TODO needs better configuration
func NewEtcdServiceDiscovery(
	endpoints []string,
	etcdDialTimeout time.Duration,
	etcdPrefix string,
	heartbeatInterval time.Duration,
	heartbeatTTL time.Duration,
	syncServersInterval time.Duration) (ServiceDiscovery, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: etcdDialTimeout,
	})
	if err != nil {
		return nil, err
	}

	// namespaced etcd :)
	cli.KV = namespace.NewKV(cli.KV, etcdPrefix)
	cli.Watcher = namespace.NewWatcher(cli.Watcher, etcdPrefix)
	cli.Lease = namespace.NewLease(cli.Lease, etcdPrefix)

	sd := &EtcdServiceDiscovery{
		cli:                 cli,
		heartbeatInterval:   heartbeatInterval,
		syncServersInterval: syncServersInterval,
		heartbeatTTL:        heartbeatTTL,
	}

	return sd, nil
}

func (sd *EtcdServiceDiscovery) bootstrapLease() error {
	// grab lease
	l, err := sd.cli.Grant(context.TODO(), int64(sd.heartbeatTTL.Seconds()))
	if err != nil {
		return err
	}
	sd.leaseID = l.ID
	return nil
}

func (sd *EtcdServiceDiscovery) bootstrapServer(server *Server) error {
	// put key
	_, err := sd.cli.Put(
		context.TODO(),
		getKey(server.ID, server.Type),
		fmt.Sprintf("%d", time.Now().Unix()),
		clientv3.WithLease(sd.leaseID),
	)
	if err != nil {
		return err
	}

	keys, err := sd.cli.Get(context.TODO(), "servers/", clientv3.WithPrefix())
	if err != nil {
		return err
	}

	logger.Log.Debugf("got servers %s", keys)
	return nil
}

// Start starts sending heartbeats, it should receive a server obj with current server info
func (sd *EtcdServiceDiscovery) Start(server *Server) {

	err := sd.bootstrapLease()
	if err != nil {
		logger.Log.Fatalf("error grabbing etcdv3 lease: %s", err.Error())
	}

	err = sd.bootstrapServer(server)
	if err != nil {
		logger.Log.Fatalf("error bootstrapping server info to etcdv3: %s", err.Error())
	}

	// send heartbeats
	heartbeatTicker := time.NewTicker(sd.heartbeatInterval)
	go func() {
		for range heartbeatTicker.C {
			err := sd.Heartbeat()
			if err != nil {
				logger.Log.Errorf("error sending heartbeat to etcd: %s", err.Error())
			}
		}
	}()

	// update servers
	syncServersTicker := time.NewTimer(sd.syncServersInterval)
	go func() {
		for range syncServersTicker.C {
			err := sd.UpdateServers()
			if err != nil {
				logger.Log.Errorf("error resyncing servers: %s", err.Error())
			}
		}
	}()
	// watch changes

	sd.watchEtcdChanges()
}

// UpdateServers gets all servers from etcd
func (sd *EtcdServiceDiscovery) UpdateServers() (map[string][]*Server, error) {
	return nil, nil
}

// Heartbeat sends a heartbeat to etcd
func (sd *EtcdServiceDiscovery) Heartbeat() error {
	logger.Log.Debugf("renewing heartbeat with lease %s", sd.leaseID)
	_, err := sd.cli.KeepAliveOnce(context.TODO(), sd.leaseID)
	if err != nil {
		return err
	}
	return nil
}

// Stop stops the service discovery client
func (sd *EtcdServiceDiscovery) Stop() {
	defer sd.cli.Close()
	logger.Log.Warn("stopping etcd service discovery client")
}

func (sd *EtcdServiceDiscovery) watchEtcdChanges() {
	w := sd.cli.Watch(context.Background(), "servers/", clientv3.WithPrefix())

	go func(chn clientv3.WatchChan) {
		for wresp := range chn {
			for _, ev := range wresp.Events {
				logger.Log.Info("hello ev %s", ev)
			}
		}
	}(w)
}

func getKey(serverID, serverType string) string {
	return fmt.Sprintf("servers/%s/%s", serverType, serverID)
}
