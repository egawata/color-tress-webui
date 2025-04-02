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
	"time"

	"github.com/crazy3lf/colorconv"
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
	brightness  js.Value
}

type app struct {
	el           dom
	tr           *tresser
	prevProgress int
}

func (a *app) Start() {
	a.el.btnGenerate = js.Global().Get("document").Call("getElementById", "generate")
	a.el.btnGenerate.Call("addEventListener", "generate", js.FuncOf(a.generate))

	a.el.inputImage = js.Global().Get("document").Call("getElementById", "input-image")
	a.el.outputImage = js.Global().Get("document").Call("getElementById", "output-image")
	a.el.progress = js.Global().Get("document").Call("getElementById", "progress")
	a.el.download = js.Global().Get("document").Call("getElementById", "download")
	a.el.radius = js.Global().Get("document").Call("getElementById", "radius")
	a.el.brightness = js.Global().Get("document").Call("getElementById", "brightness")
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

	brightness, err := strconv.ParseFloat(a.el.brightness.Get("value").String(), 64)
	if err != nil {
		log.Printf("Invalid brightness: %s", a.el.brightness.Get("value").String())
		return nil
	}
	if brightness < 0.0 || brightness > 1.0 {
		log.Printf("Invalid brightness: %f", brightness)
		return nil
	}

	img, err := a.getInputImageData()
	if err != nil {
		log.Printf("Error getting input image data: %s", err)
		return nil
	}

	a.tr = newTresser(img, pxRange, brightness)
	a.tr.SetTimeout(100 * time.Millisecond)
	js.Global().Call("setTimeout", js.FuncOf(a.continueTress), 1)
	a.prevProgress = -1
	return nil
}

func (a *app) continueTress(this js.Value, p []js.Value) interface{} {
	if a.tr == nil {
		return nil
	}

	a.tr.Continue()
	if a.tr.IsCompleted() {
		var outBuf bytes.Buffer
		err := png.Encode(&outBuf, a.tr.GetResult())
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
		a.el.progress.Set("innerHTML", "Completed!")
	} else {
		pr := a.tr.GetProgress()
		a.el.progress.Set("innerHTML", fmt.Sprintf("%.2f %% completed...", pr))
		a.prevProgress = int(pr)
		js.Global().Call("setTimeout", js.FuncOf(a.continueTress), 1)
	}

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
	darkest []*dotColor
}

const dfSize = 100

var dFinder = newDarkestFinder()

func newDarkestFinder() *darkestFinder {
	d := make([]*dotColor, dfSize)
	return &darkestFinder{darkest: d}
}

var mb = &dotColor{bright: 65536.0}

// getDarkestColor は、画像 img の座標 (tx, ty) の周囲 rng ピクセルの中で最も暗い色を取得する。
// (tx - rng, ty - rng), (tx + rng, ty + rng) を対角線とする正方形内のピクセルが対象。
func (df *darkestFinder) GetDarkestColor(img image.Image, tx, ty, rng int) (uint8, uint8, uint8, uint8) {
	minBright := 65536.0
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	var rr, rg, rb uint32
	// 座標 tx, ty の周囲 rng ピクセルの中で最も暗い色を取得
	for x := tx - rng; x <= tx+rng; x++ {
		if x < 0 || x >= w {
			continue
		}

		indX := x % dfSize
		if x < tx {
			if dk := df.darkest[indX]; dk != nil {
				if dk.bright < minBright {
					minBright = dk.bright
					rr = dk.r
					rg = dk.g
					rb = dk.b
				}
				continue
			}
		}

		mb.bright = 65536.0
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
		if df.darkest[indX] == nil {
			df.darkest[indX] = &dotColor{}
		}
		df.darkest[indX].r = mb.r
		df.darkest[indX].g = mb.g
		df.darkest[indX].b = mb.b
		df.darkest[indX].bright = mb.bright
		if mb.bright < minBright {
			minBright = mb.bright
			rr = mb.r
			rg = mb.g
			rb = mb.b
		}
	}
	return uint8(rr >> 8), uint8(rg >> 8), uint8(rb >> 8), 255
}

func main() {
	a := app{}
	a.Start()

	<-make(chan struct{})
}

type tresser struct {
	x, y          int
	width, height int
	img           image.Image
	df            *darkestFinder
	pxRange       int
	brightness    float64
	result        *image.RGBA
	timeout       time.Duration
	completed     bool
}

var resImg *image.RGBA

func newTresser(i image.Image, pRange int, brightness float64) *tresser {
	w := i.Bounds().Dx()
	h := i.Bounds().Dy()

	// 結果格納用画像がすでに確保されていて、サイズが同じ場合は再利用する。
	// それ以外の場合は新規作成する。
	if resImg == nil || resImg.Bounds().Dx() != w || resImg.Bounds().Dy() != h {
		resImg = image.NewRGBA(i.Bounds())
	}

	return &tresser{
		x:          0,
		y:          0,
		width:      w,
		height:     h,
		img:        i,
		pxRange:    pRange,
		brightness: brightness,
		df:         dFinder,
		timeout:    50 * time.Millisecond,
		result:     resImg,
	}
}

func (t *tresser) SetTimeout(d time.Duration) {
	t.timeout = d
}

func (t *tresser) GetResult() *image.RGBA {
	return t.result
}

func (t *tresser) GetProgress() float32 {
	return float32(t.x+t.y*t.width) / float32(t.width*t.height) * 100.0
}

func (t *tresser) IsCompleted() bool {
	return t.completed
}

var colorRGBA = color.RGBA{0, 0, 0, 255}

func (t *tresser) Continue() {
	start := time.Now()
	for time.Since(start) < t.timeout {
		if t.y >= t.height {
			t.completed = true
			return
		}
		r, g, b, a := t.df.GetDarkestColor(t.img, t.x, t.y, t.pxRange)
		r, g, b = t.modToDarkerColor(r, g, b)
		colorRGBA.R = r
		colorRGBA.G = g
		colorRGBA.B = b
		colorRGBA.A = a

		t.result.SetRGBA(t.x, t.y, colorRGBA)
		t.x++
		if t.x >= t.width {
			t.x = 0
			t.y++
		}
	}
}

var cacheDarker = make(map[uint32][3]uint8)

func (t *tresser) modToDarkerColor(r, g, b uint8) (uint8, uint8, uint8) {
	if c, ok := cacheDarker[uint32(r)<<16|uint32(g)<<8|uint32(b)]; ok {
		return c[0], c[1], c[2]
	}

	var h, s, v float64
	h, s, v = colorconv.RGBToHSV(r, g, b)

	// 赤系の色は色相をマイナス方向に、青系の色はプラス方向にずらす
	if h < 60.0 || h > 240.0 {
		h -= 5.0
		if h < 0.0 {
			h += 359.9
		}
	} else {
		h += 5.0
	}

	if s > 0.01 {
		s = s + (1.0-s)/2.0
		if s > 0.99 {
			s = 0.99
		}
	}

	v -= t.brightness
	if v < 0.0 {
		v = 0.0
	}

	nr, ng, nb, err := colorconv.HSVToRGB(h, s, v)
	if err != nil {
		log.Printf("failed to convert HSV(%f, %f, %f) to RGB: %v", h, s, v, err)
		return r, g, b
	}
	cacheDarker[uint32(r)<<16|uint32(g)<<8|uint32(b)] = [3]uint8{nr, ng, nb}

	return nr, ng, nb
}
