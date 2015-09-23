package store

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func oauthHTTPClient(clientID, clientSecret, refreshToken string) *http.Client {

	oAuthClient := oauth2.NewClient(oauth2.NoContext, oauthutil.NewRefreshTokenSource(&oauth2.Config{
		Scopes:       []string{AndroidpublisherScope},
		Endpoint:     google.Endpoint,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  oauthutil.TitleBarRedirectURL,
	}, refreshToken))

	return oAuthClient
}

var g_Once sync.Once

func (s *Store) googlePlay(w http.ResponseWriter, req *http.Request) {

	m, err := requiredParams(req, packageName, productID, token, developerPayload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var oauthClient *http.Client

	g_Once.Do(func() {
		oauthClient = oauthHTTPClient(
			s.config.Stores.GooglePlay.ClientID,
			s.config.Stores.GooglePlay.ClientSecret,
			s.config.Stores.GooglePlay.RefreshToken,
		)
	})

	err = s.auth(w, req)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	s.setProxyEnv(config.StoreGoogle)
	defer s.UnsetProxyEnv()

	url := playIAPUrl(s.config.Stores.GooglePlay.URL, m)

	r, err := http.NewRequest("GET", url, nil)

	rspn, err := oauthClient.Do(r)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	defer rspn.Body.Close()

	data, err := ioutil.ReadAll(rspn.Body)
	fmt.Println(string(data))

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
