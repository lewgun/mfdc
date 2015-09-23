//store 实现所有应用商店的代理功能
package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"overseaagent/pkg/config"
)

const (
	store = "store"
)

const (
	statusOK = 0
)

//ErrMissingParam missing required parameter(s) error.
type ErrMissingParam []string

//Append add a missing parameter
func (e *ErrMissingParam) Append(param string) {
	*e = append(*e, param)
}

func (e *ErrMissingParam) String() string {
	return strings.Join(*e, ", ")
}

func (e *ErrMissingParam) Error() string {
	return fmt.Sprintf("missing parameter(s): ( %s )", e.String())
}

//HandlerFunc 消息处理handler
type HandlerFunc func(p *Store, w http.ResponseWriter, r *http.Request)

//UNUSED 临时注释编译错误
func UNUSED(...interface{}) {}

//g_handlerMap handler map
var g_handlerMap = map[string]HandlerFunc{
	"google": (*Store).googlePlay,
	"apple":  (*Store).appStore,
}

func requiredParams(req *http.Request, params ...string) (map[string]string, error) {
	m := make(map[string]string)

	errs := &ErrMissingParam{}

	for _, p := range params {
		m[p] = req.FormValue(p)
		if m[p] == "" {
			errs.Append(p)
		}
	}

	if len(*errs) != 0 {
		return nil, errs
	}

	return m, nil
}

//Response response to client.
type Response struct {
	Result bool   `json:"result"`
	Data   []byte `json:"data"`
}

func (r *Response) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}

	buf.WriteString(`{"result": `)
	if r.Result {
		buf.WriteString("true")
	} else {
		buf.WriteString("false")
	}
	buf.WriteString(",")
	buf.WriteString(`"data": `)
	buf.Write(r.Data)
	buf.WriteString("}")

	return buf.Bytes(), nil

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
	if typ == config.StoreAll && s.config.Proxy.AllOn {
		s.globalProxyOn = true
		os.Setenv("HTTP_PROXY", s.config.Proxy.Address)
		return
	}

	//分商店设置代理
	if (typ == config.StoreApple && s.config.Stores.AppStore.HTTPProxyOn) ||
		(typ == config.StoreGoogle && s.config.Stores.GooglePlay.HTTPProxyOn) {
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
func (s *Store) response(w http.ResponseWriter, req *http.Request, result bool, data []byte) {
	resp := Response{
		Result: result,
		Data:   data,
	}
	rspn, _ := json.Marshal(&resp)
	s.responseJSON(w, req, rspn)
}

//Response 向客户端的消息响应
func (s *Store) responseJSON(w http.ResponseWriter, req *http.Request, jsonText []byte) {
	w.Write(jsonText)
}

func (p *Store) dispatch(path string, w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	m, err := requiredParams(r, store)
	if err != nil || (m[store] != string(config.StoreApple) && m[store] != string(config.StoreGoogle)) {

		http.Error(w, err.Error(), http.StatusBadRequest)
		return

	} else {
		g_handlerMap[m[store]](p, w, r)
	}

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
	s.setProxyEnv(config.StoreAll)
	return s
}
