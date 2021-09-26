package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/wwj31/dogactor/log"
)

type HttpHandler struct {
	HanderMap map[string]http.Handler
}

func (h *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := time.Now()
	defer func() {
		log.KV("url", r.URL.String()).KV("from", r.RemoteAddr).KV("cost", time.Now().Sub(c)).Debug("http request")
	}()

	r.ParseForm()
	log.KV("url", r.URL.String()).KV("from", r.RemoteAddr).KV("Content-Type", r.Header.Get("Content-Type")).Debug("http request")

	Try(func() {
		if method := h.HanderMap[r.URL.Path]; method != nil {
			method.ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}
	}, func(ex interface{}) {
		http.Error(w, "server error", http.StatusInternalServerError)
	})
}

func HttpGet(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: time.Second * 3,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		err = fmt.Errorf("resp is nil")
		return nil, err
	}

	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	return result, err
}

func HttpReq(r *http.Request) *http.Response {
	if r == nil {
		return nil
	}

	client := &http.Client{
		Timeout: time.Second * 3,
	}

	response, err := client.Do(r)
	if err != nil {
		log.KV("error", err).ErrorStack(2, "httpreq error")
		return nil
	}
	return response
}

func HttpPost(addr string, body []byte) (result []byte, err error) {
	bodyReader := bytes.NewReader(body)
	client := &http.Client{
		Timeout: time.Second * 3,
	}

	//http_proxy := iniconfig.String("http_proxy")
	//if http_proxy != "" {
	//	urlproxy, _ := url.Parse("http_proxy")
	//	client.Transport = &http.Transport{Proxy: http.ProxyURL(urlproxy)}
	//}

	resp, err := client.Post(addr, "application/json", bodyReader)
	if err != nil {
		return
	}
	if resp == nil {
		err = fmt.Errorf("resp is nil")
		return
	}
	defer resp.Body.Close()

	result, err = ioutil.ReadAll(resp.Body)
	return
}

func HttpPostForm(url string, data url.Values) (result []byte, err error) {
	client := &http.Client{Timeout: time.Second * 6}

	resp, err := client.PostForm(url, data)
	if err != nil {
		return
	}
	if resp == nil {
		err = fmt.Errorf("resp is nil")
		return
	}
	defer resp.Body.Close()

	result, err = ioutil.ReadAll(resp.Body)
	return
}

func HttpTransmit(w http.ResponseWriter, r *http.Request, remote string) {
	if r == nil {
		return
	}

	remoter, err := http.NewRequest(r.Method, fmt.Sprintf("http://%v%v", remote, r.URL.Path), r.Body)
	if err != nil {
		log.KV("error", err).Error("HttpTransmit actorerr")
		return
	}

	remoter.Header = r.Header

	response := HttpReq(remoter)
	if response == nil {
		return
	}

	defer response.Body.Close()
	for k, v := range response.Header {
		w.Header().Set(k, v[0])
	}
	io.Copy(w, response.Body)
}

func HttpUnmarshalBody(r *http.Request, data interface{}) (body []byte, ok bool) {
	defer r.Body.Close()
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.KV("actorerr", err).WarnStack(2, "http body read actorerr")
		return
	}

	err = json.Unmarshal(content, data)
	if err != nil {
		log.KV("actorerr", err).KV("content", string(content)).WarnStack(2, "http body json unmarshal actorerr")
		return content, false
	}
	return content, true
}

func HttpResponse(w http.ResponseWriter, data interface{}) (ok bool) {
	content, err := json.Marshal(data)
	if err != nil {
		log.KV("error", err).ErrorStack(2, "json marsh failed")
		return
	}
	if _, err := w.Write(content); err != nil {
		log.KVs(log.Fields{"content": string(content), "error": err}).ErrorStack(2, "write http error")
		return
	}
	return true
}

func HttpRespCode(w http.ResponseWriter, code int) {
	type Code struct {
		Code int
	}
	HttpResponse(w, &Code{Code: code})
}
