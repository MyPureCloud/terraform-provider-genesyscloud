package util

import (
	"testing"
)

type isValidTestCase struct {
	number        string
	expectedValue bool
	expectError   bool
}

var isValidTestCases = &[]isValidTestCase{
	// US phone numbers
	{
		number:        "+1 (919) 333-1234",
		expectedValue: true,
		expectError:   false,
	},
	{
		number:        "1-919-333-1234",
		expectedValue: false,
		expectError:   true,
	},
	// By default, add US international code if one is not given
	{
		number:        "(919) 333-1234",
		expectedValue: false,
		expectError:   true,
	},
	{
		number:        "919-333-1234",
		expectedValue: false,
		expectError:   true,
	},
	// UK phone numbers
	{
		number:        "+44 20 7123 1234",
		expectedValue: true,
		expectError:   false,
	},
	{
		number:        "+44 (020) 7123 1234",
		expectedValue: true,
		expectError:   false,
	},
	// German phone numbers
	{
		number:        "+49 (89) 1234-5678",
		expectedValue: true,
		expectError:   false,
	},
	{
		number:        "+49 089 1234-5678",
		expectedValue: true,
		expectError:   false,
	},
	{
		number:        "+49 89 1234-5678",
		expectedValue: true,
		expectError:   false,
	},
	// Indian phone numbers
	{
		number:        "+91 (22) 1234-5678",
		expectedValue: true,
		expectError:   false,
	},
	{
		number:        "+91 022 1234-5678",
		expectedValue: true,
		expectError:   false,
	},
	{
		number:        "+91 22 1234-5678",
		expectedValue: true,
		expectError:   false,
	},
	// Australian phone numbers
	{
		number:        "+61 3 1234 5678",
		expectedValue: true,
		expectError:   false,
	},
	{
		number:        "+61 03 1234 5678",
		expectedValue: true,
		expectError:   false,
	},
	{
		number:        "+61 3 1234 5678",
		expectedValue: true,
		expectError:   false,
	},
	// Edge cases
	{
		number:        "123-456-7890",
		expectedValue: false, // Assuming US number without country code
		expectError:   true,
	},
	{
		number:        "+1 123",
		expectedValue: false, // Has plus prefix but still invalid
		expectError:   false,
	},
	{
		number:        "12345",
		expectedValue: false,
		expectError:   true,
	},
	{
		number:        "+81 123",
		expectedValue: false, // Has plus prefix but still invalid
		expectError:   false,
	},
	{
		number:        "12345",
		expectedValue: false, // Has plus prefix but still invalid
		expectError:   true,
	},
	{
		number:        "+00",
		expectedValue: false,
		expectError:   true,
	},
	{
		number:        "+1",
		expectedValue: false,
		expectError:   true,
	},
	{
		number:        "0",
		expectedValue: true,
		expectError:   true,
	},
}

func testIsValidE164Number(t *testing.T, utilE164 UtilE164Service, testCases *[]isValidTestCase) {
	for _, testCase := range *testCases {
		val, err := utilE164.IsValidE164Number(testCase.number)
		if testCase.expectError {
			if err == nil {
				t.Errorf("expected error for %v, got nil", testCase.number)
			}
			continue
		} else {
			if err != nil {
				t.Errorf("expected no error for %v, got: %v", testCase.number, err)
			}
		}
		if err != nil {
			t.Errorf("expected error to be nil for %v, got: %v", testCase.number, err)
		}
		if val != testCase.expectedValue {
			t.Errorf("number: %s, expected value: %v, actual value: %v", testCase.number, testCase.expectedValue, val)
		}
	}
}

func TestUnitIsValidE164Number(t *testing.T) {

	countryCodeUS := func() string {
		return "US"
	}
	var utilE164 = *NewUtilE164Service()
	utilE164.GetDefaultCountryCodeFunc = countryCodeUS

	testIsValidE164Number(t, utilE164, isValidTestCases)
}

// Same tests as above, but use a different default if the country code is different
func TestUnitFormatsAsValidE164NumberWithAltCountryCode(t *testing.T) {

	countryCodeJP := func() string {
		return "JP"
	}
	var utilE164 = *NewUtilE164Service()
	utilE164.GetDefaultCountryCodeFunc = countryCodeJP

	testIsValidE164Number(t, utilE164, isValidTestCases)

}

