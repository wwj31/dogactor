package rank

import "fmt"

type ErrorUnMarshalRead struct {
	err error
	n   int
}

func (s ErrorUnMarshalRead) Error() string {
	return fmt.Sprintf("read error n:%v err:%v", s.n, s.err)
}
