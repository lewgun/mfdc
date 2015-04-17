package db

import (
	"fmt"

	//just for import dirver
	_ "github.com/go-sql-driver/mysql"

	"github.com/go-xorm/xorm"

	"platserver/plat/ossWeb/models"

	"github.com/lewgun/mfdc/exporter/config"
)

//MySQL implements a wrapper for xorm
type MySQL struct {
	*xorm.Engine
}

//NewMySQL new a mysql connection.
func NewMySQL(c *config.MySQLConfig) (*MySQL, error) {
	if c == nil {
		return nil, fmt.Errorf("Illegal parameter.")
	}

	//"root:123new@tcp(125.64.93.75:3306)/oss?charset=utf8"
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/oss?charset=utf8",
		c.User,
		c.Password,
		c.IP,
		c.Port)

	eng, err := xorm.NewEngine("mysql", dsn)
	if err != nil {
		return nil, err
	}

	eng.ShowSQL = true
	return &MySQL{eng}, nil
}

//AllUsers get all user.
func (m *MySQL) AllUsers() ([]models.User, error) {

	users := make([]models.User, 0)
	err := m.Find(&users)
	return users, err

}

//CertificateOfUser get all certificate belong to specific user.
func (m *MySQL) CertificateOfUser(uid uint64) ([]models.SecurityFile, error) {

	certs := make([]models.SecurityFile, 0)
	err := m.Where("c_user_id = ?", uid).Find(&certs)
	return certs, err

}

//AppOfUser get all production belong to specific user.
func (m *MySQL) AppOfUser(uid uint64) ([]models.Product, error) {

	apps := make([]models.Product, 0)
	err := m.Where("c_user_id = ?", uid).Find(&apps)
	return apps, err

}

//VersionOfApp get all version info  specific to (user, production).
func (m *MySQL) VersionOfApp(uid, pid uint64) ([]models.ProductVersion, error) {

	versions := make([]models.ProductVersion, 0)
	err := m.Where("c_user_id = ? AND c_product_id = ?", uid, pid).Find(&versions)

	return versions, err

}

//CPOfApp get all cp  specific to (user, production).
func (m *MySQL) CPOfApp(uid, pid uint64) ([]models.ChannelUser, error) {

	//得到用户所有的渠道
	cus := make([]models.ChannelUser, 0)

	//select * from channel_user where c_user_id = ? AND c_product_id = ?
	err := m.Sql(`SELECT cu.c_id,cu.c_name,cu.c_user_id,cu.c_channel_id,cu.c_product_id,cu.c_backurl,cu.c_package_name
			FROM channel c, channel_user cu
			WHERE cu.c_channel_id = c.c_number  AND cu.c_user_id = ? AND cu.c_product_id = ?
			GROUP BY cu.c_id`, uid, pid).Find(&cus)

	return cus, err

}

//SignedBinariesOfApp get all cp  specific to (user, production).
func (m *MySQL) SignedBinariesOfApp(uid, pid, vid uint64) ([]*models.ChannelVersion, error) {

	cus, err := m.CPOfApp(uid, pid)
	if err != nil {
		return nil, err
	}

	cvs := make([]*models.ChannelVersion, 0)

	for _, cv := range cus {
		temp := &models.ChannelVersion{}

		_, err = m.Where("c_channel_user_id = ? AND c_product_version_id = ? AND c_packer_status = ?",
			cv.Id,
			vid,
			2).Get(temp)
		if err != nil || temp.Id == 0 {
			continue
		}

		cvs = append(cvs, temp)
	}

	return cvs, nil

}
