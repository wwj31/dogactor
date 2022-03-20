package tools

import (
	"github.com/rs/xid"
	"github.com/satori/go.uuid"
)

func UUID() string {
	return uuid.NewV4().String()
}

func XUID() string {
	return xid.New().String()
}
