package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/wwj31/dogactor/actor/cluster/mq"
	"github.com/wwj31/dogactor/log"
	"time"
)

// https://github.com/nats-io/nats.go

func New() *Nats {
	return &Nats{
		subscribers: make(map[string]*nats.Subscription),
	}
}

var _ mq.MQ = &Nats{}

type Nats struct {
	url         string
	nc          *nats.Conn
	subscribers map[string]*nats.Subscription
}

func (n *Nats) Connect(url string) (err error) {
	n.url = url

	// Connect to a server with nats.GetDefaultOptions()
	n.nc, err = nats.Connect(n.url)
	return
}
func (n *Nats) Close() {
	n.nc.Close()
}

func (n *Nats) Pub(subj string, data []byte) error {
	return n.nc.Publish(subj, data)
}

func (n *Nats) Req(subj string, data []byte) ([]byte, error) {
	msg, err := n.nc.Request(subj, data, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return msg.Data, nil
}

func (n *Nats) SubASync(subject string, callback func(data []byte)) (err error) {
	var newSub *nats.Subscription
	newSub, err = n.nc.Subscribe(subject, func(m *nats.Msg) {
		//fmt.Printf("Received a message: %s\n", string(m.Data))
		callback(m.Data)
	})

	log.SysLog.Infow("subAsync success!", "subject", subject)
	n.subscribers[subject] = newSub
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
	sub, exist := n.subscribers[subject]
	if !exist {
		return
	}
	return sub.Unsubscribe()
}
func (n *Nats) Flush() error {
	return n.nc.Flush()
}
