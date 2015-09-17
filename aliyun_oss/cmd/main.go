package main

import (
	"fmt"
	"github.com/lewgun/mfdc/aliyun_oss"
	"io"
	"log"
	"net/http"
	"os"
)

var ox *oss.AliYun

func init() {
	ox = oss.New(oss.EndpointBeiJing, "z0cBpzR3zjH50Xj7", "7rewFS8T5KUSsELeB7h52QKZRcFvv8")
}

// 获取文件大小的接口
type Size interface {
	Size() int64
}

// 获取文件信息的接口
type Stat interface {
	Stat() (os.FileInfo, error)
}

// hello world, the web server
func HelloServer(w http.ResponseWriter, r *http.Request) {
	if "POST" == r.Method {
		file, _, err := r.FormFile("userfile")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var (
			size int64
		)

		if statInterface, ok := file.(Stat); ok {
			fileInfo, _ := statInterface.Stat()
			size = fileInfo.Size()
			fmt.Fprintf(w, "上传文件的大小为: %d", fileInfo.Size())
		}
		if sizeInterface, ok := file.(Size); ok {
			size = sizeInterface.Size()
			fmt.Fprintf(w, "上传文件的大小为: %d", sizeInterface.Size())
		}

		uuid, err := ox.WriteFile("MyMFSDKText.apk", file, int(size))
		fmt.Printf("UUID: %s\n", uuid)

		ox.OpenFile()

		ox.DeleteFile()
		http.ResponseWriter{}

		return
	}

	// 上传页面
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(200)
	html := `
<form enctype="multipart/form-data" action="/hello" method="POST">
    Send this file: <input name="userfile" type="file" />
    <input type="submit" value="Send File" />
</form>
`
	io.WriteString(w, html)
}

func main() {
	http.HandleFunc("/hello", HelloServer)
	err := http.ListenAndServe(":12315", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
