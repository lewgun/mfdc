//Read the agent's settings
package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

type StoreType string

func (s StoreType) String() string {
	return string(s)
}

const (
	GOOGLE_PLAY StoreType = "googleplay"
	APP_STORE   StoreType = "appstore"
	ALL_STORE   StoreType = "allstore"
)

var (
	ErrIllegalParam = errors.New("illegal parameter(s).")
)

type AppStore struct {
	Debug       bool `json:"debug"`
	HTTPProxyOn bool `json:"http_proxy"`
	PowerOn     bool `json:"power_on"`

	DebugURL   string `json:"debug_url"`
	ReleaseURL string `json:"release_url"`
}

func (as *AppStore) String() string {
	buf := &bytes.Buffer{}
	buf.WriteString("IsDebug: ")

	if as.Debug {
		buf.WriteString("true")

	} else {
		buf.WriteString("false")
	}

	buf.WriteString("\t")
	buf.WriteString("DebugURL: " + as.DebugURL + "\t")
	buf.WriteString("ReleaseURL: " + as.ReleaseURL)

	buf.WriteString("\t")

	buf.WriteString("http_proxy: ")

	if as.HTTPProxyOn {
		buf.WriteString("true")

	} else {
		buf.WriteString("false")
	}

	buf.WriteString("\t")

	buf.WriteString("power_on: ")

	if as.PowerOn {
		buf.WriteString("true")

	} else {
		buf.WriteString("false")
	}

	return buf.String()
}

type GooglePlay struct {
	URL          string `json:"url"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
	HTTPProxyOn  bool   `json:"http_proxy"`
	PowerOn      bool   `json:"power_on"`
}

func (gp *GooglePlay) String() string {
	buf := &bytes.Buffer{}
	buf.WriteString("client_id: " + gp.ClientID)
	buf.WriteString("\t")

	buf.WriteString("client_secret: " + gp.ClientSecret)
	buf.WriteString("\t")

	buf.WriteString("refresh_token: " + gp.RefreshToken)
	buf.WriteString("\t")

	buf.WriteString("URL: " + gp.URL)
	buf.WriteString("\t")

	buf.WriteString("http_proxy: ")

	if gp.HTTPProxyOn {
		buf.WriteString("true")

	} else {
		buf.WriteString("false")
	}

	buf.WriteString("\t")

	buf.WriteString("power_on: ")

	if gp.PowerOn {
		buf.WriteString("true")

	} else {
		buf.WriteString("false")
	}

	return buf.String()
}

//Storer
type Stores struct {
	AppStore   `json:"appstore"`
	GooglePlay `json:"googleplay"`
}

type Host struct {
	Port string `json:"port`
}

func (a *Host) String() string {
	return "Port: " + a.Port
}

type Proxy struct {
	Address string `json:"address`
	AllOn   bool   `json:"all_on"`
}

func (p *Proxy) String() string {

	buf := &bytes.Buffer{}
	buf.WriteString("address: " + p.Address)
	buf.WriteString("\t")

	buf.WriteString("all_on: ")

	if p.AllOn {
		buf.WriteString("true")

	} else {
		buf.WriteString("false")
	}

	return buf.String()

}

type Config struct {
	Stores Stores `json:"stores"`
	Host   Host   `json:"host"`
	Proxy  Proxy  `json:"proxy"`
}

func (c *Config) String() string {
	buf := &bytes.Buffer{}

	buf.WriteString("Stores:")

	gp := &c.Stores.GooglePlay
	buf.WriteString("\n\t")
	buf.WriteString(GOOGLE_PLAY.String())
	buf.WriteString(":\t")
	buf.WriteString(gp.String())

	as := &c.Stores.AppStore
	buf.WriteString("\n\t")
	buf.WriteString(APP_STORE.String())
	buf.WriteString(":\t")
	buf.WriteString(as.String())

	buf.WriteString("\nHost:")
	buf.WriteString("\n\t")
	buf.WriteString(c.Host.String())

	buf.WriteString("\nProxy:")
	buf.WriteString("\n\t")
	buf.WriteString(c.Proxy.String())
	return buf.String()
}

func (c *Config) parse(path string) error {
	if path == "" {
		return ErrIllegalParam
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, c)

	fmt.Println(c)

	return err

}

func (c *Config) adjust() error {

	if !strings.HasPrefix(c.Host.Port, ":") {
		c.Host.Port = ":" + c.Host.Port
	}

	//如果开启了全局代理 则分商店代理设置无效,优化操作
	if c.Proxy.AllOn {
		c.Stores.AppStore.HTTPProxyOn = false
		c.Stores.GooglePlay.HTTPProxyOn = false
	}

	return nil
}

//check检测配置参数是否完备
func (c *Config) check() error {

	//Host
	if c.Host.Port == "" {
		return fmt.Errorf("Host's settings are not finished.")
	}

	//Google Play
	gp := &c.Stores.GooglePlay
	if gp.PowerOn {
		if gp.ClientID == "" || gp.ClientSecret == "" || gp.RefreshToken == "" || gp.URL == "" {
			return fmt.Errorf("Google Play's settings are not finished.")
		}
	}

	//App Store
	as := &c.Stores.AppStore
	if as.PowerOn {
		if as.Debug {
			if as.DebugURL == "" {
				return fmt.Errorf("APP Store's settings are not finished.")
			}
		} else {
			if as.ReleaseURL == "" {
				return fmt.Errorf("APP Store's settings are not finished.")
			}
		}

	}

	//Proxy
	if (c.Proxy.Address == "") && (c.Proxy.AllOn || c.Stores.AppStore.HTTPProxyOn || c.Stores.GooglePlay.HTTPProxyOn) {
		return fmt.Errorf("Proxy's settings are not finished.")
	}

	return nil

}

func (c *Config) init(path string) error {

	var err error
	if err = c.parse(path); err != nil {
		return fmt.Errorf("Can't load config from: %s with error: %v.", path, err)
	}

	if err = c.adjust(); err != nil {
		return fmt.Errorf("Adjust config failed.")
	}

	return c.check()

}

//New 创建一个配置
func New(path string) *Config {

	c := &Config{}
	err := c.init(path)
	if err != nil {
		panic(err)
	}

	return c
}
