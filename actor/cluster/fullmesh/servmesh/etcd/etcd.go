package etcd

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	etcd "go.etcd.io/etcd/client/v3"
	"go.uber.org/atomic"

	"github.com/wwj31/dogactor/actor/cluster/fullmesh/servmesh"

	"github.com/wwj31/dogactor/log"
)

const (
	grantTTL       = 2 //etcd timer to live (s)
	invalidLeaseId = -1
)

type Etcd struct {
	endpoints string
	prefix    string

	ctx    context.Context
	cancel context.CancelFunc

	etcdClient *etcd.Client
	leaseId    atomic.Int64 // 租约
	revision   int64        // 当前监听版本号

	localActors sync.Map //
	handler     servmesh.MeshHandler
	stop        atomic.Int32
}

func NewEtcd(endpoints, prefix string) *Etcd {
	nEtcd := &Etcd{
		endpoints: endpoints,
		prefix:    prefix,
	}
	nEtcd.setLeaseID(invalidLeaseId)
	return nEtcd
}

func (s *Etcd) Start(h servmesh.MeshHandler) (err error) {
	s.handler = h

	logInfo := []interface{}{"endpoints", s.endpoints, "prefix", s.prefix}
	log.SysLog.Infow("etcd start", logInfo...)

	s.etcdClient, err = etcd.New(etcd.Config{Endpoints: strings.Split(s.endpoints, "_"), DialTimeout: 20 * time.Second})
	if err != nil {
		log.SysLog.Errorw("new etcd client failed", append(logInfo, "err", err)...)
		return err
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())

	go func() {
		for !s.IsStop() {
			alive, cancelAlive := s.createLeaseAndKeepAlive()
			watcher, cancel := s.watch()

			// pull all node from etcd
			s.fetchNode()

			s.run(watcher, cancel, alive, cancelAlive)
			time.Sleep(time.Second)

			if err := s.syncLocalToEtcd(); err != nil {
				log.SysLog.Errorw("failed to synchronize local node to etcd.", "err", err)
				return
			}

		}
	}()
	return
}

func (s *Etcd) Stop() {
	if s.stop.CAS(0, 1) {
		s.cancel()
	}
}

func (s *Etcd) IsStop() bool {
	return s.stop.Load() == 1
}

// RegisterService register node to etcd
func (s *Etcd) RegisterService(key, value string) error {
	s.localActors.Store(key, value)

	if s.IsStop() {
		return errors.New("etcd has stopped")
	}

	leaseId := s.getLeaseID()
	for leaseId == invalidLeaseId {
		time.Sleep(100 * time.Millisecond)
		leaseId = s.getLeaseID()
	}

	_, err := s.etcdClient.Put(context.TODO(),
		path.Join(s.prefix, key), value, etcd.WithLease(leaseId))

	if err != nil {
		log.SysLog.Errorf("RegisterService etcd failed",
			"error", err,
			"key", key,
			"value", value,
		)
	}
	return err
}

func (s *Etcd) UnregisterService(key string) error {
	s.localActors.Delete(key)

	if s.IsStop() {
		return errors.New("etcd has stopped")
	}

	_, err := s.etcdClient.Delete(context.TODO(), path.Join(s.prefix, key))
	if err != nil {
		log.SysLog.Errorf("UnregisterService etcd failed",
			"error", err,
			"key", key,
		)
	}
	return err
}

func (s *Etcd) Get(key string) (val string, err error) {
	if s.IsStop() {
		return "", errors.New("etcd has stopped")
	}

	var result *etcd.GetResponse
	result, err = s.etcdClient.Get(context.TODO(), path.Join(s.prefix, key))
	if len(result.Kvs) > 0 {
		_, val = s.shiftStruct(result.Kvs[0])
	}

	return
}

// ////////////////////////////////////////////// inner func ///////////////////////////////////////////
func (s *Etcd) createLeaseAndKeepAlive() (<-chan *etcd.LeaseKeepAliveResponse, context.CancelFunc) {
	s.setLeaseID(invalidLeaseId)

	// Grant authorization to obtain a lease
	lease, err := s.etcdClient.Grant(s.ctx, grantTTL)
	if err != nil {
		log.SysLog.Errorw("etcd keepAlive create lease failed", "err", err)
		return nil, nil
	}

	ctx, cancelAlive := context.WithCancel(s.ctx)
	alive, err := s.etcdClient.KeepAlive(ctx, lease.ID)
	if err != nil {
		log.SysLog.Errorw("etcd keepAlive failed", "err", err)
		return alive, cancelAlive
	}

	s.setLeaseID(lease.ID)
	return alive, cancelAlive
}

func (s *Etcd) getLeaseID() etcd.LeaseID {
	return (etcd.LeaseID)(s.leaseId.Load())
}

func (s *Etcd) setLeaseID(leaseID etcd.LeaseID) {
	s.leaseId.Store((int64)(leaseID))
}

func (s *Etcd) fetchNode() {
	resp, err := etcd.NewKV(s.etcdClient).Get(context.Background(), s.prefix, etcd.WithPrefix())
	if err != nil {
		return
	}
	var nodes []string
	for _, kv := range resp.Kvs {
		key, val := s.shiftStruct(kv)
		s.handler.OnNewServ(key, val, true)
		nodes = append(nodes, fmt.Sprintf("%v:%v", key, val))
	}
	log.SysLog.Infow("etcd fetch all nodes", "nodes", nodes)
}

// 把本地所有actor注册到etcd上
func (s *Etcd) syncLocalToEtcd() (err error) {
	s.localActors.Range(func(key, value interface{}) bool {
		if err = s.RegisterService(key.(string), value.(string)); err != nil {
			return false
		}
		return true
	})
	return
}

func (s *Etcd) watch() (etcd.WatchChan, context.CancelFunc) {
	ctx, cancel := context.WithCancel(s.ctx)
	return s.etcdClient.Watch(ctx, s.prefix, etcd.WithPrefix(), etcd.WithPrevKV()), cancel
}

func (s *Etcd) run(watcher etcd.WatchChan, watcherCancel context.CancelFunc,
	alive <-chan *etcd.LeaseKeepAliveResponse, cancelAlive context.CancelFunc) {
	defer cancelAlive()
	defer watcherCancel()

	for {
		select {
		case <-s.ctx.Done():
			log.SysLog.Infow("etcd stop")
			return

		case resp := <-alive:
			if resp == nil {
				log.SysLog.Warnw("etcd keepalive response a nil")
				return
			}

		case watchResp := <-watcher:
			if err := watchResp.Err(); err != nil {
				log.SysLog.Errorw("watch etcd error", "error", err)
				return
			}

			revision := watchResp.Header.GetRevision()
			if s.revision > revision {
				log.SysLog.Warnw("watch etcd revision", "last", s.revision, "revision", revision)
				break
			}
			s.revision = revision

			for _, e := range watchResp.Events {
				var kv *mvccpb.KeyValue
				if e.Type == etcd.EventTypePut {
					kv = e.Kv
				} else {
					kv = e.PrevKv
				}

				actorId, host := s.shiftStruct(kv)
				if _, load := s.localActors.Load(actorId); load {
					continue
				}
				s.handler.OnNewServ(actorId, host, e.Type == etcd.EventTypePut)
			}
		}
	}

}

func (s *Etcd) shiftStruct(kv *mvccpb.KeyValue) (k, v string) {
	_, k = path.Split(string(kv.Key))
	v = string(kv.Value)
	return
}
