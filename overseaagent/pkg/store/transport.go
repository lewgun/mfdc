package store

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"overseaagent/pkg/config"
)

type ProxyHandler func(*http.Request) (*url.URL, error)

func useProxy(addr string, conf *config.Config) bool {

	if conf.Proxy.AllOn {
		return true
	}

	//google play
	if strings.Contains(addr, "googleapis") && conf.Stores.GooglePlay.HTTPProxyOn {
		return true
	}

	if strings.Contains(addr, "apple") && conf.Stores.AppStore.HTTPProxyOn {
		return true
	}
	return false

}
func newProxyHandler(conf *config.Config) ProxyHandler {

	return func(req *http.Request) (*url.URL, error) {
		proxy := conf.Proxy.Address
		if proxy == "" {
			return nil, nil
		}

		if !useProxy(req.URL.Host, conf) {
			return nil, nil
		}
		proxyURL, err := url.Parse(proxy)
		if err != nil || !strings.HasPrefix(proxyURL.Scheme, "http") {
			// proxy was bogus. Try prepending "http://" to it and
			// see if that parses correctly. If not, we fall
			// through and complain about the original one.
			if proxyURL, err := url.Parse("http://" + proxy); err == nil {
				return proxyURL, nil
			}
		}
		if err != nil {
			return nil, fmt.Errorf("invalid proxy address %q: %v", proxy, err)
		}
		return proxyURL, nil

	}

}

func newTransport(conf *config.Config) *http.Transport {

	return &http.Transport{
		Proxy: newProxyHandler(conf),
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}

}
