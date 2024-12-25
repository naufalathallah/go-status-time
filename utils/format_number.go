package utils

import "strconv"

func formatNumber(value float64) interface{} {
	if value == 0 {
		return "-"
	}
	
	return strconv.FormatFloat(value, 'f', 2, 64)
}
