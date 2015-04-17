package worker

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/mgo.v2"
	"platserver/plat/ossWeb/models"
	//"github.com/golang/glog"
	"github.com/lewgun/mfdc/exporter/db"
	"strconv"
)

type version struct {
	eng  *db.MySQL
	mgo  *mgo.Database
	ver  *models.ProductVersion
	path string
}

func newVersion(eng *db.MySQL, mgo *mgo.Database, ver *models.ProductVersion, loc string) *version {
	a := &version{
		eng: eng,
		mgo: mgo,
		ver: ver,
	}

	if !a.init(loc) {
		return nil
	}

	return a
}

func (p *version) init(loc string) bool {

	p.path = filepath.Join(loc, fmt.Sprintf("%d", p.ver.Id))

	//origin apk dir
	if err := os.MkdirAll(filepath.Join(p.path, originDir), os.ModePerm); err != nil {
		return false
	}

	return true

}

func (p *version) export() error {

	//icon
	var (
		err error
	)

	//origin apk
	if err = p.originBinary(); err != nil {
		//glog.Info("Export Icon for app %d is failed.", p.product.Id)
		fmt.Printf("Export origin apk for user: %d production: %d  version: %d is failed.\n", p.ver.UserId, p.ver.ProductId, p.ver.Id)
	} else {
		fmt.Printf("Export origin apk for user: %d production: %d  version: %d is suessfully.\n", p.ver.UserId, p.ver.ProductId, p.ver.Id)
	}

	//signed apk
	if err = p.signedBinaries(); err != nil {
		return err
	}

	return err

}

//originBinary get the origin binary (aka no signed apk )
func (p *version) originBinary() error {
	return download(p.mgo, filepath.Join(p.path, originDir), p.ver.CodeDB)
}

//signedBinaries get the signed binaries by cp.
func (p *version) signedBinaries() error {

	cvs, err := p.eng.SignedBinariesOfApp(p.ver.UserId, p.ver.ProductId, p.ver.Id)
	if err != nil {
		return err
	}

	for _, cv := range cvs {

		//origin apk dir
		if err := os.MkdirAll(filepath.Join(p.path, signedDir, strconv.Itoa(int(cv.Id))), os.ModePerm); err != nil {
			continue
		}

		err = download(p.mgo, filepath.Join(p.path, signedDir, strconv.Itoa(int(cv.Id))), cv.CodeDB)
		if err != nil {
			fmt.Printf("Download signed binary for: User(%d) Production(%d)  Version(%d) CVID(%d) UUID(%s) failed with error: %v\n",
				p.ver.UserId,
				p.ver.ProductId,
				p.ver.Id,
				cv.Id,
				cv.CodeDB,
				err)
		} else {
			fmt.Printf("Download signed binary for: User(%d) Production(%d)  Version(%d) CVID(%d) UUID(%s) SUCESSFULLY\n",
				p.ver.UserId,
				p.ver.ProductId,
				p.ver.Id,
				cv.Id,
				cv.CodeDB)
		}
	}

	return nil
}

func (p *version) done() error {
	return nil
}
