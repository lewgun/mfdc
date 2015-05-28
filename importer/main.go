package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/cheggaaa/pb"
	"github.com/codegangsta/cli"
	"github.com/lewgun/mfdc/aliyun_oss"
)

const (
	interval = 5
)

func upload(o *oss.AliYun, files []string) {

	count := len(files)

	cnt := count / interval

	if (count % interval) != 0 {
		cnt += 1
	}

	pb := pb.StartNew(cnt)
	pb.SetWidth(160)

	var wg sync.WaitGroup

	for i := 0; i < cnt; i++ {
		wg.Add(1)
		go func(start, end int, o *oss.AliYun) {
			defer wg.Done()

			if end > count {
				end = count
			}

			for j := start; j != end; j++ {
				path := files[j]
				metaURL, binURL, err := o.Import(path)
				if err != nil {
					fmt.Printf("\nUpload file: %s failed with error: %s\n", path, err)

				} else {
					fmt.Printf("\nUpload file: %s is  Successfully:\n\tMetaURL: %s\n\tBinURL: %s\n", path, metaURL, binURL)
				}

			}

			pb.Increment()
		}(i*interval, (i+1)*interval, o)
	}

	wg.Wait()
	pb.FinishPrint("The End!")
	fmt.Println("Export task is finished.")

	//o.ListBucket(oss.MFSDKMetaBucket)
	//o.ListBucket(oss.MFSDKBinariesBucket)

}

func listDir(root string) []string {

	files := make([]string, 0)

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return err
		}

		files = append(files, path)
		return err
	})

	return files
}

//makeBuckets make the binary & meta bucket
func makeBuckets(o *oss.AliYun) error {

	var err error

	if err = o.CreateBucket(oss.MFSDKBinariesBucket, oss.PublicRead); err != nil {
		return err
	}

	return o.CreateBucket(oss.MFSDKMetaBucket, oss.Private)

}

//emptyBuckets empty buckets
func emptyBuckets(o *oss.AliYun) error {

	o.DeleteBucket(oss.MFSDKBinariesBucket)
	o.DeleteBucket(oss.MFSDKMetaBucket)
	return makeBuckets(o)
}

func action() func(c *cli.Context) {
	return func(ctx *cli.Context) {

		o := oss.New(oss.EndpointBeiJing, ctx.String("keyId"), ctx.String("keySecret"))

		if ctx.String("empty") == "t" {
			err := makeBuckets(o)
			if err != nil {
				fmt.Printf("Make bucket failed with error: %v\n", err)
				return
			}

		}

		files := listDir(ctx.String("loc"))
		upload(o, files)

	}

}

func flags() []cli.Flag {
	f := []cli.Flag{
		cli.StringFlag{
			Name:  "loc",
			Value: "D:/testexport/",
			Usage: "The location of the imported data",
		},

		cli.StringFlag{
			Name:  "empty",
			Value: "f",
			Usage: "empty bucket(s) or not (f/t)",
		},

		cli.StringFlag{
			Name:  "keyId",
			Value: "",
			Usage: "the access key id of aliyun OSS",
		},
		cli.StringFlag{
			Name:  "keySecret",
			Value: "",
			Usage: "the access Key secret of aliyun OSS",
		},
	}

	return f
}

func main() {
	app := cli.NewApp()
	app.Name = "importer"
	app.Usage = "A tool for import data to AliYun OSS!"
	app.Version = "0.1"
	app.Author = "lewgun"

	app.Action = action()

	app.Flags = flags()

	app.Run(os.Args)
}
