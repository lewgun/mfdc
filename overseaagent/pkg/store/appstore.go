package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"overseaagent/pkg/config"
)

func (s *Store) appStore(w http.ResponseWriter, req *http.Request) {

	s.setProxyEnv(config.APP_STORE)
	defer s.UnsetProxyEnv()

	req.ParseForm()
	err := s.auth(w, req)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	as := &s.config.Stores.AppStore

	url := as.ReleaseURL
	if as.Debug {
		url = as.DebugURL
	}

	m := map[string]string{
		"receipt-data": req.FormValue("receipt-data"),
	}

	data, err := json.Marshal(m)
	if err != nil {

	}

	rspn, err := http.Post(url,
		"application/json",
		bytes.NewReader(data))

	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	defer rspn.Body.Close()

	data, err = ioutil.ReadAll(rspn.Body)

	s.responseJSON(w, req, data)

}
