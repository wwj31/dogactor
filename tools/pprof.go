package tools

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"runtime"
	"time"

	"github.com/wwj31/dogactor/log"
)

func PProfInit(port int32) {
	go Try(func() {
		addr := fmt.Sprintf("0.0.0.0:%v", port)
		s := &http.Server{
			Addr:         addr,
			Handler:      &HttpHandler{HanderMap: pprofHandlers()},
			ReadTimeout:  3 * time.Minute,
			WriteTimeout: 3 * time.Minute,
		}
		log.SysLog.Infow("pprof start", "addr", addr)
		err := s.ListenAndServe()
		if err != nil {
			log.SysLog.Errorw("pprof start", "err", err)
		}
	})
}

const HTTPPrefixPProf = "/debug/pprof"

// PProfHandlers returns a map of pprof handlers keyed by the HTTP path.
func pprofHandlers() map[string]http.Handler {
	// set only when there's no existing setting
	if runtime.SetMutexProfileFraction(-1) == 0 {
		// 1 out of 5 mutex observe are reported, on average
		runtime.SetMutexProfileFraction(5)
	}

	m := make(map[string]http.Handler)

	m[HTTPPrefixPProf+"/"] = http.HandlerFunc(pprof.Index)
	m[HTTPPrefixPProf+"/allocs"] = pprof.Handler("allocs")
	m[HTTPPrefixPProf+"/block"] = pprof.Handler("block")
	m[HTTPPrefixPProf+"/cmdline"] = http.HandlerFunc(pprof.Cmdline)
	m[HTTPPrefixPProf+"/goroutine"] = pprof.Handler("goroutine")
	m[HTTPPrefixPProf+"/heap"] = pprof.Handler("heap")
	m[HTTPPrefixPProf+"/mutex"] = pprof.Handler("mutex")
	m[HTTPPrefixPProf+"/profile"] = http.HandlerFunc(pprof.Profile)
	m[HTTPPrefixPProf+"/threadcreate"] = pprof.Handler("threadcreate")
	m[HTTPPrefixPProf+"/trace "] = http.HandlerFunc(pprof.Trace)
	m[HTTPPrefixPProf+"/symbol"] = http.HandlerFunc(pprof.Symbol)
	return m
}
