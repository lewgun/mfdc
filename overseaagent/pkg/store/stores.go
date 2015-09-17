//store 实现所有应用商店的代理功能
package store

import (
	"encoding/json"
	"net/http"
	"os"

	"overseaagent/pkg/config"

	"fmt"
)

//HandlerFunc 消息处理handler
type HandlerFunc func(p *Store, w http.ResponseWriter, r *http.Request)

//UNUSED 临时注释编译错误
func UNUSED(...interface{}) {}

//g_handlerMap handler map
var g_handlerMap = map[string]HandlerFunc{
	"/":           (*Store).index,
	"/googleplay": (*Store).googlePlay,
	"/appstore":   (*Store).appStore,
}

//Store 所有商店的基础结构
type Store struct {
	config        *config.Config
	globalProxyOn bool //全局代理
}

//setProxyEnv 设置http代理
func (s *Store) setProxyEnv(typ config.StoreType) {

	if s.globalProxyOn {
		return
	}

	//设置全局代理
	if typ == config.ALL_STORE && s.config.Proxy.AllOn {
		s.globalProxyOn = true
		os.Setenv("HTTP_PROXY", s.config.Proxy.Address)
		return
	}

	//分商店设置代理
	if (typ == config.APP_STORE && s.config.Stores.AppStore.HTTPProxyOn) ||
		(typ == config.GOOGLE_PLAY && s.config.Stores.GooglePlay.HTTPProxyOn) {
		os.Setenv("HTTP_PROXY", s.config.Proxy.Address)
	}

}

//取消http代理
func (s *Store) UnsetProxyEnv() {

	if s.globalProxyOn {
		return
	}

	os.Unsetenv("HTTP_PROXY")
}

//Auth 权限认证
func (s *Store) auth(w http.ResponseWriter, req *http.Request) error {

	//todo 添加权限操作
	return nil

}

//Response 向客户端的消息响应
func (s *Store) response(w http.ResponseWriter, req *http.Request, data interface{}) {
	jsonText, err := json.Marshal(data)
	UNUSED(err)

	n, err := w.Write(jsonText)
	UNUSED(n)
}

//Response 向客户端的消息响应
func (s *Store) responseJSON(w http.ResponseWriter, req *http.Request, jsonText []byte) {
	w.Write(jsonText)
}

func (p *Store) dispatch(path string, w http.ResponseWriter, r *http.Request) {
	if fn, ok := g_handlerMap[path]; ok {
		fn(p, w, r)
	} else {
		http.NotFound(w, r)
	}

}

func (s *Store) index(w http.ResponseWriter, req *http.Request) {
	err := s.auth(w, req)
	if err != nil {
		return
	}

	//todo. do something real here
	var data interface{} = "pls send your request to /googleplay or /appstore."

	s.response(w, req, data)

}

func (p *Store) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	p.dispatch(r.URL.Path, w, r)

}

//New 创建一个store
func New(c *config.Config) *Store {
	s := &Store{
		config: c,
	}

	//若有必要 开启全局代理
	s.setProxyEnv(config.ALL_STORE)
	return s
}
