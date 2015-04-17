package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/golang/glog"

	"github.com/lewgun/mfdc/exporter/config"
	"github.com/lewgun/mfdc/exporter/db"
	"github.com/lewgun/mfdc/exporter/worker"
)

//mgoFlags mgo's command flags
func mgoFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "mip",
			Value: "",
			Usage: "MongoDB's ip",
			//EnvVar: "LEGACY_COMPAT_LANG,APP_LANG,LANG",
		},
		cli.StringFlag{
			Name:  "mport",
			Value: "",
			Usage: "MongoDB's port",
		},

		cli.StringFlag{
			Name:  "muser",
			Value: "",
			Usage: "MongoDB's user",
		},
		cli.StringFlag{
			Name:  "mpass",
			Value: "",
			Usage: "MongoDB's password",
		},

		cli.StringFlag{
			Name:  "mdb",
			Value: "fileServer",
			Usage: "MongoDB's current database",
		},
	}

}

//mysqlFlags mysql's command flags
func mysqlFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "ip",
			Value: "",
			Usage: "MySQL's ip",
			//EnvVar: "LEGACY_COMPAT_LANG,APP_LANG,LANG",
		},
		cli.StringFlag{
			Name:  "port",
			Value: "",
			Usage: "MySQL's port",
		},

		cli.StringFlag{
			Name:  "user",
			Value: "",
			Usage: "MySQL's user",
		},
		cli.StringFlag{
			Name:  "pass",
			Value: "",
			Usage: "MySQL's password",
		},

		cli.StringFlag{
			Name:  "db",
			Value: "oss",
			Usage: "MySQL's current database",
		},
	}

}

func action() func(c *cli.Context) {
	return func(c *cli.Context) {

		mysqlConfig := config.ParseMySQLConfig(c)
		if mysqlConfig == nil {
			glog.Fatal("Parse MySQL's config failed. Please recheck your config.")
		}

		glog.Infoln("Parse MySQL's config successfully.")

		mgoConfig := config.ParseMGOConfig(c)
		if mgoConfig == nil {
			glog.Fatal("Parse MongoDB's config failed. Please recheck your config.")
		}
		glog.Infoln("Parse MongoDB's config successfully.")

		mysql, err := db.NewMySQL(mysqlConfig)
		if err != nil {
			glog.Fatal(fmt.Sprintf("Init mysql connection failed with error: %v\n", err))
		}
		glog.Infoln("Init mysql connection successfully.")
		defer mysql.Close()

		mgo, err := db.NewMongoDB(mgoConfig)
		if err != nil {
			glog.Fatal(fmt.Sprintf("Init mysql connection failed with error: %v\n", err))
		}

		defer mgo.Close()

		glog.Infoln("Init MongoDB connection successfully.")
		w := worker.New(mysql, mgo.DB(c.String("mdb")), c.String("loc"))
		if w == nil {
			glog.Fatal("Create a worker failed.")
		}
		w.Export()

	}
}

func flags() []cli.Flag {
	f := []cli.Flag{
		cli.StringFlag{
			Name:  "loc",
			Value: "D:/testexport/",
			Usage: "The location for store exported data",
		},

		cli.BoolFlag{
			Name:  "alsologtostderr",
			Usage: "also log to stderr or not",
		},
		cli.StringFlag{
			Name:  "log_dir",
			Value: "D:/testexport/",
			Usage: "log file directory",
		},
	}

	f = append(f, mgoFlags()...)
	f = append(f, mysqlFlags()...)
	return f
}

func main() {
	app := cli.NewApp()
	app.Name = "exporter"
	app.Usage = "A tool for export data from MongoDB!"
	app.Version = "0.1"
	app.Author = "lewgun"

	app.Action = action()

	app.Flags = flags()

	app.Run(os.Args)
}