func TestUnitFormatsAsValidE164NumberError(t *testing.T) {
	type testCase struct {
		number string
	}

	countryCodeUS := func() string {
		return "US"
	}
	var utilE164 = *NewUtilE164Service()
	utilE164.GetDefaultCountryCodeFunc = countryCodeUS

	testCases := &[]testCase{
		// Random characters
		{
			number: "+1@@##3i-[0340-231234",
		},
		{
			number: "adahsioidaodah",
		},
		// Invalid international codes
		{
			number: "+4239 (372) (332 20 7123) 23223 4334 232323 1234",
		},
		{
			number: "+59",
		},
		// Edge cases
		{
			number: "102082308230320927092371982317932179821938143986439187639846398634963493496349634983419834",
		},
		{
			number: "0",
		},
	}

	for _, testCase := range *testCases {
		val, err := utilE164.IsValidE164Number(testCase.number)
		// We expect the error to be present for these cases
		if err == nil {
			t.Errorf("expected error to be to not be nil for number: %v, value: %v", testCase.number, val)
		}
	}
}

func TestUnitFormatsAsCalculatedE164Number(t *testing.T) {
	type testCase struct {
		number        string
		expectedValue string
	}

	countryCodeUS := func() string {
		return "US"
	}
	var utilE164 = *NewUtilE164Service()
	utilE164.GetDefaultCountryCodeFunc = countryCodeUS

	testCases := &[]testCase{
		// US phone numbers
		{
			number:        "+1 (919) 333-1234",
			expectedValue: "+19193331234",
		},
		{
			number:        "1-919-333-1234",
			expectedValue: "+19193331234",
		},
		// By default, add US international code if one is not given
		{
			number:        "(919) 333-1234",
			expectedValue: "+19193331234",
		},
		{
			number:        "919-333-1234",
			expectedValue: "+19193331234",
		},
		// UK phone numbers
		{
			number:        "+44 20 7123 1234",
			expectedValue: "+442071231234",
		},
		{
			number:        "+44 (020) 7123 1234",
			expectedValue: "+442071231234",
		},
		// German phone numbers
		{
			number:        "+49 (89) 1234-5678",
			expectedValue: "+498912345678",
		},
		{
			number:        "+49 089 1234-5678",
			expectedValue: "+498912345678",
		},
		{
			number:        "+49 89 1234-5678",
			expectedValue: "+498912345678",
		},
		// Indian phone numbers
		{
			number:        "+91 (22) 1234-5678",
			expectedValue: "+912212345678",
		},
		{
			number:        "+91 022 1234-5678",
			expectedValue: "+912212345678",
		},
		{
			number:        "+91 22 1234-5678",
			expectedValue: "+912212345678",
		},
		// Australian phone numbers
		{
			number:        "+61 3 1234 5678",
			expectedValue: "+61312345678",
		},
		{
			number:        "+61 03 1234 5678",
			expectedValue: "+61312345678",
		},
		{
			number:        "+61 3 1234 5678",
			expectedValue: "+61312345678",
		},
		// Edge cases
		{
			number:        "123-456-7890",
			expectedValue: "+11234567890", // Assuming US number without country code
		},
		{
			number:        "+1 123",
			expectedValue: "+1123", // Invalid but still formatted
		},
		{
			number:        "12345",
			expectedValue: "+112345", // Invalid but still formatted
		},
		{
			number:        "+1",
			expectedValue: "+1",
		},
		{
			number:        "1",
			expectedValue: "+1",
		},
		{
			number:        "+81",
			expectedValue: "+81",
		},
		{
			number:        "81",
			expectedValue: "+181",
		},
		{
			number:        "0",
			expectedValue: "+0",
		},
		{
			number:        "0",
			expectedValue: "+0",
		},
	}
	for _, testCase := range *testCases {
		val := utilE164.FormatAsCalculatedE164Number(testCase.number)
		if val != testCase.expectedValue {
			t.Errorf("number: %s, expected value: %s, actual value: %s", testCase.number, testCase.expectedValue, val)
		}
	}
}

