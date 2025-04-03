package tresser

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createSolidColorImage(width, height int, r, g, b uint8) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	col := color.RGBA{r, g, b, 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, col)
		}
	}
	return img
}

func createGradientImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			val := uint8(255 - (x * 255 / width))
			img.SetRGBA(x, y, color.RGBA{val, val, val, 255})
		}
	}
	return img
}

func createMultiColorImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	centerX := width / 2
	centerY := height / 2

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b := uint8(200), uint8(200), uint8(200)

			if x >= centerX-2 && x <= centerX+2 &&
				y >= centerY-2 && y <= centerY+2 {
				r, g, b = 50, 30, 40
			}

			img.SetRGBA(x, y, color.RGBA{r, g, b, 255})
		}
	}
	return img
}

func TestGetDarkestColorSolidImage(t *testing.T) {
	img := createSolidColorImage(10, 10, 100, 150, 200)

	df := newDarkestFinder()

	r, g, b, a := df.GetDarkestColor(img, 5, 5, 3)

	assert.Equal(t, uint8(100), r, "赤色の値が期待通りではありません")
	assert.Equal(t, uint8(150), g, "緑色の値が期待通りではありません")
	assert.Equal(t, uint8(200), b, "青色の値が期待通りではありません")
	assert.Equal(t, uint8(255), a, "アルファ値が期待通りではありません")
}
func TestGetDarkestColorMultiColor(t *testing.T) {
	img := createMultiColorImage(10, 10)

	df := newDarkestFinder()

	r, g, b, a := df.GetDarkestColor(img, 5, 5, 3)

	assert.Equal(t, uint8(50), r, "赤色の値が期待通りではありません")
	assert.Equal(t, uint8(30), g, "緑色の値が期待通りではありません")
	assert.Equal(t, uint8(40), b, "青色の値が期待通りではありません")
	assert.Equal(t, uint8(255), a, "アルファ値が期待通りではありません")
}

func TestGetDarkestColorBoundary(t *testing.T) {
	img := createGradientImage(10, 10)

	testCases := []struct {
		name     string
		x, y     int
		rng      int
		expected uint8
	}{
		{"左端", 0, 5, 2, 204}, // 左端なので範囲内（0-2）で最も暗いのは x=2 の位置
		{"右端", 9, 5, 2, 26},  // 右端なので範囲内（7-9）で最も暗いのは x=9 の位置
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			df := newDarkestFinder()
			r, _, _, _ := df.GetDarkestColor(img, tc.x, tc.y, tc.rng)
			assert.Equal(t, tc.expected, r, "期待される暗さと異なります")
		})
	}
}

func TestGetDarkestColorDifferentPositions(t *testing.T) {
	img := createMultiColorImage(10, 10)

	txTestCases := []struct {
		name string
		tx   int
		ty   int
		rng  int
	}{
		{"tx=1,ty=5", 1, 5, 3},
		{"tx=5,ty=5", 5, 5, 3},
		{"tx=8,ty=5", 8, 5, 3},
	}

	for _, tc := range txTestCases {
		t.Run(tc.name, func(t *testing.T) {
			df := newDarkestFinder()
			r, g, b, a := df.GetDarkestColor(img, tc.tx, tc.ty, tc.rng)

			if tc.tx == 5 && tc.ty == 5 {
				assert.Equal(t, uint8(50), r, "中央位置での赤色の値が期待通りではありません")
				assert.Equal(t, uint8(30), g, "中央位置での緑色の値が期待通りではありません")
				assert.Equal(t, uint8(40), b, "中央位置での青色の値が期待通りではありません")
			} else {
				if r == 50 {
					assert.Equal(t, uint8(30), g, "暗い色の緑色の値が期待通りではありません")
					assert.Equal(t, uint8(40), b, "暗い色の青色の値が期待通りではありません")
				} else {
					assert.Equal(t, uint8(200), r, "明るい色の赤色の値が期待通りではありません")
					assert.Equal(t, uint8(200), g, "明るい色の緑色の値が期待通りではありません")
					assert.Equal(t, uint8(200), b, "明るい色の青色の値が期待通りではありません")
				}
			}
			assert.Equal(t, uint8(255), a, "アルファ値が期待通りではありません")
		})
	}

	tyTestCases := []struct {
		name string
		tx   int
		ty   int
		rng  int
	}{
		{"tx=5,ty=1", 5, 1, 3},
		{"tx=5,ty=8", 5, 8, 3},
	}

	for _, tc := range tyTestCases {
		t.Run(tc.name, func(t *testing.T) {
			df := newDarkestFinder()
			r, g, b, a := df.GetDarkestColor(img, tc.tx, tc.ty, tc.rng)

			if r == 50 {
				assert.Equal(t, uint8(30), g, "暗い色の緑色の値が期待通りではありません")
				assert.Equal(t, uint8(40), b, "暗い色の青色の値が期待通りではありません")
			} else {
				assert.Equal(t, uint8(200), r, "明るい色の赤色の値が期待通りではありません")
				assert.Equal(t, uint8(200), g, "明るい色の緑色の値が期待通りではありません")
				assert.Equal(t, uint8(200), b, "明るい色の青色の値が期待通りではありません")
			}
			assert.Equal(t, uint8(255), a, "アルファ値が期待通りではありません")
		})
	}
}

func TestGetDarkestColorCaching(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 2))

	for x := 0; x < 10; x++ {
		val := uint8(100 + x*15)
		img.SetRGBA(x, 0, color.RGBA{val, val, val, 255})
	}

	for x := 0; x < 10; x++ {
		img.SetRGBA(x, 1, color.RGBA{50, 50, 50, 255})
	}

	df := newDarkestFinder()

	df.GetDarkestColor(img, 5, 1, 2)

	r, g, b, _ := df.GetDarkestColor(img, 6, 1, 2)

	assert.Equal(t, uint8(50), r, "キャッシュされた赤色の値が期待通りではありません")
	assert.Equal(t, uint8(50), g, "キャッシュされた緑色の値が期待通りではありません")
	assert.Equal(t, uint8(50), b, "キャッシュされた青色の値が期待通りではありません")
}
