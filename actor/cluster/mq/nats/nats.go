package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/wwj31/dogactor/actor/cluster/mq"
	"time"
)

// https://github.com/nats-io/nats.go

func New() *Nats {
	return &Nats{}
}

type Nats struct {
	url string
	nc  *nats.Conn
}

func (n Nats) Connect(url string) (err error) {
	n.url = url

	// Connect to a server with nats.GetDefaultOptions()
	n.nc, err = nats.Connect(n.url)
	return
}

func (n Nats) Pub(msg interface{}) {
}

func (n Nats) SubASync(subject string, callback func(msg mq.MSG)) error {
	_, err := n.nc.Subscribe(subject, func(m *nats.Msg) {
		//fmt.Printf("Received a message: %s\n", string(m.Data))
		callback(m.Data)
	})
	return err
}

func (n Nats) SubSync(subject string) (mq.MSG, error) {
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
