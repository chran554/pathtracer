package sunflower

import (
	"math"
	"math/rand"
)

// Sunflower distributes n points evenly within a circle with radius 1.
// Parameter alpha controls point distribution on the edge. Typical values 1-2, higher values more points on the edge.
// The parameter pointNumber is the index of a point. It is in the range [1,n] .
// https://stackoverflow.com/questions/28567166/uniformly-distribute-x-points-inside-a-circle
func Sunflower(amountPoints int, alpha float64, pointNumber int, randomize bool) (x float64, y float64) { // example: amountPoints=500, alpha=2, pointNumber=[1..amountPoints]
	pointIndex := float64(pointNumber)
	if randomize {
		pointIndex += rand.Float64() - 0.5
	}

	b := math.Round(alpha * math.Sqrt(float64(amountPoints))) // number of boundary points
	phi := (math.Sqrt(5.0) + 1.0) / 2.0                       // golden ratio
	r := sunflowerRadius(pointIndex, float64(amountPoints), b)
	theta := 2.0 * math.Pi * float64(pointIndex) / (phi * phi)

	return r * math.Cos(theta), r * math.Sin(theta)
}

func sunflowerRadius(i float64, n float64, b float64) float64 {
	r := float64(1) // put on the boundary
	if i <= (n - b) {
		r = math.Sqrt(i-0.5) / math.Sqrt(n-(b+1.0)/2.0) // apply square root
	}
	return r
}
