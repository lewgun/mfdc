package store

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"overseaagent/pkg/config"
	"overseaagent/pkg/oauthutil"
)

const (
	packageName      = "packageName"
	productID        = "productId"
	token            = "token"
	developerPayload = "developerPayload"
)

const (
	AndroidpublisherScope = "https://www.googleapis.com/auth/androidpublisher"
)

func playIAPUrl(base string, params map[string]string) string {
	urlBuf := bytes.NewBufferString(base)
	if !strings.HasSuffix(base, "/") {
		urlBuf.WriteString("/")
	}
	urlBuf.WriteString(params[packageName] + "/")
	urlBuf.WriteString("purchases/products/")
	urlBuf.WriteString(params[productID] + "/")
	urlBuf.WriteString("tokens/")
	urlBuf.WriteString(params[token])

	return urlBuf.String()
}

func oauthHTTPClient(conf *config.Config) *http.Client {

	c := oauth2.NewClient(oauth2.NoContext, oauthutil.NewRefreshTokenSource(&oauth2.Config{
		Scopes:       []string{AndroidpublisherScope},
		Endpoint:     google.Endpoint,
		ClientID:     conf.Stores.GooglePlay.ClientID,
		ClientSecret: conf.Stores.GooglePlay.ClientSecret,
		RedirectURL:  oauthutil.TitleBarRedirectURL,
	}, conf.Stores.GooglePlay.RefreshToken))

	return c
}

var oauthClientOnce sync.Once
var oauthClient *http.Client

func (s *Store) googlePlay(w http.ResponseWriter, req *http.Request) {

	m, err := requiredParams(req, packageName, productID, token, developerPayload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.auth(w, req)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	oauthClientOnce.Do(func() {
		oauthClient = oauthHTTPClient(s.config)
	})

	url := playIAPUrl(s.config.Stores.GooglePlay.URL, m)

	r, err := http.NewRequest("GET", url, nil)

	rspn, err := oauthClient.Do(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer rspn.Body.Close()

	data, err := ioutil.ReadAll(rspn.Body)

	//返回状态检测,只读取部分应答字段
	var rspnStub struct {
		PurchaseState    int    `json:"purchaseState`
		DeveloperPayload string `jsong:"developerPayload"`
	}

	err = json.Unmarshal(data, &rspnStub)
	if err != nil {
		http.Error(w, "internal error ", http.StatusInternalServerError)
		return
	}

	var result bool
	if rspnStub.DeveloperPayload == m[developerPayload] && rspnStub.PurchaseState == statusOK {
		result = true
	}

	s.response(w, req, result, data)

}
