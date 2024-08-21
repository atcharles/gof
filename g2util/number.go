package g2util

import (
	"strconv"
)

// FloatPrecision 浮点型精度包装
func FloatPrecision(f *float64, p int) float64 {
	s1 := strconv.FormatFloat(*f, 'f', p, 64)
	*f, _ = strconv.ParseFloat(s1, 64)
	return *f
}
