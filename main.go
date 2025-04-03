package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"strconv"
	"strings"
	"syscall/js"
	"time"

	"github.com/egawata/color-tress-webui/tresser"
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
	btnGenerate      js.Value
	inputImage       js.Value
	outputImage      js.Value
	progress         js.Value
	download         js.Value
	radius           js.Value
	brightnessReduct js.Value
}

type app struct {
	el           dom
	tr           *tresser.Tresser
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
	a.el.brightnessReduct = js.Global().Get("document").Call("getElementById", "brightnessReduct")
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

	brightnessReduct, err := strconv.ParseFloat(a.el.brightnessReduct.Get("value").String(), 64)
	if err != nil {
		log.Printf("Invalid brightness reduction: %s", a.el.brightnessReduct.Get("value").String())
		return nil
	}
	if brightnessReduct < 0.0 || brightnessReduct > 1.0 {
		log.Printf("Invalid brightness reduction: %f", brightnessReduct)
		return nil
	}

	img, err := a.getInputImageData()
	if err != nil {
		log.Printf("Error getting input image data: %s", err)
		return nil
	}

	a.tr = tresser.NewTresser(img, pxRange, brightnessReduct)
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

func main() {
	a := app{}
	a.Start()

	<-make(chan struct{})
}
