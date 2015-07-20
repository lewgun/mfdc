//Package oss implements operate aliyun oss.
package oss

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"bytes"
	alioss "github.com/PinIdea/oss-aliyun-go"
	"github.com/lewgun/mfdc/uuid"
)

const (

	//MFSDKMetaBucket is the mfsdk meta bucket
	MFSDKMetaBucket = "mfsdk-meta-bucket"

	//MFSDKBinariesBucket is the mfsdk binary bucket
	MFSDKBinariesBucket = "mfsdk-binaries-bucket"
)

const (
	//PublicRead anybody can read it.
	PublicRead = alioss.PublicRead

	//Private nobody can r&w it except me.
	Private = alioss.Private
)

const (

	//EndpointBeiJing setup your oss at BeiJing.
	EndpointBeiJing = "oss-cn-beijing"
)

func unused(u ...interface{}) {

}

//OSS is the data struct for operate aliyun oss
type AliYun struct {
	ali *alioss.OSS
}

//New new a OSS instance.
func New(endport, key, secret string) *AliYun {
	return &AliYun{
		ali: alioss.New(endport, key, secret),
	}
}

//CreateBucket create a bucket with name and perm.
func (oss *AliYun) CreateBucket(name string, perm alioss.ACL) error {
	b := alioss.Bucket{
		OSS:  oss.ali,
		Name: name,
	}
	return b.PutBucket(perm)
}

// Bucket returns a Bucket with the given name.
func (oss *AliYun) Bucket(name string) *alioss.Bucket {
	return oss.ali.Bucket(name)
}

// updateMeta 上传meta信息
func (oss *AliYun) uploadMeta(uuid, name string) (string, error) {

	bucket := oss.ali.Bucket(MFSDKMetaBucket)

	err := bucket.Put(uuid, []byte(name), "text/plain", alioss.Private)
	if err != nil {
		return "", err
	}

	return bucket.URL(uuid), nil

}

//updateBinary 上传2进制文件
func (oss *AliYun) UploadBinary(path string) (string, error) {

	bucket := oss.ali.Bucket(MFSDKBinariesBucket)

	base := filepath.Base(path)

	var (
		uuid string
	)
	idx := strings.Index(base, "_")
	if idx > -1 {
		uuid = base[:idx]

	} else {
		uuid = base
	}

	multi, err := bucket.InitMulti(uuid, "application/octet-stream", alioss.PublicRead)

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer f.Close()

	stats, err := f.Stat()

	parts, err := multi.PutAll(f, stats.Size())

	multi.Complete(parts)

	return bucket.URL(uuid), nil

}

//ListBucket list all file in the bucket.
func (oss *AliYun) ListBucket(name string) {

	b := oss.Bucket(name)
	results, err := b.List("", "", "", 200)
	if err != nil {
		return
	}

	fmt.Println("Bucket: ", results.Name)

	for _, res := range results.Contents {
		fmt.Println("\tObject: ", res.Key)
	}

}

//Import import a file.
// just for importer.
func (oss *AliYun) Import(path string) (metaURL, binURL string, err error) {

	base := filepath.Base(path)
	fields := strings.SplitN(base, "_", 2)

	metaURL, err = oss.uploadMeta(fields[0], fields[1])
	if err != nil {
		return
	}

	binURL, err = oss.UploadBinary(path)
	return

}

//DeleteBucket delete the bucket with name.
func (oss *AliYun) DeleteBucket(name string) error {

	bucket := oss.ali.Bucket(name)
	return bucket.DelBucket()
}

// Delete delete a signed apk/ipa.
func (oss *AliYun) DeleteFile(uuid string) error {

	//delete meta
	bucket := oss.ali.Bucket(MFSDKMetaBucket)
	bucket.Del(uuid)

	//delete binary
	bucket = oss.ali.Bucket(MFSDKBinariesBucket)
	return bucket.Del(uuid)
}

//BinaryName return the binary name.
func (oss *AliYun) FileName(uuid string) string {

	bucket := oss.ali.Bucket(MFSDKMetaBucket)

	// get object
	content, err := bucket.Get(uuid)
	if err != nil {
		return ""
	}

	return string(content)

}

//OpenFile open file for read. please close it by yourself.
func (oss *AliYun) OpenFile(uuid string) (io.ReadCloser, error) {
	bucket := oss.ali.Bucket(MFSDKBinariesBucket)
	return bucket.GetReader(uuid)

}

//WriteFile save apk&ipa to oss.
func (oss *AliYun) WriteFile(name string, src multipart.File, size int) (string, error) {

	uuid := uuid.New()

	bucket := oss.Bucket(MFSDKBinariesBucket)
	multi, err := bucket.InitMulti(uuid, "application/octet-stream", alioss.PublicRead)
	if err != nil {
		return "", err
	}

	parts, err := multi.PutAll(src, int64(size))
	if err != nil {
		return "", err
	}

	if err = multi.Complete(parts); err != nil {
		return "", err
	}

	if _, err = oss.uploadMeta(uuid, name); err != nil {
		oss.DeleteFile(uuid)
		return "", err
	}

	return uuid, nil

}

//UploadFile copy opend file to oss.
func (oss *AliYun) UploadFile(name string, src *os.File) (string, error) {
	if name == "" || src == nil {
		return "", fmt.Errorf("Illegal parameter.")
	}

	uuid := uuid.New()

	bucket := oss.Bucket(MFSDKBinariesBucket)

	multi, err := bucket.InitMulti(uuid, "application/octet-stream", alioss.PublicRead)

	stats, err := src.Stat()

	parts, err := multi.PutAll(src, stats.Size())
	if err != nil {
		return "", err
	}

	if err = multi.Complete(parts); err != nil {
		return "", err
	}

	_, err = oss.uploadMeta(uuid, name)
	if err != nil {
		oss.DeleteFile(uuid)
		return "", err
	}

	return uuid, nil
}

//WriteBytes write bytes to oss.
func (oss *AliYun) WriteBytes(name string, src *bytes.Reader) (string, error) {
	if name == "" {
		return "", fmt.Errorf("Illegal parameter.")
	}

	uuid := uuid.New()

	bucket := oss.Bucket(MFSDKBinariesBucket)

	multi, err := bucket.InitMulti(uuid, "application/octet-stream", alioss.PublicRead)

	parts, err := multi.PutAll(src, int64(src.Len()))
	if err != nil {
		return "", err
	}

	if err = multi.Complete(parts); err != nil {
		return "", err
	}

	_, err = oss.uploadMeta(uuid, name)
	if err != nil {
		oss.DeleteFile(uuid)
		return "", err
	}

	return uuid, nil
}
