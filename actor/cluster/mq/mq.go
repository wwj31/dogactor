package mq

type MQ interface {
	Connect(url string) error
	Close()
	Pub(subject string, data []byte) error
	Req(subject string, data []byte) ([]byte, error)
	SubASync(subject string, callback func(data []byte)) error
	UnSub(subject string) (err error)
}
