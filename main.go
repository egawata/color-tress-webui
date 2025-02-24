package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"strconv"
	"strings"
	"syscall/js"
)

type imgFormat int

const (
	imgFormatUnknown imgFormat = iota
	imgFormatJPEG
	imgFormatPNG
)

func (f imgFormat) String() string {
	switch f {
	case imgFormatJPEG:
		return "JPEG"
	case imgFormatPNG:
		return "PNG"
	default:
		return "Unknown"
	}
}

type dom struct {
	btnGenerate js.Value
	inputImage  js.Value
	outputImage js.Value
	progress    js.Value
	download    js.Value
	radius      js.Value
}

type app struct {
	el dom
}

func (a *app) Start() {
	a.el.btnGenerate = js.Global().Get("document").Call("getElementById", "generate")
	a.el.btnGenerate.Call("addEventListener", "generate", js.FuncOf(a.generate))

	a.el.inputImage = js.Global().Get("document").Call("getElementById", "input-image")
	a.el.outputImage = js.Global().Get("document").Call("getElementById", "output-image")
	a.el.progress = js.Global().Get("document").Call("getElementById", "progress")
	a.el.download = js.Global().Get("document").Call("getElementById", "download")
	a.el.radius = js.Global().Get("document").Call("getElementById", "radius")
}

func (a *app) generate(this js.Value, p []js.Value) interface{} {
	pxRange, err := strconv.Atoi(a.el.radius.Get("value").String())
	if err != nil {
		log.Printf("Invalid radius: %s", a.el.radius.Get("value").String())
		return nil
	}
	if pxRange < 1 || pxRange > 100 {
		log.Printf("Invalid radius: %d", pxRange)
		return nil
	}

	img, err := a.getInputImageData()
	if err != nil {
		log.Printf("Error getting input image data: %s", err)
		return nil
	}

	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	resImage := image.NewRGBA(image.Rect(0, 0, width, height))

	totalPx := width * height
	var prevCompleted int = 0

	df := newDarkestFinder(width, height)

	for x := range width {
		for y := range height {
			resImage.SetRGBA(x, y, df.GetDarkestColor(img, x, y, pxRange))
			completed := int(float32(x*height+y) / float32(totalPx) * 100.0)
			if prevCompleted < completed {
				if completed%10 == 0 {
					log.Printf("[PROGRESS]%d %% completed...", completed)
				}
				//a.el.progress.Set("innerHTML", fmt.Sprintf("%d %% completed...", completed))
				//fmt.Printf("\r[%-50s] %d%%", strings.Repeat("#", completed/2), completed)
				prevCompleted = completed
			}
		}
	}

	log.Printf("Generating image...")
	var outBuf bytes.Buffer
	err = png.Encode(&outBuf, resImage)
	if err != nil {
		log.Fatalf("failed to encode image: %v", err)
		return nil
	}
	outBase64Data := base64.StdEncoding.EncodeToString(outBuf.Bytes())
	outSrc := "data:image/png;base64," + outBase64Data
	a.el.outputImage.Set("src", outSrc)
	a.el.outputImage.Set("width", a.el.inputImage.Get("width"))
	a.el.outputImage.Set("height", a.el.inputImage.Get("height"))
	a.el.download.Set("disabled", false)

	return nil
}

// getInputImageData は、input-image img タグ内にロードされている画像データを取得する。
func (a *app) getInputImageData() (image.Image, error) {
	data := a.el.inputImage.Get("src").String()

	dArray := strings.Split(data, ",")
	if len(dArray) != 2 {
		return nil, fmt.Errorf("Invalid data URL: %s", data)
	}

	iFmt := dArray[0]
	var format imgFormat
	if strings.Contains(iFmt, "image/jpeg") {
		format = imgFormatJPEG
	} else if strings.Contains(iFmt, "image/png") {
		format = imgFormatPNG
	} else {
		return nil, fmt.Errorf("Unsupported format: %s", format)
	}

	encoded := dArray[1]
	//log.Printf("encoded: %s", encoded)

	imgData, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("Error decoding base64: %s", err)
	}

	var img image.Image
	switch format {
	case imgFormatJPEG:
		img, err = jpeg.Decode(bytes.NewReader(imgData))
	case imgFormatPNG:
		img, err = png.Decode(bytes.NewReader(imgData))
	default:
		return nil, fmt.Errorf("DONOTREACH")
	}
	if err != nil {
		return nil, fmt.Errorf("Error decoding image as %s: %s", format, err)
	}

	return img, nil
}

type dotColor struct {
	r, g, b uint32
	bright  float64
}

type darkestFinder struct {
	darkest [][]*dotColor
}

func newDarkestFinder(w, h int) *darkestFinder {
	d := make([][]*dotColor, w)
	for i := range d {
		d[i] = make([]*dotColor, h)
	}
	return &darkestFinder{darkest: d}
}

// getDarkestColor は、画像 img の座標 (tx, ty) の周囲 rng ピクセルの中で最も暗い色を取得する。
// (tx - rng, ty - rng), (tx + rng, ty + rng) を対角線とする正方形内のピクセルが対象。
func (df *darkestFinder) GetDarkestColor(img image.Image, tx, ty, rng int) color.RGBA {
	minBright := 65536.0
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	var rr, rg, rb uint32
	// 座標 tx, ty の周囲 rng ピクセルの中で最も暗い色を取得
	for x := tx - rng; x <= tx+rng; x++ {
		if x < 0 || x >= w {
			continue
		}

		if dk := df.darkest[x][ty]; dk != nil {
			if df.darkest[x][ty].bright < minBright {
				minBright = dk.bright
				rr = dk.r
				rg = dk.g
				rb = dk.b
			}
			continue
		}

		mb := &dotColor{bright: 65536.0}
		for y := ty - rng; y <= ty+rng; y++ {
			if y < 0 || y >= h {
				continue
			}
			r, g, b, _ := img.At(x, y).RGBA()
			bright := 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
			if bright < mb.bright {
				mb.bright = bright
				mb.r = r
				mb.g = g
				mb.b = b
			}
		}
		df.darkest[x][ty] = mb
		if mb.bright < minBright {
			minBright = mb.bright
			rr = mb.r
			rg = mb.g
			rb = mb.b
		}
	}
	return color.RGBA{uint8(rr >> 8), uint8(rg >> 8), uint8(rb >> 8), 255}
}

func main() {
	a := app{}
	a.Start()

	<-make(chan struct{})
}
