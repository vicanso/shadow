package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/chai2010/webp"
	"github.com/fogleman/primitive/primitive"
	"github.com/nfnt/resize"
)

// 获取query参数
func getQuery(query map[string][]string, key string) string {
	data := query[key]
	if data == nil {
		return ""
	}
	return strings.Join(data, "")
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

// 读取图像数据，根据请求的url或者base64数据
func getImage(req *http.Request) (image.Image, string, error) {
	query := req.URL.Query()

	file := getQuery(query, "file")
	if len(file) != 0 {
		f, err := os.Open(file)
		if err != nil {
			return nil, "", err
		}
		defer f.Close()
		return image.Decode(f)
	}

	url := getQuery(query, "url")
	if len(url) != 0 {
		c := &http.Client{
			Timeout: 10 * time.Second,
		}
		res, err := c.Get(url)
		if err != nil {
			return nil, "", err
		}
		defer res.Body.Close()
		buf, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, "", err
		}
		return image.Decode(bytes.NewReader(buf))
	}
	body, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		return nil, "", err
	}
	base64Data, err := jsonparser.GetString(body, "base64")
	if err != nil {
		return nil, "", err
	}
	buf, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, "", err
	}
	return image.Decode(bytes.NewReader(buf))
}

// 获取参数确认输出（图片或者base64）
func responseImage(w http.ResponseWriter, req *http.Request, img image.Image, imgType string) {
	// 图片数据可以缓存，设置客户端缓存 一年
	w.Header().Set("Cache-Control", "public, max-age=31536000, s-maxage=600")
	query := req.URL.Query()

	outputType := getQuery(query, "type")
	if len(outputType) == 0 {
		outputType = imgType
	}

	quality, _ := strconv.Atoi(getQuery(query, "quality"))

	buf := bytes.NewBuffer(nil) //开辟一个新的空buff
	var err error
	switch outputType {
	default:
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: quality})
	case "png":
		err = png.Encode(buf, img)
	case "webp":
		// 默认转换质量
		if quality == 0 {
			err = webp.Encode(buf, img, &webp.Options{Lossless: true})
		} else {
			err = webp.Encode(buf, img, &webp.Options{Lossless: false, Quality: float32(quality)})
		}
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "encode image data faial"}`))
		return
	}

	data := buf.Bytes()
	if getQuery(query, "output") != "base64" {
		w.Header().Set("Content-Type", "image/"+outputType)
		w.Write(data)
		return
	}

	base64Data := base64.StdEncoding.EncodeToString(data)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"base64":"` + base64Data + `", "type": "` + outputType + `"}`))
}

// 图片压缩处理（保持原有尺寸，调整质量）
func optimServe(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s %s", req.RemoteAddr, req.Method, req.URL)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	img, imgType, err := getImage(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "load image data faial"}`))
		return
	}

	origBounds := img.Bounds()
	// 对图片做压缩处理（原尺寸不变化 ）
	thumbnail := resize.Resize(uint(origBounds.Dx()), 0, img, resize.Lanczos3)
	responseImage(w, req, thumbnail, imgType)
}

// 生成阴影缩略图
func shadowServe(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s %s", req.RemoteAddr, req.Method, req.URL)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	query := req.URL.Query()
	width := getQuery(query, "width")
	height := getQuery(query, "height")
	// 模糊倍数
	times := getQuery(query, "times")
	if len(width) == 0 || len(height) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "width and height can't be null"}`))
		return
	}
	img, imgType, err := getImage(req)
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
	responseImage(w, req, thumbnail, imgType)
}

// 调整图像尺寸
func resizeServe(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s %s", req.RemoteAddr, req.Method, req.URL)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	query := req.URL.Query()
	width := getQuery(query, "width")
	height := getQuery(query, "height")
	if len(width) == 0 && len(height) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "both width and height are null"}`))
		return
	}
	img, imgType, err := getImage(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "load image data faial"}`))
		return
	}
	scaleWidth, _ := strconv.Atoi(width)
	scaleHeight, _ := strconv.Atoi(height)

	thumbnail := resize.Resize(uint(scaleWidth), uint(scaleHeight), img, resize.Lanczos3)
	responseImage(w, req, thumbnail, imgType)
}

func primitiveServe(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s %s", req.RemoteAddr, req.Method, req.URL)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	img, imgType, err := getImage(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "load image data faial"}`))
		return
	}
	query := req.URL.Query()
	width, _ := strconv.Atoi(getQuery(query, "width"))
	height, _ := strconv.Atoi(getQuery(query, "height"))
	times, _ := strconv.Atoi(getQuery(query, "times"))

	if width != 0 || height != 0 {
		img = resize.Resize(uint(width), uint(height), img, resize.Bilinear)
	}

	if times == 0 {
		times = 128
	}

	origBounds := img.Bounds()

	bg := primitive.MakeColor(primitive.AverageImageColor(img))

	model := primitive.NewModel(img, bg, origBounds.Dy(), runtime.NumCPU())
	for i := 0; i < times; i++ {

		// find optimal shape and add it to the model
		model.Step(primitive.ShapeType(1), 128, 0)
	}
	thumbnail := model.Context.Image()
	responseImage(w, req, thumbnail, imgType)

}

func pingServe(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("pong"))
}

func main() {
	http.HandleFunc("/ping", pingServe)
	http.HandleFunc("/@images/shadow", shadowServe)
	http.HandleFunc("/@images/resize", resizeServe)
	http.HandleFunc("/@images/optim", optimServe)
	http.HandleFunc("/@images/primitive", primitiveServe)

	log.Println("server is at :3015")
	if err := http.ListenAndServe(":3015", nil); err != nil {
		log.Fatal(err)
	}
}
