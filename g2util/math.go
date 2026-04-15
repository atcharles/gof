package g2util

import (
	"cmp"
	"fmt"
	"math"
	"strconv"
)

// CeilFloat64 按小数位截取float 最大值
func CeilFloat64(val float64, precision int) float64 {
	output := math.Pow10(precision)
	return math.Ceil(val*output) / output
}

func Clamp[T cmp.Ordered](val, minVal, maxVal T) T {
	if val < minVal {
		return minVal
	}
	if val > maxVal {
		return maxVal
	}
	return val
}

// FloorFloat64 按小数位截取float 最小值
func FloorFloat64(val float64, precision int) float64 {
	output := math.Pow10(precision)
	return math.Floor(val*output) / output
}

// MustParsePreFloat64 按小数位截取float
func MustParsePreFloat64(value float64, pre int) float64 {
	f1 := fmt.Sprintf(fmt.Sprintf("%%.%df", pre), value)
	v, err := strconv.ParseFloat(f1, 64)
	if err != nil {
		return 0
	}
	return v
}

// RoundFloat64 按小数位截取float 四舍五入
func RoundFloat64(val float64, precision int) float64 {
	output := math.Pow10(precision)
	return math.Round(val*output) / output
}
