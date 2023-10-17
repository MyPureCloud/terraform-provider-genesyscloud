package lists

import (
	"testing"
)

type AreEquivalentTestCase struct {
	arrayA     []string
	arrayB     []string
	equivalent bool
}

func TestAreEquivalent(t *testing.T) {
	testCases := []AreEquivalentTestCase{
		{
			arrayA:     []string{},
			arrayB:     []string{},
			equivalent: true,
		},
		{
			arrayA:     []string{"foo", "bar"},
			arrayB:     []string{"bar", "foo"},
			equivalent: true,
		},
		{
			arrayA:     []string{"y", "x", "foo", "bar"},
			arrayB:     []string{"x", "bar", "foo", "y"},
			equivalent: true,
		},
		{
			arrayA:     []string{"x", "x", "x"},
			arrayB:     []string{"x", "x"},
			equivalent: false,
		},
		{
			arrayA:     []string{"x", "x"},
			arrayB:     []string{"x", "y"},
			equivalent: false,
		},
	}

	for _, tc := range testCases {
		arrayACopy := make([]string, len(tc.arrayA))
		arrayBCopy := make([]string, len(tc.arrayB))
		copy(arrayACopy, tc.arrayA)
		copy(arrayBCopy, tc.arrayB)

		result := AreEquivalent(tc.arrayA, tc.arrayB)
		if result != tc.equivalent {
			t.Errorf("Got %v for lists %v and %v. Should have got %v.", result, tc.arrayA, tc.arrayB, tc.equivalent)
		}

		// Ensure the sort function hasn't manipulated the arrays that were passed into AreEquivalent()
		if len(tc.arrayA) != len(arrayACopy) {
			t.Errorf("arrayA has changed after going through the function. Should be %v, got %v", arrayACopy, tc.arrayA)
		}
		if len(tc.arrayB) != len(arrayBCopy) {
			t.Errorf("arrayB has changed after going through the function. Should be %v, got %v", arrayBCopy, tc.arrayB)
		}
		for k, v := range tc.arrayA {
			if v != arrayACopy[k] {
				t.Errorf("arrayA has changed after going through the function. Should be %v, got %v", arrayACopy, tc.arrayA)
			}
		}
		for k, v := range tc.arrayB {
			if v != arrayBCopy[k] {
				t.Errorf("arrayB has changed after going through the function. Should be %v, got %v", arrayBCopy, tc.arrayB)
			}
		}
	}
}
