package store

import (
	"os"
	"strings"
)

const (
	rootAPPStore   = ".apple.com"
	rootGooglePlay = ".googleapis.com"
)

//noProxys 设置no proxy的辅助类型
type noProxys []string

func (np *noProxys) append(h string) {
	*np = append(*np, h)
}

func (np *noProxys) set() {
	os.Setenv("NO_PROXY", strings.Join(*np, ","))
}
func (np *noProxys) unset() {
	os.Unsetenv("NO_PROXY")
}
