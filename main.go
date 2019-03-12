package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/araddon/gou"

	"github.com/lytics/cloudstorage"
	"github.com/lytics/cloudstorage/awss3"
)

// config cloudstorage config
var config *cloudstorage.Config

// init
func init() {
	config = &cloudstorage.Config{
		Type:       awss3.StoreType,
		AuthMethod: awss3.AuthAccessKey,
		Settings:   gou.JsonHelper{},
		TmpDir:     "/tmp/localcache",
		BaseUrl:    os.Getenv("AWS_BASE_URL"),
		Bucket:     os.Getenv("AWS_BUCKET"),
		Region:     os.Getenv("AWS_REGION"),
	}

	config.Settings[awss3.ConfKeyAccessKey] = os.Getenv("AWS_ACCESS_KEY")
	config.Settings[awss3.ConfKeyAccessSecret] = os.Getenv("AWS_ACCESS_SECRET")
}

// handler cloud zip handler
func handler(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Content-Disposition", "attachment; filename="+config.Bucket+".zip")
	w.Header().Add("Content-Type", "application/zip")

	files, ok := r.URL.Query()["files"]
	if !ok || len(files) < 1 {
		http.Error(w, "Specify at least one file with ?files='file1,file2,fileN'", 500)
		return
	}

	store, err := cloudstorage.NewStore(config)
	if err != nil {
		panic(err)
	}
	wrt, err := store.NewWriter(time.Now().String(), map[string]string{})
	if err != nil {
		panic(err)
	}
	pr, pw := io.Pipe()
	io.TeeReader(io.TeeReader(pr, wrt), w)
	zipWriter := zip.NewWriter(pw)

	for _, file := range strings.Split(files[0], ",") {
		read, err := store.NewReader(file)
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer read.Close()

		f, err := zipWriter.Create(config.Bucket + "/" + file)
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(f, read); err != nil {
			panic(err)
		}

	}
	defer zipWriter.Close()

}

// main main function
func main() {

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
