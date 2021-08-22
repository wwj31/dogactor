package etcd

import (
	"context"
	"errors"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.uber.org/atomic"
	"strings"
	"sync"
	"time"

	"github.com/wwj31/dogactor/actor"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
)

const (
	ETCD_TIMEOUT   = 5 * time.Second //etcd连接超时
	ETCD_GRANT_TTL = 6               //etcd jtimer to live
	InValidLeaseId = -1
)

type Etcd struct {
	endpoints string
	prefix    string

	etcdCliet *clientv3.Client
	leaseId   atomic.Int64 // 租约
	revision  int64        // 当前监听版本号
	registErr error        // 注册错误(当任意注册发生错误，则全部推倒重来)

	localActors sync.Map //
	actorSystem *actor.System
	stop        atomic.Int32
	wg          sync.WaitGroup
}

func NewEtcd(endpoints, prefix string) *Etcd {
	etcd := &Etcd{
		endpoints: endpoints,
		prefix:    prefix,
	}
	etcd.setLeaseID(InValidLeaseId)
	return etcd
}

// 初始化并启动etcd本地服务
// example1: etcd := newEtcd(....).Start()
func (s *Etcd) Start(actorSystem *actor.System) error {
	s.actorSystem = actorSystem

	logger.KV("endpoints", s.endpoints).KV("prefix", s.prefix).Info("etcd start")

	client, err := clientv3.New(clientv3.Config{Endpoints: strings.Split(s.endpoints, "_"), DialTimeout: ETCD_TIMEOUT})
	if err != nil {
		log.KV("err", err).Error("new etcd client failed")
		return err
	}

	s.etcdCliet = client
	s.wg.Add(1)
	tools.GoEngine(func() {
		defer s.wg.Done()
		for {
			if s.IsStop() {
				logger.KV("endpoints", s.endpoints).KV("prefix", s.prefix).Info("etcd stop")
				return
			}
			s.run()
			time.Sleep(time.Second)
		}
	})
	return nil
}

func (s *Etcd) Stop() {
	if s.stop.CAS(0, 1) {
		s.wg.Wait()
	}
}

func (s *Etcd) IsStop() bool {
	return s.stop.Load() == 1
}

// 注册kv
func (s *Etcd) RegistService(key, value string) error {
	s.localActors.Store(key, value)
	if s.IsStop() {
		s.registErr = errors.New("etcd has stoped")
		return s.registErr
	}

	leaseId := s.getLeaseID()
	if leaseId == InValidLeaseId {
		return nil
	}

	resp, err := s.etcdCliet.Put(context.TODO(), s.prefix+key, value, clientv3.WithLease(leaseId))
	if err != nil {
		logger.KV("error", err).KV("revision", resp.Header.GetRevision()).KV("key", key).KV("value", value).Error("put etcd failed")
		s.registErr = err
	}
	return err
}

//////////////////////////////////////////////// inner func ///////////////////////////////////////////
func (s *Etcd) keepAlive() (<-chan *clientv3.LeaseKeepAliveResponse, context.CancelFunc, clientv3.WatchChan, context.CancelFunc, bool) {
	if s.IsStop() {
		return nil, nil, nil, nil, false
	}

	logger.Info("etcd keepAlive start")

	s.setLeaseID(InValidLeaseId)

	ctx, _ := context.WithTimeout(context.TODO(), ETCD_TIMEOUT)
	lease, err := s.etcdCliet.Grant(ctx, ETCD_GRANT_TTL)
	if err != nil {
		logger.KV("err", err).Error("etcd keepAlive create lease failed")
		return nil, nil, nil, nil, false
	}

	ctx, cancelAlive := context.WithCancel(context.TODO())
	alive, err := s.etcdCliet.KeepAlive(ctx, lease.ID)
	if err != nil {
		logger.KV("err", err).Error("etcd keepAlive failed")
		return alive, cancelAlive, nil, nil, false
	}

	s.setLeaseID(lease.ID)

	s.registErr = nil
	if err := s.syncLocalToEtcd(); err != nil {
		logger.KV("err", err).Error("etcd syncLocalToEtcd failed")
		return alive, cancelAlive, nil, nil, false
	}

	ctx, cancelWatch := context.WithCancel(context.TODO())
	watch := s.etcdCliet.Watch(ctx, s.prefix, clientv3.WithPrefix(), clientv3.WithPrevKV())
	s.initAlreadyInEtcd()

	logger.KV("lease", s.getLeaseID()).Info("etcd keepAlive success!")
	return alive, cancelAlive, watch, cancelWatch, true
}

func (s *Etcd) getLeaseID() clientv3.LeaseID {
	return (clientv3.LeaseID)(s.leaseId.Load())
}

func (s *Etcd) setLeaseID(leaseID clientv3.LeaseID) {
	s.leaseId.Store((int64)(leaseID))
}

func (s *Etcd) initAlreadyInEtcd() {
	resp, err := clientv3.NewKV(s.etcdCliet).Get(context.Background(), s.prefix, clientv3.WithPrefix())
	if err != nil {
		return
	}
	for _, kv := range resp.Kvs {
		key, val := s.shiftStruct(kv)
		s.actorSystem.DispatchEvent("$etcd", &actor.Ev_clusterUpdate{ActorId: key, Host: val, Add: true})
	}
}

//把本地所有actor注册到etcd上
func (s *Etcd) syncLocalToEtcd() (err error) {
	s.localActors.Range(func(key, value interface{}) bool {
		if err = s.RegistService(key.(string), value.(string)); err != nil {
			return false
		}
		return true
	})
	return
}

func (s *Etcd) run() {
	alive, cancelAlive, watch, cancelWatch, ok := s.keepAlive()
	defer func() {
		if cancelAlive != nil {
			cancelAlive()
		}
		if cancelWatch != nil {
			cancelWatch()
		}
	}()

	if !ok {
		return
	}

	ticker := time.NewTicker(time.Second * 5) //间隔检查注册错误
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if s.registErr != nil || s.IsStop() {
				return
			}
		case resp := <-alive:
			if resp == nil {
				return
			}
		case watchResp := <-watch:
			if err := watchResp.Err(); err != nil {
				log.KV("error", err).Debug("watch etcd error")
				return
			}

			revision := watchResp.Header.GetRevision()
			if s.revision > revision {
				log.KV("last", s.revision).KV("revision", revision).Debug("watch etcd revision")
				break
			}
			s.revision = revision

			for _, e := range watchResp.Events {
				key, val := s.shiftStruct(e.Kv)
				log.KV("actorId", key).KV("revision", revision).KV("put", e.Type == clientv3.EventTypePut).Debug("watch etcd")
				s.actorSystem.DispatchEvent("$etcd", &actor.Ev_clusterUpdate{ActorId: key, Host: val, Add: e.Type == clientv3.EventTypePut})
			}
		}
	}
}

func (s *Etcd) shiftStruct(kv *mvccpb.KeyValue) (k, v string) {
	k = string(kv.Key)
	v = string(kv.Value)
	k = strings.Replace(k, s.prefix, "", 1)
	return
}
