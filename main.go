package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/nfnt/resize"
)

// 获取图片数据
func loadImage(path string) (img image.Image, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return jpeg.Decode(file)
}

func loadCover(author, name string) (img image.Image, err error) {
	// file := "/covers/" + author + "/" + name + "/cover.jpg"
	file := "/Users/xieshuzhou/tmp/" + author + "/" + name + "/cover.jpg"
	return loadImage(file)
}

// 图片转换为[]bytes
func jpegToBytes(img image.Image) []byte {
	buf := bytes.NewBuffer(nil) //开辟一个新的空buff

	jpeg.Encode(buf, img, nil) //写入buffer
	return buf.Bytes()
}

// 图片转换为base64
func jpegToBase64(img image.Image) string {
	return base64.StdEncoding.EncodeToString(jpegToBytes(img))
}

// 对图片进行缩小处理
func scaleImage(img image.Image, width int, height int, times int) image.Image {
	bound := img.Bounds()
	dx := bound.Dx()
	// 先做缩小
	thumbnail := resize.Resize(uint(dx/times), 0, img, resize.Lanczos3)
	// 再做放大
	newImg := resize.Resize(uint(width), 0, thumbnail, resize.Lanczos3)

	bound = newImg.Bounds()
	dx = bound.Dx()
	dy := bound.Dy()

	rgba := image.NewRGBA(image.Rect(0, 0, width, height))

	// 计算要显示的尺寸大小
	offsetY := 0
	if dy > height {
		offsetY = (dy - height) / 2
	}
	maxY := dy - offsetY
	if maxY > height {
		maxY = height
	}
	// 将象素一下复制
	for x := 0; x < dx; x++ {
		for y := 0; y < maxY; y++ {
			pixel := newImg.At(x, y+offsetY)
			rgba.Set(x, y, pixel)
		}
	}
	return rgba
}

func pingServe(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("pong"))
}

func getQuery(query map[string][]string, key string) string {
	data := query[key]
	if data == nil {
		return ""
	}
	return strings.Join(data, "")
}

// 根据封面生成
func shadowServe(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s %s", req.RemoteAddr, req.Method, req.URL)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	query := req.URL.Query()
	author := getQuery(query, "author")
	name := getQuery(query, "name")
	width := getQuery(query, "width")
	height := getQuery(query, "height")
	// 模糊倍数
	times := getQuery(query, "times")
	if len(author) == 0 || len(name) == 0 || len(width) == 0 || len(height) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "author, name, width and height can't be null"}`))
		return
	}

	img, err := loadCover(author, name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "load image data faial"}`))
		return
	}

	scaleWidth, err := strconv.Atoi(width)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "width isn't a number"}`))
		return
	}

	scaleHeight, err := strconv.Atoi(height)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "height isn't a number"}`))
		return
	}

	scaleTimes := 15
	if len(times) != 0 {
		v, _ := strconv.Atoi(times)
		if v > 0 {
			scaleTimes = v
		}
	}

	thumbnail := scaleImage(img, scaleWidth, scaleHeight, scaleTimes)

	if getQuery(query, "type") == "image" {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(jpegToBytes(thumbnail))
		return
	}

	base64Data := jpegToBase64(thumbnail)
	resStr := strings.Replace(`{"base64":"${1}", "type": "jpeg"}`, "${1}", base64Data, 1)
	w.Write([]byte(resStr))
}

func resizeServe(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s %s", req.RemoteAddr, req.Method, req.URL)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	query := req.URL.Query()
	author := getQuery(query, "author")
	name := getQuery(query, "name")
	width := getQuery(query, "width")
	height := getQuery(query, "height")
	if len(author) == 0 || len(name) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "author and name can't be null"}`))
		return
	}
	if len(width) == 0 && len(height) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "both width and height are null"}`))
		return
	}

	img, err := loadCover(author, name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "load image data faial"}`))
		return
	}

	scaleWidth, _ := strconv.Atoi(width)
	scaleHeight, _ := strconv.Atoi(height)

	thumbnail := resize.Resize(uint(scaleWidth), uint(scaleHeight), img, resize.Lanczos3)

	if getQuery(query, "type") == "image" {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(jpegToBytes(thumbnail))
		return
	}

	base64Data := jpegToBase64(thumbnail)
	resStr := strings.Replace(`{"base64":"${1}", "type": "jpeg"}`, "${1}", base64Data, 1)
	w.Write([]byte(resStr))
}

func main() {
	http.HandleFunc("/ping", pingServe)
	http.HandleFunc("/@images/shadow", shadowServe)
	http.HandleFunc("/@images/resize", resizeServe)

	log.Println("server is at :3015")
	if err := http.ListenAndServe(":3015", nil); err != nil {
		log.Fatal(err)
	}
}
