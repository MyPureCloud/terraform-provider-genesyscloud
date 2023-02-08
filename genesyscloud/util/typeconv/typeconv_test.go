package typeconv

import (
	"strconv"
	"testing"
)

func TestFloat32to64(t *testing.T) {
	float32Value := float32(0.666)
	float64ValueByUtil := Float32to64(&float32Value)
	float64ValueByConst := float64(float32Value)
	strFloat64ValueByUtil := strconv.FormatFloat(*float64ValueByUtil, 'f', -1, 64)
	strFloat64ValueByConst := strconv.FormatFloat(float64ValueByConst, 'f', -1, 64)

	if strFloat64ValueByUtil != "0.666" {
		t.Errorf("Fraction does not match with original by util conversion: %s", strFloat64ValueByUtil)
	}
	if strFloat64ValueByConst != "0.6660000085830688" {
		t.Errorf("Fraction does not match with 32 to 64 construction: %s", strFloat64ValueByConst)
	}
}
