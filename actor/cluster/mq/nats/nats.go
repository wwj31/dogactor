package nats

import (
	"context"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/wwj31/dogactor/actor/cluster/mq"
	"github.com/wwj31/dogactor/log"
	"github.com/wwj31/dogactor/tools"
	"time"
)

// https://github.com/nats-io/nats.go

func New() *Nats {
	ctx, cancel := context.WithCancel(context.Background())
	return &Nats{
		subs:   make(map[string]*SubInfo),
		ctx:    ctx,
		cancel: cancel,
	}
}

var _ mq.MQ = &Nats{}

type SubInfo struct {
	cancel context.CancelFunc
	sub    *nats.Subscription
	exit   chan struct{}
}
type Nats struct {
	ctx    context.Context
	cancel context.CancelFunc

	url  string
	nc   *nats.Conn
	js   nats.JetStreamContext
	subs map[string]*SubInfo
}

func (n *Nats) Connect(url string) (err error) {
	n.url = url

	opts := []nats.Option{
		nats.ReconnectWait(3 * time.Second),
		nats.MaxReconnects(300),
		nats.ReconnectHandler(func(conn *nats.Conn) {
			log.SysLog.Errorw("nats reconnect to server", "url", conn.ConnectedUrl())
		}),

		nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
			if err != nil {
				log.SysLog.Errorw("nats disconnect", "url", conn.ConnectedUrl(), "err", err)
			}
		}),

		nats.ClosedHandler(func(conn *nats.Conn) {
			log.SysLog.Infow("nats connect closed successfully", "url", conn.ConnectedUrl())
		}),
	}
	// Connect to a server with nats.GetDefaultOptions()
	if n.nc, err = nats.Connect(n.url, opts...); err != nil {
		return
	}
	if n.js, err = n.nc.JetStream(); err != nil {
		return
	}
	return
}

func (n *Nats) Close() {
	n.cancel()
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

	sub, subErr := n.js.PullSubscribe(
		"",
		"consumer:"+subject,
		nats.BindStream(subject),
		nats.AckAll())
	if subErr != nil {
		return subErr
	}

	ctx, cancel := context.WithCancel(n.ctx)

	exit := make(chan struct{})
	n.subs[subject] = &SubInfo{
		cancel: cancel,
		sub:    sub,
		exit:   exit,
	}

	go func() {
		defer func() {
			select {
			case exit <- struct{}{}:
			default:
			}
		}()

		for {
			deadline, deadlineCancel := context.WithTimeout(ctx, time.Minute)
			msgs, err := sub.Fetch(100, nats.Context(deadline))
			deadlineCancel()

			if err != nil {
				if err == context.Canceled {
					log.SysLog.Infow("exit the pull loop", "subject", subject)
					return
				}

				if err != context.DeadlineExceeded {
					log.SysLog.Errorw("subscribe fetch got err", "subject", subject, "err", err)
				}
			}

			tools.Try(func() {
				if len(msgs) > 0 {
					if err := msgs[len(msgs)-1].Ack(); err != nil {
						log.SysLog.Errorf("msg ack failed ", "subject", subject, "err", err)
					}
				}

				for _, msg := range msgs {
					callback(msg.Data)
				}
			})
		}
	}()

	log.SysLog.Infow("subAsync success!", "subject", subject)
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

func (n *Nats) UnSub(subject string, drained bool) (err error) {
	sub, exist := n.subs[subject]
	if !exist {
		return
	}
	delete(n.subs, subject)

	sub.cancel()

	select {
	case <-sub.exit:
	case <-time.After(10 * time.Second):
		return fmt.Errorf("nats go fetch exit timeout ")
	}

	err = sub.sub.Unsubscribe()
	if err != nil {
		return
	}

	if !drained {
		err = n.js.DeleteStream(subject)
		if err != nil {
			return
		}
	}

	return
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
		Retention: nats.LimitsPolicy,
		Subjects:  []string{id + ".>"},
		Storage:   nats.MemoryStorage,
		MaxAge:    time.Minute,
	}

	if _, err := n.js.AddStream(cfg); err != nil {
		if jsErr, ok := err.(nats.JetStreamError); !ok || jsErr.APIError().ErrorCode != nats.JSErrCodeStreamNameInUse {
			return err
		}
	}
	return nil
}
