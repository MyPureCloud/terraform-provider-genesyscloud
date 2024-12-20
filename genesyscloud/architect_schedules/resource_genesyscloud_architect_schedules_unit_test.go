package architect_schedules

import (
	"terraform-provider-genesyscloud/genesyscloud/util/lists"
	"testing"
	"time"
)

func TestUnitGetByDaysFromRRule(t *testing.T) {
	type inputAndResults struct {
		input          string
		expectedOutput []string
	}

	testCases := []inputAndResults{
		{
			input:          "",
			expectedOutput: []string{},
		},
		{
			input:          "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO",
			expectedOutput: []string{time.Monday.String()},
		},
		{
			input:          "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=12;BYDAY=TH,FR",
			expectedOutput: []string{time.Thursday.String(), time.Friday.String()},
		},
		{
			input:          "FREQ=YEARLY;COUNT=3;INTERVAL=1;BYMONTH=11;BYMONTHDAY=1",
			expectedOutput: []string{},
		},
		{
			input: "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,TU,WE,TH,FR,SA,SU",
			expectedOutput: []string{
				time.Monday.String(),
				time.Tuesday.String(),
				time.Wednesday.String(),
				time.Thursday.String(),
				time.Friday.String(),
				time.Saturday.String(),
				time.Sunday.String(),
			},
		},
		{
			input:          "FREQ=MONTHLY;BYDAY=MO,FR;INTERVAL=1;BYMONTHDAY=12;",
			expectedOutput: []string{time.Monday.String(), time.Friday.String()},
		},
	}

	for _, testCase := range testCases {
		result := getDaysFromRRule(testCase.input)
		if len(result) != len(testCase.expectedOutput) {
			t.Errorf("getDaysFromRRule(%q) returned %d results, expected %d", testCase.input, len(result), len(testCase.expectedOutput))
		}
		if !lists.AreEquivalent(result, testCase.expectedOutput) {
			t.Errorf("lists are not equivalent, expected %v, got %v", testCase.expectedOutput, result)
		}
	}
}

func TestUnitVerifyStartDateConformsToRRule(t *testing.T) {
	const scheduleName = "example schedule"

	type inputAndResults struct {
		dateTime    string
		rrule       string
		expectError bool
	}

	testCases := []inputAndResults{
		{
			dateTime:    "2018-05-07T09:00:00.000000", // Monday
			rrule:       "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO",
			expectError: false,
		},
		{
			dateTime:    "2018-11-16T09:43:00.000000", // Friday
			rrule:       "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,TU,WE,TH,FR,SA,SU",
			expectError: false,
		},
		{
			dateTime:    "2016-03-31T10:00:00.000000", // Thursday
			rrule:       "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO",
			expectError: true,
		},
		{
			dateTime:    "2020-11-16T07:31:00.000000", // Monday
			rrule:       "FREQ=YEARLY;COUNT=3;INTERVAL=1;BYMONTH=11;BYMONTHDAY=1",
			expectError: false,
		},
	}

	for i, testCase := range testCases {
		dt, err := time.Parse(timeFormat, testCase.dateTime)
		if err != nil {
			t.Errorf("failed to parse date time %q: %s", testCase.dateTime, err)
		}
		err = verifyStartDateConformsToRRule(dt, testCase.rrule, scheduleName)
		if err != nil && !testCase.expectError {
			t.Errorf("verifyStartDateConformsToRRule returned unexpected error on test case #%d: %s", i+1, err.Error())
		}
		if err == nil && testCase.expectError {
			t.Errorf("Expected verifyStartDateConformsToRRule to return an error for test case #%d, got nil", i+1)
		}
	}
}
