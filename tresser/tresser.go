package tresser

import (
	"image"
	"image/color"
	"log"
	"time"

	"github.com/crazy3lf/colorconv"
)

type Tresser struct {
	x, y          int
	width, height int
	img           image.Image
	df            *darkestFinder
	pxRange       int
	brightnessReduct float64
	result        *image.RGBA
	timeout       time.Duration
	completed     bool
	cacheDarker   map[uint32][3]uint8
}

var resImg *image.RGBA

func NewTresser(i image.Image, pRange int, brightnessReduct float64) *Tresser {
	w := i.Bounds().Dx()
	h := i.Bounds().Dy()

	if resImg == nil || resImg.Bounds().Dx() != w || resImg.Bounds().Dy() != h {
		resImg = image.NewRGBA(i.Bounds())
	}

	return &Tresser{
		x:               0,
		y:               0,
		width:           w,
		height:          h,
		img:             i,
		pxRange:         pRange,
		brightnessReduct: brightnessReduct,
		df:              dFinder,
		timeout:         50 * time.Millisecond,
		result:          resImg,
		cacheDarker:     make(map[uint32][3]uint8),
	}
}

func (t *Tresser) SetTimeout(d time.Duration) {
	t.timeout = d
}

func (t *Tresser) GetResult() *image.RGBA {
	return t.result
}

func (t *Tresser) GetProgress() float32 {
	return float32(t.x+t.y*t.width) / float32(t.width*t.height) * 100.0
}

func (t *Tresser) IsCompleted() bool {
	return t.completed
}

func (t *Tresser) Continue() {
	start := time.Now()
	colorRGBA := color.RGBA{0, 0, 0, 255}
	
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

func (t *Tresser) modToDarkerColor(r, g, b uint8) (uint8, uint8, uint8) {
	if c, ok := t.cacheDarker[uint32(r)<<16|uint32(g)<<8|uint32(b)]; ok {
		return c[0], c[1], c[2]
	}

	var h, s, v float64
	h, s, v = colorconv.RGBToHSV(r, g, b)

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

	v -= t.brightnessReduct
	if v < 0.0 {
		v = 0.0
	}

	nr, ng, nb, err := colorconv.HSVToRGB(h, s, v)
	if err != nil {
		log.Printf("failed to convert HSV(%f, %f, %f) to RGB: %v", h, s, v, err)
		return r, g, b
	}
	t.cacheDarker[uint32(r)<<16|uint32(g)<<8|uint32(b)] = [3]uint8{nr, ng, nb}

	return nr, ng, nb
}