func TestUnitFormatsAsCalculatedE164NumberWithAltCountryCode(t *testing.T) {
	type testCase struct {
		number        string
		expectedValue string
	}

	countryCodeJP := func() string {
		return "JP"
	}
	var utilE164 = *NewUtilE164Service()
	utilE164.GetDefaultCountryCodeFunc = countryCodeJP

	testCases := &[]testCase{
		// US phone numbers
		{
			number:        "+1 (919) 333-1234",
			expectedValue: "+19193331234",
		},
		{
			number:        "1-919-333-1234",
			expectedValue: "+8119193331234",
		},
		// By default, add US international code if one is not given
		{
			number:        "(919) 333-1234",
			expectedValue: "+819193331234",
		},
		{
			number:        "919-333-1234",
			expectedValue: "+819193331234",
		},
		// UK phone numbers
		{
			number:        "+44 20 7123 1234",
			expectedValue: "+442071231234",
		},
		{
			number:        "+44 (020) 7123 1234",
			expectedValue: "+442071231234",
		},
		// German phone numbers
		{
			number:        "+49 (89) 1234-5678",
			expectedValue: "+498912345678",
		},
		{
			number:        "+49 089 1234-5678",
			expectedValue: "+498912345678",
		},
		{
			number:        "+49 89 1234-5678",
			expectedValue: "+498912345678",
		},
		// Indian phone numbers
		{
			number:        "+91 (22) 1234-5678",
			expectedValue: "+912212345678",
		},
		{
			number:        "+91 022 1234-5678",
			expectedValue: "+912212345678",
		},
		{
			number:        "+91 22 1234-5678",
			expectedValue: "+912212345678",
		},
		// Australian phone numbers
		{
			number:        "+61 3 1234 5678",
			expectedValue: "+61312345678",
		},
		{
			number:        "+61 03 1234 5678",
			expectedValue: "+61312345678",
		},
		{
			number:        "+61 3 1234 5678",
			expectedValue: "+61312345678",
		},
		// Edge cases
		{
			number:        "123-456-7890",
			expectedValue: "+811234567890", // Assuming US number without country code
		},
		{
			number:        "+81 123",
			expectedValue: "+81123", // Invalid but still formatted
		},
		{
			number:        "12345",
			expectedValue: "+8112345", // Invalid but still formatted
		},
		{
			number:        "+1",
			expectedValue: "+1",
		},
		{
			number:        "1",
			expectedValue: "+1",
		},
		{
			number:        "+81",
			expectedValue: "+81",
		},
		{
			number:        "81",
			expectedValue: "+8181",
		},
		{
			number:        "0",
			expectedValue: "+0",
		},
	}

	for _, testCase := range *testCases {
		val := utilE164.FormatAsCalculatedE164Number(testCase.number)
		if val != testCase.expectedValue {
			t.Errorf("number: %s, expected value: %s, actual value: %s", testCase.number, testCase.expectedValue, val)
		}
	}
}

func TestUnitFormatsAsCalculatedE164NumberError(t *testing.T) {
	type testCase struct {
		number        string
		expectedValue string
	}

	countryCodeUS := func() string {
		return "US"
	}
	var utilE164 = *NewUtilE164Service()
	utilE164.GetDefaultCountryCodeFunc = countryCodeUS

	testCases := &[]testCase{
		// Random characters
		{
			number:        "+1@@##3i-[0340-231234",
			expectedValue: "+1@@##3i-[0340-231234",
		},
		{
			number:        "adahsioidaodah",
			expectedValue: "+adahsioidaodah",
		},
		// Invalid international codes
		{
			number:        "+423 9 (372) (332 20 7123) 23223 4334 232323 1234",
			expectedValue: "+4230",
		},
		{
			number:        "+59",
			expectedValue: "+59",
		},
		// Edge cases
		{
			number:        "102082308230320927092371982317932179821938143986439187639846398634963493496349634983419834",
			expectedValue: "+10",
		},
		{
			number:        "0",
			expectedValue: "+0",
		},
	}

	for _, testCase := range *testCases {
		val := utilE164.FormatAsCalculatedE164Number(testCase.number)
		// We expect the value to be blank for invalid calculated values
		if val != testCase.expectedValue {
			t.Errorf("expected value to be %s for number: %s, value: %s", testCase.expectedValue, testCase.number, val)
		}
	}
}
