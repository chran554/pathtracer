package util

import (
	"math"

	"github.com/ungerik/go3d/float64/vec3"
)

const (
	radPerDeg = math.Pi / 180.0
	degPerRad = 180.0 / math.Pi
)

func DegToRad(degrees float64) float64 {
	return radPerDeg * degrees
}

func RadToDeg(radians float64) float64 {
	return degPerRad * radians
}

func ClampFloat64(min float64, max float64, value float64) float64 {
	if value < min {
		return min
	} else if value > max {
		return max
	} else {
		return value
	}
}

func ClampInt(min int, max int, value int) int {
	if value < min {
		return min
	} else if value > max {
		return max
	} else {
		return value
	}
}

func Cosine(a *vec3.T, b *vec3.T) float64 {
	return vec3.Dot(a, b) / math.Sqrt(a.LengthSqr()*b.LengthSqr())
}

func CosinePositive(a *vec3.T, b *vec3.T) bool {
	return vec3.Dot(a, b) >= 0
}

func CosineNegative(a *vec3.T, b *vec3.T) bool {
	return vec3.Dot(a, b) < 0
}

func Max32(a, b float32) float32 {
	if a >= b {
		return a
	}
	return b
}

func Min32(a, b float32) float32 {
	if a <= b {
		return a
	}
	return b
}

func MaxInt(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

func MinInt(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
