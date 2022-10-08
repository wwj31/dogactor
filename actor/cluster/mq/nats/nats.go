package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/wwj31/dogactor/actor/cluster/mq"
	"time"
)

// https://github.com/nats-io/nats.go

func New() *Nats {
	return &Nats{
		subscribers: make(map[string]*nats.Subscription),
	}
}

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

func (n *Nats) Pub(msg interface{}) {
	// todo .......
}

func (n *Nats) SubASync(subject string, callback func(msg mq.MSG)) (err error) {
	var newSub *nats.Subscription
	newSub, err = n.nc.Subscribe(subject, func(m *nats.Msg) {
		//fmt.Printf("Received a message: %s\n", string(m.Data))
		callback(m.Data)
	})

	n.subscribers[subject] = newSub
	return
}

func (n *Nats) SubSync(subject string) (mq.MSG, error) {
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
