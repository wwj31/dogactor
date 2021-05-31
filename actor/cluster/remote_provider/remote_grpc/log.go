package remote_grpc

import (
	"fmt"
	"strings"

	"github.com/wwj31/godactor/log"
)

var logger = log.New(log.TAG_DEBUG_I)

func (s *RemoteMgr) remoteinfo(param ...string) {
	sessions := []string{}
	clients := []string{}

	s.sessions.IterCb(func(key string, v interface{}) {
		sessions = append(sessions, fmt.Sprintf("[host=%v sessionId=%v]", key, v.(*remoteHandler).Id()))
	})
	session := "\n--------------------------------- session ---------------------------------\n%s\n--------------------------------- session ---------------------------------"
	logger.Info(fmt.Sprintf(session, strings.Join(sessions, "\n")))

	s.clients.IterCb(func(key string, v interface{}) {
		clients = append(clients, fmt.Sprintf("[host=%v]", key))
	})
	client := "\n--------------------------------- client ---------------------------------\n%s\n--------------------------------- client ---------------------------------"
	logger.Info(fmt.Sprintf(client, strings.Join(clients, "\n")))
}
