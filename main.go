package main

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/h2non/bimg"
)

//go:embed frontend/dist
var ffs embed.FS

const address = "localhost:12313"
const fileheapFolderName = "fileheap"

var formMaxSize = int64(math.Pow(2, 30))

func main() {
	http.HandleFunc("/api/thumbnail", func(w http.ResponseWriter, r *http.Request) {
		onerr := func(err error) {
			fmt.Printf("thumbnail#%s\n", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		err := r.ParseMultipartForm(formMaxSize)
		if err != nil {
			onerr(fmt.Errorf("parse multipart form: %w", err))
			return
		}

		keys := [4]string{"top", "left", "width", "height"}
		var extractRect [4]int
		for i, key := range keys {
			extractRect[i], err = strconv.Atoi(r.Form.Get(key))
			if err != nil {
				onerr(fmt.Errorf("bad value in \"%v\": %w", key, err))
				return
			}
		}

		file, _, err := r.FormFile("img")
		if err != nil {
			onerr(fmt.Errorf("get form file: %w", err))
			return
		}

		imgBuf, err := io.ReadAll(file)
		if err != nil {
			onerr(fmt.Errorf("read form image to buf: %w", err))
			return
		}

		finalImg, err := Thumbnail(imgBuf, ExtractRect{extractRect[0], extractRect[1], extractRect[2], extractRect[3]})
		if err != nil {
			onerr(fmt.Errorf("run thumbnailer: %w", err))
			return
		}

		imgId := time.Now().UnixMilli()
		err = bimg.Write(makeImgPath(imgId), finalImg)
		if err != nil {
			onerr(fmt.Errorf("write result: %w", err))
			return
		}

		fmt.Fprintf(w, "./api/thumbnail/get?id=%d", imgId)
	})

	http.HandleFunc("/api/thumbnail/get", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		imgId, err := strconv.ParseInt(r.Form.Get("id"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Disposition", "attachment")
		w.Header().Set("Content-Type", "image/jpeg")

		imgpath := makeImgPath(imgId)
		http.ServeFile(w, r, imgpath)

		os.Remove(imgpath)
	})

	dist, err := fs.Sub(ffs, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", http.FileServerFS(dist))

	err = os.Mkdir(fileheapFolderName, 0750)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		log.Fatal(err)
	}

	fmt.Printf("listen http://%s\n", address)
	log.Fatal(http.ListenAndServe(address, nil))
}

func makeImgPath(imgId int64) string {
	return fmt.Sprintf("%s/%d.jpg", fileheapFolderName, imgId)
}
