//Package config implements parse the config for communicate with MySQL and MongoDB.
package config

import (
	"fmt"
	"reflect"

	"github.com/codegangsta/cli"
)

//FileName 文件真实名称与数据库中的uuid的映射
type FileName struct {
	FileId   string //文件UUID
	FileName string //文件真实名称
}

//MySQLConfig is the config for communicate with mysql.
type MySQLConfig struct {
	IP   string "ip"
	Port string "port"

	User     string "user"
	Password string "pass"

	DB string "db"
}

//MGOConfig is the config for communicate with MongoDB.
type MGOConfig struct {
	IP   string "mip"
	Port string "mport"

	User     string "muser"
	Password string "mpass"

	DB string "mdb"
}

func parseConfigHelper(c *cli.Context, conf interface{}) error {
	if c == nil || conf == nil {
		return fmt.Errorf("Illegal parameter(s).")
	}

	var (
		mysqlConf *MySQLConfig
		mgoConf   *MGOConfig
	)

	mysqlConf, ok := conf.(*MySQLConfig)
	if !ok {
		mgoConf, ok = conf.(*MGOConfig)
	}

	//都不能被转化
	if (mgoConf == nil) && (mysqlConf == nil) {
		return fmt.Errorf("Unknown data structure.")
	}

	if (mgoConf != nil) && (mysqlConf != nil) {
		return fmt.Errorf("Cast failed.")
	}

	t := reflect.TypeOf(conf).Elem()
	v := reflect.ValueOf(conf).Elem()
	for i := 0; i < t.NumField(); i++ {
		val := c.String(string(t.Field(i).Tag))
		v.Field(i).SetString(val)
	}

	return nil

}

//ParseMySQLConfig parse MySQL config from context.
func ParseMySQLConfig(c *cli.Context) *MySQLConfig {
	if c == nil {
		return nil
	}

	conf := &MySQLConfig{}
	if parseConfigHelper(c, conf) != nil {
		return nil
	}
	fmt.Println(conf)
	return conf

}

//ParseMGOConfig parse MongoDB config from context.
func ParseMGOConfig(c *cli.Context) *MGOConfig {
	if c == nil {
		return nil
	}

	conf := &MGOConfig{}
	if parseConfigHelper(c, conf) != nil {
		return nil
	}
	return conf

}
