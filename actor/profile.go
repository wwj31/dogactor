package actor

import (
	"github.com/wwj31/dogactor/log"
	"net/http"
)

func profileHttp(sys *System) {
	var mux = http.NewServeMux()
	mux.HandleFunc("/actorinfo", ActorInfo(sys))

	go func() {
		if err := http.ListenAndServe(sys.profileAddr, mux); err != nil {
			log.SysLog.Errorf("profile http err:%v", err)
		}
	}()
}

func ActorInfo(sys *System) func(writer http.ResponseWriter, r *http.Request) {
	return func(writer http.ResponseWriter, r *http.Request) {
		info := sys.actorInfo()
		if _, err := writer.Write([]byte(info)); err != nil {
			log.SysLog.Errorf("ActorInfo write err:%v", err)
		}
	}
}
