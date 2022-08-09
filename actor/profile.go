package actor

import (
	"encoding/json"
	"github.com/spf13/cast"
	"github.com/wwj31/dogactor/log"
	"net/http"
	"sort"
)

func profileHttp(sys *System) {
	var mux = http.NewServeMux()
	mux.HandleFunc("/local", ActorInfo(sys))
	mux.HandleFunc("/cluster", ClusterInfo(sys))

	go func() {
		if err := http.ListenAndServe(sys.profileAddr, mux); err != nil {
			log.SysLog.Errorf("profile http err:%v", err)
		}
	}()
}

type Profile struct {
	ActorId     string
	Len         int
	Cap         int
	LastMsg     string
	ProcessTime string
}

func ActorInfo(sys *System) func(writer http.ResponseWriter, r *http.Request) {
	return func(writer http.ResponseWriter, r *http.Request) {
		var actors []Profile
		sys.actorCache.Range(func(key, value interface{}) bool {
			obj := value.(*actor)
			actors = append(actors, Profile{
				ActorId:     cast.ToString(key),
				Len:         len(obj.mailBox.ch),
				Cap:         cap(obj.mailBox.ch),
				LastMsg:     obj.mailBox.lastMsgName,
				ProcessTime: obj.mailBox.processingTime.String(),
			})
			return true
		})

		sort.SliceStable(actors, func(i, j int) bool {
			return actors[i].ActorId < actors[j].ActorId
		})

		bytes, _ := json.Marshal(actors)
		if _, err := writer.Write(bytes); err != nil {
			log.SysLog.Errorf("ActorInfo write err:%v", err)
		}
	}
}

func ClusterInfo(sys *System) func(writer http.ResponseWriter, r *http.Request) {
	return func(writer http.ResponseWriter, r *http.Request) {
		rsp, _ := sys.RequestWait(sys.cluster.ID(), "nodeinfo")
		if _, err := writer.Write([]byte(rsp.(string))); err != nil {
			log.SysLog.Errorf("ActorInfo write err:%v", err)
		}
	}
}
