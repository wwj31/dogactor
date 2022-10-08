package mq

type MSG interface{}

type MQ interface {
	Connect(url string) error
	Pub(msg interface{})
	SubASync(subject string, callback func(msg MSG)) error
	SubSync(subject string) (MSG, error)
	UnSub(subject string) (err error)
}
