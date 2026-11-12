package sunflower

import (
	"math"
	"math/rand"
)

// Sunflower returns a point in a disk with radius 1. amountPoints must be positive,
// and pointNumber is a one-based index in [1, amountPoints].
//
// With randomize=false, points follow the deterministic sunflower layout.
// Parameter alpha controls the number of points placed on the boundary;
// typical values are 0 or 1-2, with larger values giving more weight to the edge.
//
// With randomize=true and alpha=0, each call samples uniformly over disk area,
// independently of amountPoints and pointNumber. Otherwise, randomize jitters
// the seed index by up to half a step along the boundary-weighted sunflower layout.
//
// https://stackoverflow.com/questions/28567166/uniformly-distribute-x-points-inside-a-circle
func Sunflower(amountPoints int, alpha float64, pointNumber int, randomize bool) (x float64, y float64) { // example: amountPoints=500, alpha=2, pointNumber=[1..amountPoints]
	if randomize && alpha == 0 {
		// Disk area is proportional to radius squared. Sample it independently
		// of the angle to avoid both boundary bias and spiral correlations.
		r := math.Sqrt(rand.Float64())
		theta := 2.0 * math.Pi * rand.Float64()
		return r * math.Cos(theta), r * math.Sin(theta)
	}

	pointIndex := float64(pointNumber)
	if randomize {
		pointIndex += rand.Float64() - 0.5
	}

	numberOfBoundaryPoints := 0.0
	if alpha > 0.0 {
		numberOfBoundaryPoints = math.Round(alpha * math.Sqrt(float64(amountPoints))) // number of boundary points
	}

	phi := (math.Sqrt(5.0) + 1.0) / 2.0 // golden ratio
	r := sunflowerRadius(pointIndex, float64(amountPoints), numberOfBoundaryPoints)
	theta := 2.0 * math.Pi * pointIndex / (phi * phi)

	return r * math.Cos(theta), r * math.Sin(theta)
}

func sunflowerRadius(seedIndex float64, amountSeeds float64, numberOfBoundaryPoints float64) float64 {
	r := float64(1) // put on the boundary

	if seedIndex <= (amountSeeds - numberOfBoundaryPoints) {
		r = math.Sqrt(seedIndex-0.5) / math.Sqrt(amountSeeds-(numberOfBoundaryPoints+1.0)/2.0)
	}

	return r
}
