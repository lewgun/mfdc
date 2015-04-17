//Package worker do the actual export work.
package worker

import (
	"fmt"
	"github.com/cheggaaa/pb"
	"github.com/golang/glog"
	"github.com/lewgun/mfdc/exporter/config"
	"github.com/lewgun/mfdc/exporter/db"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io"
	"os"
	"path/filepath"
	"platserver/plat/ossWeb/models"
	"sync"
)

func download(mgo *mgo.Database, dir, uuid string) error {

	var (
		err error
	)
	meta := &config.FileName{}
	if err = mgo.C(db.MetaCollection).Find(bson.M{"fileid": uuid}).One(meta); err != nil {
		//glog.Warning("Get metainfo for %s is failed with error: %v", uuid, err )
		fmt.Printf("Get metainfo for %s is failed with error: %v\n", uuid, err)
	}

	src, err := mgo.GridFS(db.BinaryStore).Open(uuid)
	if err != nil {
		return fmt.Errorf("Open binary %s failed with error: %v\n", uuid, err)
	}
	defer src.Close()

	name := meta.FileId + "_" + meta.FileName

	dst, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)

	return err

}

//Worker implements the actual export work.
type Worker struct {
	loc string

	eng *db.MySQL
	mgo *mgo.Database

	users []models.User
	pb    *pb.ProgressBar
}

//New new a worker.
func New(eng *db.MySQL, mgo *mgo.Database, loc string) *Worker {
	if eng == nil || mgo == nil {
		return nil
	}

	w := &Worker{}

	if !w.init(eng, mgo, loc) {
		return nil
	}

	return w
}

func (w *Worker) init(eng *db.MySQL, mgo *mgo.Database, loc string) bool {

	w.eng = eng
	w.mgo = mgo
	w.loc = loc

	return true
}

func (w *Worker) prepare() error {

	var (
		err error
	)

	if w.users, err = w.eng.AllUsers(); err != nil {
		return err
	}

	count := len(w.users)

	w.pb = pb.StartNew(count)
	w.pb.SetWidth(160)
	return nil
}

//Export Do do the export work.
func (w *Worker) Export() error {

	glog.Info("Export work is starting\n")

	var (
		err error
	)

	w.prepare()

	var wg sync.WaitGroup
	for _, u := range w.users {
		wg.Add(1)
		go func(uid uint64) {
			defer wg.Done()
			t := newTask(w.eng, w.mgo, uid, w.loc)
			if t != nil {
				glog.Info("Start export task for user: %d is succssfully.\n", uid)

			} else {
				glog.Info("Start export task for user: %d is failed.\n", uid)
			}

			if err = t.export(); err != nil {
				glog.Info("Run export task for user: %d is failed with error %v.\n", uid, err)

			} else {
				glog.Info("Run export task for user: %d is successfully.\n", uid, err)
			}

			t.done()
			w.pb.Increment()
		}(u.Id)

	}

	wg.Wait()
	w.pb.FinishPrint("The End!")
	fmt.Println("Export task is finished.")
	return nil
}
