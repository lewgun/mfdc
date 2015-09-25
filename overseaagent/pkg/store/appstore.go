package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	receiptData   = "receipt-data"
	transactionID = "transactionId"
)

const (
	statusSandbox = 21007 //sandbox 模式
)

func (s *Store) appStore(w http.ResponseWriter, req *http.Request) {

	m, err := requiredParams(req, receiptData /*, TRANSACTION_ID*/) //暂未使用
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.auth(w, req)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	data, err := json.Marshal(m)
	r := bytes.NewReader(data)

	//返回状态检测,只读取部分应答字段
	var rspnStub struct {
		Status int `json:"status`
	}

	as := &s.config.Stores.AppStore
	url := as.ReleaseURL

	var result bool
	for {
		r.Seek(0, os.SEEK_SET)
		data, err = s.appStorePOST(url, r) //首选release模式,如果出现21007错误,则再次尝试sandbox模式
		if err != nil {
			fmt.Println(err)
			http.Error(w, "internal error ", http.StatusInternalServerError)
			return
		}

		err = json.Unmarshal(data, &rspnStub)
		if err != nil {
			http.Error(w, "internal error ", http.StatusInternalServerError)
			return
		}

		if rspnStub.Status == statusSandbox {
			url = as.DebugURL
			continue //retry with sandbox url.
		}

		if rspnStub.Status == statusOK {
			result = true
		}

		break
	}

	s.response(w, req, result, data)

}

//appStorePOST Http POST to appstore for verify.
func (s *Store) appStorePOST(url string, r io.Reader) ([]byte, error) {

	rspn, err := http.Post(url,
		"application/json",
		r)

	if err != nil {
		return nil, err
	}

	defer rspn.Body.Close()

	return ioutil.ReadAll(rspn.Body)

}
