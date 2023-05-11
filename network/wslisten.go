package network

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"

	"github.com/wwj31/dogactor/log"
)

func StartWSListen(addr string, newCodec func() DecodeEncoder, newHandler func() SessionHandler) Listener {
	l := &WebSocketListener{
		addr:       addr,
		newCodec:   newCodec,
		newHandler: newHandler,
	}
	return l
}

type WebSocketListener struct {
	addr       string
	ctx        context.Context
	cancel     context.CancelFunc
	newCodec   func() DecodeEncoder
	newHandler func() SessionHandler
}

func (w *WebSocketListener) Start(exceptPort ...int) error {
	go func() {
		w.ctx, w.cancel = context.WithCancel(context.Background())
		log.SysLog.Infow("ws listen", "addr", w.addr)
		err := http.ListenAndServe(w.addr, http.HandlerFunc(w.msg))
		if err != nil {
			log.SysLog.Errorw("web socket stop err", "err", err)
		}
	}()
	return nil
}

func (w *WebSocketListener) Stop() {
	w.cancel()
}

func (w *WebSocketListener) Port() int {
	str := strings.Split(w.addr, ":")
	port, _ := strconv.Atoi(str[len(str)-1])
	return port
}

func (w *WebSocketListener) msg(wt http.ResponseWriter, r *http.Request) {
	defaultUpgrade := websocket.Upgrader{}
	conn, err := defaultUpgrade.Upgrade(wt, r, nil)
	if err != nil {
		log.SysLog.Errorw("ws upgrade failed ", "err", err)
		return
	}

	defer conn.Close()
	session := newWSSession(conn, w.newCodec(), w.newHandler())
	select {
	case <-session.start().Done():
	case <-w.ctx.Done():
	}
}
