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
		onerr := func(err error, where string) {
			fmt.Printf("thumbnail#%s# %s\n", where, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		err := r.ParseMultipartForm(formMaxSize)
		if err != nil {
			onerr(err, "parse multipart form")
			return
		}

		keys := [4]string{"top", "left", "width", "height"}
		var extractRect [4]int
		for i, key := range keys {
			extractRect[i], err = strconv.Atoi(r.Form.Get(key))
			if err != nil {
				onerr(err, "bad value in "+key)
				return
			}
		}

		file, _, err := r.FormFile("img")
		if err != nil {
			onerr(err, "get form file")
			return
		}

		// начало обработки изображения
		buffer, err := io.ReadAll(file)
		if err != nil {
			onerr(err, "io read all")
			return
		}

		baseImg, err := bimg.NewImage(buffer).Extract(extractRect[0], extractRect[1], extractRect[2], extractRect[3])
		if err != nil {
			onerr(err, "bimg extract")
			return
		}

		fg := bimg.NewImage(baseImg)

		size, err := fg.Size()
		if err != nil {
			onerr(err, "fg size")
			return
		}

		width := float64(size.Width)
		height := float64(size.Height)
		top := 0.
		left := 0.

		if width > height {
			height = height * (400 / width)
			width = 400
			top = (300 - height) / 2
		} else {
			width = width * (300 / height)
			height = 300
			left = (400 - width) / 2
		}

		wm, err := fg.Process(bimg.Options{
			Width:   int(width),
			Height:  int(height),
			Enlarge: true,
		})
		if err != nil {
			onerr(err, "fg resize")
			return
		}

		bg := bimg.NewImage(baseImg)

		finalImg, err := bg.Process(bimg.Options{
			Width:        400,
			Height:       300,
			Gravity:      bimg.GravityCentre,
			GaussianBlur: bimg.GaussianBlur{Sigma: 15},
			Crop:         true,
			Quality:      95,
			Enlarge:      true,
			WatermarkImage: bimg.WatermarkImage{
				Left: int(left),
				Top:  int(top),
				Buf:  wm,
			},
		})
		if err != nil {
			onerr(err, "bg process")
			return
		}

		imgId := time.Now().UnixMilli()
		err = bimg.Write(makeImgPath(imgId), finalImg)
		if err != nil {
			onerr(err, "write result")
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
