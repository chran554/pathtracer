package sunflower

import (
	"fmt"
	"math"
	"math/rand"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"strconv"
	"testing"
	"time"
)

func Test_sunflower(t *testing.T) {
	width := 300
	height := 300
	amount := 4000
	randomize := true

	colors := []*color.Color{
		color.NewColor(1.0, 1.0, 1.0),
		color.NewColor(1.0, 0.0, 1.0),
		color.NewColor(1.0, 1.0, 0.0),
		color.NewColor(0.0, 1.0, 1.0),
		color.NewColor(1.0, 0.0, 0.0),
		color.NewColor(0.0, 0.0, 1.0),
		color.NewColor(0.0, 1.0, 0.0),
		color.NewColor(1.0, 1.0, 1.0),
		color.NewColor(1.0, 0.0, 1.0),
		color.NewColor(1.0, 1.0, 0.0),
	}

	// ------------------------------------

	rand.Seed(time.Now().UnixMicro())

	halfWidth := float64(width / 2)
	halfHeight := float64(height / 2)

	image := floatimage.NewFloatImage("sunflower", width, height)

	for i := range amount {
		x, y := Sunflower(amount, 0.0, i+1, randomize)
		x2 := int(halfWidth * (1 + x))
		y2 := int(halfHeight * (1 - y))
		c := colors[i*len(colors)/amount]
		image.SetPixel(x2, y2, c)
	}

	floatimage.WriteImage("sunflower_["+strconv.Itoa(width)+"x"+strconv.Itoa(height)+"]x"+strconv.Itoa(amount)+"_random.png", image)

	//fmt.Printf("%+v\n", test)
}

func TestSunflowerDeterministicLayout(t *testing.T) {
	// Reference coordinates from the original, non-randomized sunflower layout.
	tests := []struct {
		amountPoints int
		alpha        float64
		pointNumber  int
		x, y         float64
	}{
		{1, 0, 1, -0.73736887807831974, 0.67549029426152385},
		{2, 0, 1, -0.42572012025057032, 0.38999450322687035},
		{2, 0, 2, 0.087425724716959877, -0.99617104086482777},
		{100, 0, 50, 0.57501648820511009, 0.40846477874866044},
		{100, 0, 100, 0.32925552282504578, 0.94424085946817926},
		{100, 1, 90, -0.69648710844430983, 0.67970262237307233},
		{100, 1, 91, 0.055935271716783613, -0.9984343971328159},
		{100, 2, 80, -0.88209487888149674, -0.33192887954442651},
		{100, 2, 81, 0.92802419328067165, -0.37251992790421101},
		{1, 2, 1, -0.73736887807831974, 0.67549029426152385},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("n=%d/alpha=%g/k=%d", tt.amountPoints, tt.alpha, tt.pointNumber), func(t *testing.T) {
			x, y := Sunflower(tt.amountPoints, tt.alpha, tt.pointNumber, false)
			if math.IsNaN(x) || math.IsNaN(y) || math.Abs(x-tt.x) > 1e-13 || math.Abs(y-tt.y) > 1e-13 {
				t.Fatalf("got (%g, %g), want (%g, %g)", x, y, tt.x, tt.y)
			}
		})
	}
}

func TestSunflowerRandomizedUniformDisk(t *testing.T) {
	// Equal radial-area bins crossed with equal angular sectors detect both
	// radial bias and correlation between radius and angle. Each fixed index
	// must cover the whole disk, including the single-point and last-point cases.
	const samples = 65536
	const radialBins = 4
	const angularBins = 8

	tests := []struct {
		amountPoints int
		pointNumber  int
	}{
		{1, 1},
		{10, 1},
		{10, 10},
		{100000, 100000},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("n=%d/k=%d", tt.amountPoints, tt.pointNumber), func(t *testing.T) {
			var counts [radialBins][angularBins]int
			var sumX, sumY, sumRadiusSquared float64
			boundarySamples := 0

			for range samples {
				x, y := Sunflower(tt.amountPoints, 0, tt.pointNumber, true)
				radiusSquared := x*x + y*y

				if math.IsNaN(radiusSquared) || math.IsInf(radiusSquared, 0) || radiusSquared > 1+1e-14 {
					t.Fatalf("invalid disk sample (%g, %g)", x, y)
				}

				if radiusSquared >= 1-1e-12 {
					boundarySamples++
				}

				angle := math.Atan2(y, x)
				if angle < 0 {
					angle += 2 * math.Pi
				}

				// Allow floating-point rounding at the upper endpoints.
				radialBin := min(int(radiusSquared*radialBins), radialBins-1)
				angularBin := min(int(angle/(2*math.Pi)*angularBins), angularBins-1)
				counts[radialBin][angularBin]++
				sumX += x
				sumY += y
				sumRadiusSquared += radiusSquared
			}

			// A narrow rim has negligible area; tolerate rounding, but not a
			// discrete probability mass on the boundary.
			if boundarySamples > 2 {
				t.Errorf("%d of %d samples landed on the boundary", boundarySamples, samples)
			}
			if math.Abs(sumX/samples) > 0.025 || math.Abs(sumY/samples) > 0.025 {
				t.Errorf("sample centroid (%g, %g) is not near the origin", sumX/samples, sumY/samples)
			}
			if math.Abs(sumRadiusSquared/samples-0.5) > 0.01 {
				t.Errorf("mean radius squared is %g, want approximately 0.5", sumRadiusSquared/samples)
			}

			const expected = samples / (radialBins * angularBins)
			// This tolerance exceeds eleven standard deviations per bin for
			// uniform samples, while catching the old restricted spiral arcs.
			for r := range radialBins {
				for a := range angularBins {
					if math.Abs(float64(counts[r][a]-expected)) > expected/4 {
						t.Errorf("area bin [%d][%d] has %d samples, want approximately %d", r, a, counts[r][a], expected)
					}
				}
			}
		})
	}
}

func TestSunflowerRandomizedPositiveAlphaBoundary(t *testing.T) {
	// Positive alpha still deliberately places the last seeds on the rim.
	for _, alpha := range []float64{1, 2} {
		for range 1000 {
			x, y := Sunflower(100, alpha, 100, true)
			if math.IsNaN(x) || math.IsNaN(y) || math.Abs(x*x+y*y-1) > 1e-14 {
				t.Fatalf("alpha=%g: boundary sample (%g, %g) is not on the unit circle", alpha, x, y)
			}
		}
	}
}
