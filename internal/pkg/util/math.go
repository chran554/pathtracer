package util

import (
	"github.com/ungerik/go3d/float64/vec3"
	"math"
)

const (
	radPerDeg = math.Pi / 180.0
)

func DegToRad(degrees float64) float64 {
	return radPerDeg * degrees
}

func Clamp(min float64, max float64, value float64) float64 {
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
