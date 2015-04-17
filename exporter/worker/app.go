package worker

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"os"
	"path/filepath"
	"platserver/plat/ossWeb/models"
	//"github.com/golang/glog"
	"github.com/lewgun/mfdc/exporter/db"
	"sync"
)

type app struct {
	eng     *db.MySQL
	mgo     *mgo.Database
	product *models.Product
	path    string
}

func newApp(eng *db.MySQL, mgo *mgo.Database, p *models.Product, loc string) *app {
	a := &app{
		eng:     eng,
		mgo:     mgo,
		product: p,
	}

	if !a.init(loc) {
		return nil
	}

	return a
}

func (p *app) init(loc string) bool {

	p.path = filepath.Join(loc, fmt.Sprintf("%d", p.product.Id))

	//icon
	if err := os.MkdirAll(filepath.Join(p.path, iconDir), os.ModePerm); err != nil {
		return false
	}

	//versions
	if err := os.MkdirAll(filepath.Join(p.path, versionDir), os.ModePerm); err != nil {
		return false
	}

	return true

}

func (p *app) export() error {

	//icon
	var (
		err error
	)

	if err = p.icon(); err != nil {
		//glog.Info("Export Icon for app %d is failed.", p.product.Id)
		fmt.Printf("Export Icon for user: %d app %d is failed. with error: %v\n", p.product.UserId, p.product.Id, err)

	} else {
		fmt.Printf("Export Icon for user: %d app %d is successfuly.\n", p.product.UserId, p.product.Id)
	}

	//版本
	if err = p.versions(); err != nil {
		return err
	}

	return err

}

func (p *app) icon() error {
	return download(p.mgo, filepath.Join(p.path, iconDir), p.product.Icon)
}

//versions download the all versions for specific production.
func (p *app) versions() error {
	versions, err := p.eng.VersionOfApp(p.product.UserId, p.product.Id)
	if err != nil {
		return err
	}

	//download all versions
	var wg sync.WaitGroup
	for i := range versions {
		wg.Add(1)
		go func(ver *models.ProductVersion) {
			defer wg.Done()
			p := newVersion(p.eng, p.mgo, ver, filepath.Join(p.path, versionDir))
			if p == nil {
				return
			}
			p.export()
			p.done()
		}(&versions[i]) //!!!!!!!!!!!!!

	}

	wg.Wait()

	return err
}

func (p *app) done() error {
	return nil
}
