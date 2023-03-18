package etcd

import (
	"context"
	"errors"
	"github.com/wwj31/dogactor/actor/cluster/fullmesh/servmesh"
	"go.etcd.io/etcd/api/v3/mvccpb"
	etcd "go.etcd.io/etcd/client/v3"
	"go.uber.org/atomic"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/wwj31/dogactor/log"
)

const (
	timeout        = 5 * time.Second //etcd连接超时
	grantTTL       = 6               //etcd timer to live
	invalidLeaseId = -1
)

type Etcd struct {
	endpoints string
	prefix    string

	etcdClient *etcd.Client
	leaseId    atomic.Int64 // 租约
	revision   int64        // 当前监听版本号
	retry      bool         // 当出现任何错误时，重新建立链接

	localActors sync.Map //
	handler     servmesh.ServMeshHander
	stop        atomic.Int32
	wg          sync.WaitGroup
}

func NewEtcd(endpoints, prefix string) *Etcd {
	nEtcd := &Etcd{
		endpoints: endpoints,
		prefix:    prefix,
	}
	nEtcd.setLeaseID(invalidLeaseId)
	return nEtcd
}

func (s *Etcd) Start(h servmesh.ServMeshHander) (err error) {
	s.handler = h

	logInfo := []interface{}{"endpoints", s.endpoints, "prefix", s.prefix}
	log.SysLog.Infow("etcd start", logInfo...)

	s.etcdClient, err = etcd.New(etcd.Config{Endpoints: strings.Split(s.endpoints, "_"), DialTimeout: timeout})
	if err != nil {
		log.SysLog.Errorw("new etcd client failed", append(logInfo, "err", err)...)
		return err
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for !s.IsStop() {
			s.run()
			time.Sleep(time.Second)
		}

		log.SysLog.Infow("etcd stop", logInfo...)
	}()
	return
}

func (s *Etcd) Stop() {
	if s.stop.CAS(0, 1) {
		s.wg.Wait()
	}
}

func (s *Etcd) IsStop() bool {
	return s.stop.Load() == 1
}

// RegisterService register node to etcd
func (s *Etcd) RegisterService(key, value string) error {
	s.localActors.Store(key, value)

	if s.IsStop() {
		s.retry = true
		return errors.New("etcd has stopped")
	}

	leaseId := s.getLeaseID()
	for leaseId == invalidLeaseId {
		time.Sleep(100 * time.Millisecond)
		leaseId = s.getLeaseID()
	}

	resp, err := s.etcdClient.Put(context.TODO(),
		path.Join(s.prefix, key), value, etcd.WithLease(leaseId))

	if err != nil {
		log.SysLog.Errorf("RegisterService etcd failed",
			"error", err,
			"revision", resp.Header.GetRevision(),
			"key", key,
			"value", value,
		)
		s.retry = true
	}
	return err
}

func (s *Etcd) UnregisterService(key string) error {
	s.localActors.Delete(key)

	if s.IsStop() {
		s.retry = true
		return errors.New("etcd has stopped")
	}

	resp, err := s.etcdClient.Delete(context.TODO(), path.Join(s.prefix, key))
	if err != nil {
		log.SysLog.Errorf("UnregisterService etcd failed",
			"error", err,
			"revision", resp.Header.GetRevision(),
			"key", key,
		)
		s.retry = true
	}
	return err
}

//////////////////////////////////////////////// inner func ///////////////////////////////////////////
func (s *Etcd) keepAlive() (<-chan *etcd.LeaseKeepAliveResponse, context.CancelFunc, bool) {
	if s.IsStop() {
		return nil, nil, false
	}

	s.setLeaseID(invalidLeaseId)

	// Grant authorization to obtain a lease
	ctx, _ := context.WithTimeout(context.TODO(), timeout)
	lease, err := s.etcdClient.Grant(ctx, grantTTL)
	if err != nil {
		log.SysLog.Errorw("etcd keepAlive create lease failed", "err", err)
		return nil, nil, false
	}

	ctx, cancelAlive := context.WithCancel(context.TODO())
	alive, err := s.etcdClient.KeepAlive(ctx, lease.ID)
	if err != nil {
		log.SysLog.Errorw("etcd keepAlive failed", "err", err)
		return alive, cancelAlive, false
	}

	s.setLeaseID(lease.ID)

	s.retry = false
	log.SysLog.Infow("etcd keepAlive success!", "lease", s.getLeaseID())
	return alive, cancelAlive, true
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
	for _, kv := range resp.Kvs {
		key, val := s.shiftStruct(kv)
		s.handler.OnNewServ(key, val, true)
	}
}

//把本地所有actor注册到etcd上
func (s *Etcd) syncLocalToEtcd() (err error) {
	s.localActors.Range(func(key, value interface{}) bool {
		if err = s.RegisterService(key.(string), value.(string)); err != nil {
			return false
		}
		return true
	})
	return
}

func (s *Etcd) run() {
	alive, cancelAlive, ok := s.keepAlive()
	if !ok {
		return
	}
	defer cancelAlive()

	if err := s.syncLocalToEtcd(); err != nil {
		log.SysLog.Errorw("failed to synchronize local node to etcd.", "err", err)
		return
	}

	ctx, cancelWatch := context.WithCancel(context.TODO())
	defer cancelWatch()
	watch := s.etcdClient.Watch(ctx, s.prefix, etcd.WithPrefix(), etcd.WithPrevKV())

	s.fetchNode()

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if s.retry || s.IsStop() {
				return
			}
		case resp := <-alive:
			if resp == nil {
				log.SysLog.Warnw("etcd keepalive response a nil")
				return
			}
		case watchResp := <-watch:
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
				key, val := s.shiftStruct(e.Kv)
				log.SysLog.Infow("watch etcd", "actorId", key, "revision", revision, "put", e.Type == etcd.EventTypePut)
				s.handler.OnNewServ(key, val, e.Type == etcd.EventTypePut)
			}
		}
	}
}

func (s *Etcd) shiftStruct(kv *mvccpb.KeyValue) (k, v string) {
	_, k = path.Split(string(kv.Key))
	v = string(kv.Value)
	return
}
