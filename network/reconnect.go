package network

type tcpReconnectHandler struct {
	reconnect chan struct{}
}

func (s *tcpReconnectHandler) OnSessionClosed() {
	s.reconnect <- struct{}{}
}
func (s *tcpReconnectHandler) OnSessionCreated(NetSession) {}
func (s *tcpReconnectHandler) OnRecv([]byte)               {}
