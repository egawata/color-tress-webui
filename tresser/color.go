package tresser

import (
	"image"
)

type dotColor struct {
	r, g, b uint32
	bright  float64
}

type darkestFinder struct {
	darkest []*dotColor
}

const dfSize = 100

var dFinder = newDarkestFinder()
var mb = &dotColor{bright: 65536.0}

func newDarkestFinder() *darkestFinder {
	d := make([]*dotColor, dfSize)
	return &darkestFinder{darkest: d}
}

func (df *darkestFinder) GetDarkestColor(img image.Image, tx, ty, rng int) (uint8, uint8, uint8, uint8) {
	minBright := 65536.0
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	var rr, rg, rb uint32
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
