//Package db implements the comunicate with MySQL and MongoDB.
package db

import (
	"fmt"
	mgo "gopkg.in/mgo.v2"

	"github.com/lewgun/mfdc/exporter/config"
)

const (
	//BinaryStore 存放二进制文件集合
	BinaryStore = "fs"

	//MetaCollection 存放二进制文件名和真实文件名对应关系的集合
	MetaCollection = "name"
)

//MongoDB implements a mgo wrapper.
type MongoDB struct {
	*mgo.Session
}

//NewMongoDB new a session with MongoDB.
func NewMongoDB(c *config.MGOConfig) (*MongoDB, error) {
	if c == nil {
		return nil, fmt.Errorf("Illegal parameter.")
	}

	session, err := mgo.Dial(c.IP)
	if err != nil {
		return nil, err
	}

	session.SetMode(mgo.Monotonic, true)

	adminDB := session.DB("admin")
	err = adminDB.Login(c.User, c.Password)
	if err != nil {
		session.Close()
		return nil, err
	}

	return &MongoDB{
		session,
	}, nil

}
