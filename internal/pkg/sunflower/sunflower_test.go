package sunflower

import (
	"math/rand"
	"pathtracer/internal/pkg/color"
	img "pathtracer/internal/pkg/image"
	"strconv"
	"testing"
	"time"
)

func Test_sunflower(t *testing.T) {
	t.Run("sunflower", func(t *testing.T) {
		width := 300
		height := 300
		amount := 4000
		randomize := true

		colors := []color.Color{
			{R: 1.0, G: 1.0, B: 1.0},
			{R: 1.0, G: 0.0, B: 1.0},
			{R: 1.0, G: 1.0, B: 0.0},
			{R: 0.0, G: 1.0, B: 1.0},
			{R: 1.0, G: 0.0, B: 0.0},
			{R: 0.0, G: 0.0, B: 1.0},
			{R: 0.0, G: 1.0, B: 0.0},
			{R: 1.0, G: 1.0, B: 1.0},
			{R: 1.0, G: 0.0, B: 1.0},
			{R: 1.0, G: 1.0, B: 0.0},
		}

		// ------------------------------------

		rand.Seed(time.Now().UnixMicro())

		halfWidth := float64(width / 2)
		halfHeight := float64(height / 2)

		image := img.NewFloatImage("sunflower", width, height)

		for i := 0; i < amount; i++ {
			//x, y := Sunflower(amount, 2.0, i+1, randomize)
			x, y := Sunflower(amount, 0.0, i+1, randomize)
			x2 := int(halfWidth * (1 + x))
			y2 := int(halfHeight * (1 - y))
			image.SetPixel(x2, y2, &colors[i*len(colors)/amount])
		}

		img.WriteImage("sunflower_["+strconv.Itoa(width)+"x"+strconv.Itoa(height)+"]x"+strconv.Itoa(amount)+"_random.png", image)

		//fmt.Printf("%+v\n", test)
	})
}
