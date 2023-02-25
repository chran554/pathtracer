package util

func Clamp(min float64, max float64, value float64) float64 {
	if value < min {
		return min
	} else if value > max {
		return max
	} else {
		return value
	}
}
