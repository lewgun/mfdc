package worker

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/lewgun/mfdc/exporter/db"
	"gopkg.in/mgo.v2"
	"os"
	"path/filepath"
	"strconv"
)

const (
	certDir    = "certs"
	appsDir    = "apps"
	iconDir    = "icon"
	originDir  = "origin"
	signedDir  = "signed"
	versionDir = "versions"
)

//Task task implements a export task.
type task struct {
	path string
	uid  uint64

	eng *db.MySQL
	mgo *mgo.Database
}

//newTask create a task.
func newTask(eng *db.MySQL, mgo *mgo.Database, uid uint64, loc string) *task {
	t := &task{
		eng: eng,
		mgo: mgo,
		uid: uid,
	}

	if !t.init(loc) {
		return nil
	}

	return t
}

func (t *task) init(loc string) bool {

	t.path = filepath.Join(loc, fmt.Sprintf("%d", t.uid)) //

	//证书
	if err := os.MkdirAll(filepath.Join(t.path, certDir), os.ModePerm); err != nil {
		return false
	}

	//apps
	if err := os.MkdirAll(filepath.Join(t.path, appsDir), os.ModePerm); err != nil {
		return false
	}

	return true
}

//export run a export task.
func (t *task) export() error {

	glog.Info("Export data for user: %d is starting.", t.uid)
	var (
		err error
	)

	//证书
	if err = t.certificates(); err != nil {
		return err
	}

	glog.Info("Export certificates for user %d is succssfully.", t.uid)

	//产品
	if err = t.apps(); err != nil {
		return err
	}

	return err
}

//done finish a export task.
func (t *task) done() error {
	return nil
}

//productions download the all apps for specific user.
func (t *task) apps() error {
	apps, err := t.eng.AppOfUser(t.uid)
	if err != nil {
		return err
	}

	//download all apps
	for _, a := range apps {
		p := newApp(t.eng, t.mgo, &a, filepath.Join(t.path, appsDir))
		if p == nil {
			fmt.Printf("New Export app: %s  task failed.\n", filepath.Join(t.path, appsDir, strconv.Itoa(int(a.Id))))
			continue
		}
		if err = p.export(); err != nil {
			fmt.Printf(" Export app: %s  task failed. with error: %v\n",
				filepath.Join(t.path, appsDir, strconv.Itoa(int(a.Id))),
				err)

		} else {
			fmt.Printf(" Export app: %s  task is successfully\n",
				filepath.Join(t.path, appsDir, strconv.Itoa(int(a.Id))),
				err)
		}
		p.done()
	}

	return err
}

//certificates download the all certificate for specific user.
func (t *task) certificates() error {

	certs, err := t.eng.CertificateOfUser(t.uid)
	if err != nil {
		return err
	}

	//download all certificates
	for _, cert := range certs {
		err = t.exportCertificate(cert.FileCode)
		if err == nil {
			glog.Info("Download certificate %s is sucessfully.\n", cert.FileCode)

		} else {
			//glog.Info("Download certificate %s is failed. with error: %v\n", cert.FileCode, err )
			fmt.Printf("Download certificate %s is failed.  %v\n", cert.FileCode, err)
		}
	}

	return nil

}

func (t *task) exportCertificate(uuid string) error {
	return download(t.mgo, filepath.Join(t.path, certDir), uuid)

}
