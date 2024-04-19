package util

import "testing"

func TestUnitGetQueryParamValueFromUri(t *testing.T) {
	type testCase struct {
		url           string
		param         string
		expectedValue string
	}

	testCases := &[]testCase{
		{
			url:           "api/v2/example?after=12345",
			param:         "after",
			expectedValue: "12345",
		},
		{
			url:           "api/v2/example?foo=bar&after=abcd",
			param:         "after",
			expectedValue: "abcd",
		},
		{
			url:           "api/v2/example?foo=bar&after=abcd",
			param:         "foo",
			expectedValue: "bar",
		},
		{
			url:           "api/v2/example?foo=bar&after=abcd",
			param:         "nonexistent",
			expectedValue: "",
		},
		{
			url:           "api/v2/example",
			param:         "after",
			expectedValue: "",
		},
	}

	for _, testCase := range *testCases {
		val, err := GetQueryParamValueFromUri(testCase.url, testCase.param)
		if err != nil {
			t.Errorf("expected error to be nil, got: %v", err)
		}
		if val != testCase.expectedValue {
			t.Errorf("expected value: %s, actual value: %s", testCase.expectedValue, val)
		}
	}
}
