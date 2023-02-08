package typeconv

import (
	"fmt"
	"strconv"
)

func Float32to64(float32Value *float32) *float64 {
	if float32Value == nil {
		return nil
	}

	floatString := fmt.Sprintf("%f", *float32Value)
	float64Value, _ := strconv.ParseFloat(floatString, 64)
	return &float64Value
}

func Float64to32(float64Value *float64) *float32 {
	if float64Value == nil {
		return nil
	}

	float32Value := float32(*float64Value)
	return &float32Value
}
