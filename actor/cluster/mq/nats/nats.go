package nats

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/wwj31/dogactor/actor/cluster/mq"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
	"time"
)

// https://github.com/nats-io/nats.go

func New() *Nats {
	return &Nats{
		cancelSubs: make(map[string]func() error),
	}
}

var _ mq.MQ = &Nats{}

type Nats struct {
	url        string
	nc         *nats.Conn
	js         nats.JetStreamContext
	cancelSubs map[string]func() error
}

func (n *Nats) Connect(url string) (err error) {
	n.url = url

	// Connect to a server with nats.GetDefaultOptions()
	if n.nc, err = nats.Connect(n.url); err != nil {
		return
	}
	if n.js, err = n.nc.JetStream(); err != nil {
		return
	}
	return
}
func (n *Nats) Close() {
	n.nc.Close()
}

func (n *Nats) Pub(subj string, data []byte) error {
	return n.nc.Publish(subj+".msg", data)
}

func (n *Nats) Req(subj string, data []byte) ([]byte, error) {
	msg, err := n.nc.Request(subj, data, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return msg.Data, nil
}

func (n *Nats) SubASync(subject string, callback func(data []byte)) (err error) {
	if err = n.addStream(subject); err != nil {
		return
	}

	sub, subErr := n.js.PullSubscribe("", "consumer:"+subject, nats.BindStream(subject))
	if subErr != nil {
		return subErr
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			tctx, _ := context.WithTimeout(ctx, 10*time.Second)
			msgs, err := sub.Fetch(100, nats.Context(tctx))
			if err != nil {
				log.SysLog.Errorw("subscribe fetch got err", "subject", subject, "err", err)
			}

			select {
			case <-ctx.Done():
				log.SysLog.Infow("exit the pull loop", "subject", subject)
				return
			default:
			}

			tools.Try(func() {
				for _, msg := range msgs {
					callback(msg.Data)
					if err := msg.Ack(); err != nil {
						log.SysLog.Errorf("msg ack failed ", "subject", subject, "err", err)
					}
				}
			})
		}
	}()

	log.SysLog.Infow("subAsync success!", "subject", subject)
	n.cancelSubs[subject] = func() error {
		cancel()
		return sub.Unsubscribe()
	}
	return
}

func (n *Nats) SubSync(subject string) ([]byte, error) {
	sub, err := n.nc.SubscribeSync(subject)
	if err != nil {
		return nil, err
	}

	m, err := sub.NextMsg(5 * time.Minute)
	if err != nil {
		return nil, err
	}
	return m.Data, nil
}

func (n *Nats) UnSub(subject string) (err error) {
	cancel, exist := n.cancelSubs[subject]
	if !exist {
		return
	}
	return cancel()
}
func (n *Nats) Flush() error {
	return n.nc.Flush()
}

func (n *Nats) addStream(id string) error {
	// ### Creating the stream
	// Define the stream configuration, specifying `WorkQueuePolicy` for
	// retention, and create the stream.
	cfg := &nats.StreamConfig{
		Name:      id,
		Retention: nats.WorkQueuePolicy,
		Subjects:  []string{id + ".>"},
		Storage:   nats.MemoryStorage,
	}

	if _, err := n.js.AddStream(cfg); err != nil {
		if jsErr, ok := err.(nats.JetStreamError); !ok || jsErr.APIError().ErrorCode != nats.JSErrCodeStreamNameInUse {
			return err
		}
	}
	return nil
}
