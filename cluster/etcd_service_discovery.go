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
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
)

// EtcdServiceDiscovery struct
type EtcdServiceDiscovery struct {
	cli                 *clientv3.Client
	heartbeatInterval   time.Duration
	syncServersInterval time.Duration
	heartbeatTTL        time.Duration
	leaseID             clientv3.LeaseID
	serverMapByType     sync.Map
	serverMapByID       sync.Map
	running             bool
	server              *Server
}

// NewEtcdServiceDiscovery ctor
// TODO needs better configuration
func NewEtcdServiceDiscovery(
	endpoints []string,
	etcdDialTimeout time.Duration,
	etcdPrefix string,
	heartbeatInterval time.Duration,
	heartbeatTTL time.Duration,
	syncServersInterval time.Duration,
	server *Server) (ServiceDiscovery, error) {
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
		running:             false,
		server:              server,
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
		server.GetDataAsJSONString(),
		clientv3.WithLease(sd.leaseID),
	)
	if err != nil {
		return err
	}

	sd.SyncServers()

	return nil
}

// Init starts the service discovery client
func (sd *EtcdServiceDiscovery) Init() error {
	err := sd.bootstrapLease()
	if err != nil {
		return err
	}

	err = sd.bootstrapServer(sd.server)
	if err != nil {
		return err
	}

	// send heartbeats
	heartbeatTicker := time.NewTicker(sd.heartbeatInterval)
	go func() {
		for range heartbeatTicker.C {
			err := sd.Heartbeat()
			if err != nil {
				log.Errorf("error sending heartbeat to etcd: %s", err.Error())
			}
		}
	}()

	// update servers
	syncServersTicker := time.NewTimer(sd.syncServersInterval)
	go func() {
		for range syncServersTicker.C {
			err := sd.SyncServers()
			if err != nil {
				log.Errorf("error resyncing servers: %s", err.Error())
			}
		}
	}()

	go sd.watchEtcdChanges()
	return nil
}

// AfterInit executes after Init
func (sd *EtcdServiceDiscovery) AfterInit() {
}

// BeforeShutdown executes before shutting down
func (sd *EtcdServiceDiscovery) BeforeShutdown() {
}

// Shutdown executes on shutdown and will clean etcd
func (sd *EtcdServiceDiscovery) Shutdown() error {
	return nil
}

// GetServers returns a map with all types as keys and an array of servers
func (sd *EtcdServiceDiscovery) GetServers() map[string]*Server {
	// TODO implement
	return nil
}

// GetServersByType returns a slice with all the servers of a certain type
func (sd *EtcdServiceDiscovery) GetServersByType(serverType string) []*Server {
	// TODO implement
	return nil
}

// GetServer returns a server given it's id
func (sd *EtcdServiceDiscovery) GetServer(id string) *Server {
	// TODO implement
	return nil
}

// SyncServers gets all servers from etcd
func (sd *EtcdServiceDiscovery) SyncServers() error {
	keys, err := sd.cli.Get(context.TODO(), "servers/", clientv3.WithPrefix())
	if err != nil {
		return err
	}

	for _, kv := range keys.Kvs {
		sv := string(kv.Key)
		var svData map[string]string
		err := json.Unmarshal(kv.Value, &svData)
		if err != nil {
			log.Warnf("failed to load data for server %s", sv)
		}
		splittedServer := strings.Split(sv, "/")
		if len(splittedServer) != 3 {
			log.Error("error getting server %s type and id (server name can't contain /)", sv)
			continue
		}
		svType := splittedServer[1]
		svID := splittedServer[2]
		newSv := NewServer(svID, svType, svData)
		log.Debugf("server: %s/%s loaded from etcd with data %s", newSv.Type, newSv.ID, newSv.Data)
		listSvByType, ok := sd.serverMapByType.Load(svType)
		if !ok {
			listSvByType = []*Server{}
		}
		listSvByType = append(listSvByType.([]*Server), newSv)
		sd.serverMapByType.Store(svType, listSvByType)
	}
	sd.serverMapByType.Range(func(k, v interface{}) bool {
		log.Debugf("type: %s, servers: %s", k, v)
		return true
	})

	return nil
}

// Heartbeat sends a heartbeat to etcd
func (sd *EtcdServiceDiscovery) Heartbeat() error {
	log.Debugf("renewing heartbeat with lease %s", sd.leaseID)
	_, err := sd.cli.KeepAliveOnce(context.TODO(), sd.leaseID)
	if err != nil {
		return err
	}
	return nil
}

// Stop stops the service discovery client
func (sd *EtcdServiceDiscovery) Stop() {
	defer sd.cli.Close()
	log.Warn("stopping etcd service discovery client")
}

func (sd *EtcdServiceDiscovery) watchEtcdChanges() {
	w := sd.cli.Watch(context.Background(), "servers/", clientv3.WithPrefix())

	go func(chn clientv3.WatchChan) {
		for wResp := range chn {
			for _, ev := range wResp.Events {
				log.Info("hello ev %s", ev)
			}
		}
	}(w)
}

func getKey(serverID, serverType string) string {
	return fmt.Sprintf("servers/%s/%s", serverType, serverID)
}
