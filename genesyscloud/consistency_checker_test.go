package genesyscloud

import (
	"testing"
)

func TestSliceEqual(t *testing.T) {
	cc := &consistencyCheck{}
	m := make(map[string]interface{})
	m["role_id"] = "role_id_1"
	m["division_id"] = "division_id_1"

	m1 := make(map[string]interface{})
	m1["role_id"] = "role_id_2"
	m1["division_id"] = "division_id_2"

	m2 := make(map[string]interface{})
	m2["role_id"] = "role_id_1"
	m2["division_id"] = "division_id_2"

	m3 := make(map[string]interface{})
	m3["role_id"] = "role_id_2"
	m3["division_id"] = "division_id_1"

	m4 := make(map[string]interface{})
	m4["role_id"] = "role_id_3"
	m4["division_id"] = "division_id_3"

	// maps

	if !cc.sliceEqual([]interface{}{m, m1}, []interface{}{m, m1}) {
		t.Fatalf("Should be true")
	}

	if !cc.sliceEqual([]interface{}{m1, m}, []interface{}{m1, m}) {
		t.Fatalf("Should be true")
	}

	if !cc.sliceEqual([]interface{}{m, m1}, []interface{}{m1, m}) {
		t.Fatalf("Should be true")
	}

	if !cc.sliceEqual([]interface{}{m, m1, m2}, []interface{}{m1, m2, m}) {
		t.Fatalf("Should be true")
	}

	if cc.sliceEqual([]interface{}{m, m1, m2}, []interface{}{m1, m2, m4}) {
		t.Fatalf("Should be false")
	}

	if cc.sliceEqual([]interface{}{m, m1}, []interface{}{m1}) {
		t.Fatalf("Should be false")
	}

	if cc.sliceEqual([]interface{}{m1, m1}, []interface{}{m1, m}) {
		t.Fatalf("Should be false")
	}

	if cc.sliceEqual([]interface{}{m}, []interface{}{m1}) {
		t.Fatalf("Should be false")
	}

	if cc.sliceEqual([]interface{}{m, m}, []interface{}{m1, m}) {
		t.Fatalf("Should be false")
	}

	// interfaces

	if !cc.sliceEqual([]interface{}{"one", "two"}, []interface{}{"one", "two"}) {
		t.Fatalf("Should be true")
	}

	if !cc.sliceEqual([]interface{}{"one", "two"}, []interface{}{"two", "one"}) {
		t.Fatalf("Should be true")
	}

	if cc.sliceEqual([]interface{}{"one", "two"}, []interface{}{"one", "one"}) {
		t.Fatalf("Should be false")
	}

	if cc.sliceEqual([]interface{}{"one", "two"}, []interface{}{"one"}) {
		t.Fatalf("Should be false")
	}

	if cc.sliceEqual([]interface{}{"two"}, []interface{}{"two", "one"}) {
		t.Fatalf("Should be false")
	}
}
