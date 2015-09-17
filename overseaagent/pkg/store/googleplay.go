package store

import (
	"bytes"
	"errors"
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
	PACKAGE_NAME = "packageName"
	PRODUCT_ID   = "productId"
	TOKEN        = "token"
)

const (
	AndroidpublisherScope = "https://www.googleapis.com/auth/androidpublisher"
)

var (
	ErrMissingParam = errors.New("missing parameter(s).")
)

func requiredParams(req *http.Request, params ...string) (map[string]string, error) {

	m := make(map[string]string)

	for _, p := range params {
		m[p] = req.FormValue(p)
		if m[p] == "" {
			return nil, ErrMissingParam
		}
	}
	return m, nil
}

func playIAPUrl(base string, params map[string]string) string {
	urlBuf := bytes.NewBufferString(base)
	if !strings.HasSuffix(base, "/") {
		urlBuf.WriteString("/")
	}
	urlBuf.WriteString(params[PACKAGE_NAME] + "/")
	urlBuf.WriteString("purchases/products/")
	urlBuf.WriteString(params[PRODUCT_ID] + "/")
	urlBuf.WriteString("tokens/")
	urlBuf.WriteString(params[TOKEN])

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

	s.setProxyEnv(config.GOOGLE_PLAY)
	defer s.UnsetProxyEnv()

	var oauthClient *http.Client

	g_Once.Do(func() {
		oauthClient = oauthHTTPClient(
			s.config.Stores.GooglePlay.ClientID,
			s.config.Stores.GooglePlay.ClientSecret,
			s.config.Stores.GooglePlay.RefreshToken,
		)
	})

	req.ParseForm()
	err := s.auth(w, req)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	m, err := requiredParams(req, PACKAGE_NAME, PRODUCT_ID, TOKEN)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	url := playIAPUrl(s.config.Stores.GooglePlay.URL, m)

	r, err := http.NewRequest("GET", url, nil)

	rspn, err := oauthClient.Do(r)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	defer rspn.Body.Close()

	data, err := ioutil.ReadAll(rspn.Body)

	s.responseJSON(w, req, data)

}
