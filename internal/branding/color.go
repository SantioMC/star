package branding

import (
	"math"
	"strconv"
)

func luminance(hex string) float64 {
	value, err := strconv.ParseUint(hex[1:], 16, 32)
	if err != nil {
		return 0
	}

	channels := []uint64{value >> 16 & 0xff, value >> 8 & 0xff, value & 0xff}
	weights := []float64{0.2126, 0.7152, 0.0722}

	total := 0.0
	for index, channel := range channels {
		srgb := float64(channel) / 255

		linear := srgb / 12.92
		if srgb > 0.04045 {
			linear = math.Pow((srgb+0.055)/1.055, 2.4)
		}

		total += linear * weights[index]
	}

	return total
}

func Foreground(hex string) string {
	background := luminance(hex)

	darkContrast := (background + 0.05) / (luminance("#111111") + 0.05)
	lightContrast := 1.05 / (background + 0.05)

	if darkContrast >= lightContrast {
		return "#111111"
	}

	return "#ffffff"
}
