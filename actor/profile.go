package actor

import (
	"fmt"
	"github.com/liushuochen/gotable"
	"github.com/spf13/cast"
	"github.com/wwj31/dogactor/log"
	"net/http"
	"sort"
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

		table, err := gotable.Create("actorId", "len", "cap", "last msg", "process time")
		if err != nil {
			errStr := fmt.Sprintf("Create table failed: %v", err.Error())
			log.SysLog.Errorf(errStr)
			_, _ = writer.Write([]byte(errStr))
			return
		}

		var actors []map[string]string
		sys.actorCache.Range(func(key, value interface{}) bool {
			obj := value.(*actor)
			actors = append(actors, map[string]string{
				"actorId":      cast.ToString(key),
				"len":          cast.ToString(len(obj.mailBox.ch)),
				"cap":          cast.ToString(cap(obj.mailBox.ch)),
				"last msg":     obj.mailBox.lastMsgName,
				"process time": obj.mailBox.processingTime.String(),
			})
			return true
		})

		sort.SliceStable(actors, func(i, j int) bool {
			return actors[i]["actorId"] < actors[j]["actorId"]
		})

		table.AddRows(actors)

		str := table.String()
		if _, err := writer.Write([]byte(str)); err != nil {
			log.SysLog.Errorf("ActorInfo write err:%v", err)
		}
	}
}
