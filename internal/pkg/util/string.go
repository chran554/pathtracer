package util

import "fmt"

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func FormatInt(i int) string {
	separator := byte(' ')

	// Format integer with thousands separators without using external libs.
	// Works for negative numbers and arbitrarily large magnitudes of int.
	if i == 0 {
		return "0"
	}

	neg := i < 0

	// Compute absolute value safely without overflowing on MinInt by using the
	// trick: abs(i) = uint64(-(i+1)) + 1 when i < 0.
	ii := int64(i)
	var n uint64
	if ii < 0 {
		n = uint64(-(ii + 1)) + 1
	} else {
		n = uint64(ii)
	}

	// Build the string in reverse using a fixed-size buffer (int up to 64 bits -> max 20 digits)
	var buf [32]byte // room for digits, separators, and sign
	idx := len(buf)
	groupCount := 0
	for n > 0 {
		if groupCount == 3 {
			idx--
			buf[idx] = separator
			groupCount = 0
		}
		idx--
		buf[idx] = byte('0' + (n % 10))
		n /= 10
		groupCount++
	}

	if neg {
		idx--
		buf[idx] = '-'
	}

	return string(buf[idx:])
}
